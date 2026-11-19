package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// responderJSON escribe una respuesta JSON estándar con el código HTTP proporcionado.
func responderJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// responderError escribe un error en el formato contractual { "error": { "code": "...", "message": "..." } }.
func responderError(w http.ResponseWriter, status int, code, message string) {
	responderJSON(w, status, ErrorRespuesta{
		Error: DetalleError{
			Code:    code,
			Message: message,
		},
	})
}

// extraerIPCliente obtiene la IP real de la solicitud, soportando cabeceras de proxy.
func extraerIPCliente(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		partes := strings.Split(xff, ",")
		if len(partes) > 0 {
			ip := strings.TrimSpace(partes[0])
			if ip != "" {
				return ip
			}
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	return contactoIP(r)
}

// obtenerParametroID extrae el identificador {id} de la URL de forma segura.
func obtenerParametroID(r *http.Request) string {
	if id := r.PathValue("id"); id != "" {
		return id
	}
	partes := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(partes) > 0 {
		return partes[len(partes)-1]
	}
	return ""
}

// loginHandler maneja POST /auth/login
// Reglas:
// - Límite de intentos: 5 fallidos → bloqueo temporal de 15 min por IP+email
// - Cookie HttpOnly con 8h de inactividad
// - Respuesta 200 con el objeto user
func loginHandler(db *sql.DB, limitador *authLimitador) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			responderError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
			return
		}

		if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
			responderError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "El contenido debe ser application/json")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<14)
		var in LoginEntrada
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&in); err != nil {
			responderError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido o mal formado")
			return
		}

		email := strings.ToLower(strings.TrimSpace(in.Email))
		password := in.Password
		ip := extraerIPCliente(r)

		if email == "" || password == "" {
			responderError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Debes proporcionar correo y contraseña")
			return
		}

		// Verificación de bloqueo por intentos fallidos
		if bloqueado, restante := limitador.estaBloqueado(ip, email); bloqueado {
			w.Header().Set("Retry-After", strconv.Itoa(int(restante.Seconds())))
			responderError(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "Demasiados intentos fallidos. Bloqueo temporal de 15 minutos.")
			return
		}

		usuario, err := authBuscarUsuarioPorEmail(r.Context(), db, email)
		if err != nil {
			log.Printf("auth: error al consultar usuario: %v", err)
			responderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error interno al procesar el inicio de sesión")
			return
		}

		// Si el usuario no existe o está inactivo
		if usuario == nil || !usuario.Activo {
			bloqueado, _ := limitador.registrarFallo(ip, email)
			if bloqueado {
				responderError(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "Demasiados intentos fallidos. Bloqueo temporal de 15 minutos.")
				return
			}
			responderError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Credenciales inválidas")
			return
		}

		valido, err := verificarPassword(password, usuario.PasswordHash)
		if err != nil || !valido {
			bloqueado, _ := limitador.registrarFallo(ip, email)
			if bloqueado {
				responderError(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "Demasiados intentos fallidos. Bloqueo temporal de 15 minutos.")
				return
			}
			responderError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Credenciales inválidas")
			return
		}

		// Credenciales válidas: reset de fallos
		limitador.registrarExito(ip, email)

		token := generarTokenSesion()
		ahora := time.Now()
		sesion := &Sesion{
			ID:           token,
			UsuarioID:    usuario.ID,
			IP:           ip,
			UserAgent:    r.UserAgent(),
			CreadoEn:     ahora,
			UltimoAcceso: ahora,
			ExpiraEn:     ahora.Add(duracionSesion),
		}

		if err := authCrearSesion(r.Context(), db, sesion); err != nil {
			log.Printf("auth: error al crear sesión: %v", err)
			responderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "No se pudo crear la sesión")
			return
		}

		fijarCookieSesion(w, r, token, int(duracionSesion.Seconds()))

		responderJSON(w, http.StatusOK, LoginRespuesta{
			User: usuario.ConvertirAPublico(false),
		})
	}
}

// logoutHandler maneja POST /auth/logout
func logoutHandler(db *sql.DB) http.HandlerFunc {
	return requerirAutenticado(db, func(w http.ResponseWriter, r *http.Request) {
		token := tokenSesionDesdeContexto(r.Context())
		if token != "" {
			_ = authEliminarSesion(r.Context(), db, token)
		}
		fijarCookieSesion(w, r, "", -1)
		w.WriteHeader(http.StatusNoContent)
	})
}

// meHandler maneja GET /auth/me
func meHandler(db *sql.DB) http.HandlerFunc {
	return requerirAutenticado(db, func(w http.ResponseWriter, r *http.Request) {
		usuario := usuarioDesdeContexto(r.Context())
		if usuario == nil {
			responderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "No autenticado")
			return
		}
		responderJSON(w, http.StatusOK, LoginRespuesta{
			User: usuario.ConvertirAPublico(false),
		})
	})
}

// listarUsuariosCore ejecuta la consulta paginada de usuarios.
func listarUsuariosCore(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		page, _ := strconv.Atoi(query.Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(query.Get("limit"))
		if limit < 1 || limit > 100 {
			limit = 10
		}

		items, total, err := authListarUsuarios(r.Context(), db, page, limit)
		if err != nil {
			log.Printf("auth: error al listar usuarios: %v", err)
			responderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al listar usuarios")
			return
		}

		responderJSON(w, http.StatusOK, ListaPaginada[UsuarioPublico]{
			Items: items,
			Total: total,
			Page:  page,
		})
	}
}

// listarUsuariosHandler maneja GET /usuarios (solo admin)
func listarUsuariosHandler(db *sql.DB) http.HandlerFunc {
	return requerirRol(db, []string{RolAdmin}, listarUsuariosCore(db))
}

// crearUsuarioCore valida y guarda un nuevo usuario con contraseña hasheada en Argon2.
func crearUsuarioCore(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
			responderError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "El contenido debe ser application/json")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<14)
		var in UsuarioCrearEntrada
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			responderError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido o mal formado")
			return
		}

		in.Nombre = strings.TrimSpace(in.Nombre)
		in.Email = strings.ToLower(strings.TrimSpace(in.Email))
		in.Rol = strings.ToLower(strings.TrimSpace(in.Rol))

		if in.Nombre == "" || utf8.RuneCountInString(in.Nombre) < 2 || utf8.RuneCountInString(in.Nombre) > 150 {
			responderError(w, http.StatusBadRequest, "VALIDATION_ERROR", "El nombre es obligatorio y debe tener entre 2 y 150 caracteres")
			return
		}

		if in.Email == "" || utf8.RuneCountInString(in.Email) > 190 {
			responderError(w, http.StatusBadRequest, "VALIDATION_ERROR", "El correo es obligatorio y no puede exceder 190 caracteres")
			return
		}
		if _, err := mail.ParseAddress(in.Email); err != nil {
			responderError(w, http.StatusBadRequest, "VALIDATION_ERROR", "El formato del correo electrónico no es válido")
			return
		}

		if in.Rol != RolAdmin && in.Rol != RolEditor && in.Rol != RolEmpleado {
			responderError(w, http.StatusBadRequest, "VALIDATION_ERROR", "El rol debe ser admin, editor o empleado")
			return
		}

		if len(in.PasswordTemporal) < 8 {
			responderError(w, http.StatusBadRequest, "VALIDATION_ERROR", "La contraseña temporal debe tener al menos 8 caracteres")
			return
		}

		// Validar si el correo ya existe
		existente, err := authBuscarUsuarioPorEmail(r.Context(), db, in.Email)
		if err != nil {
			log.Printf("auth: error al comprobar existencia de email: %v", err)
			responderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al validar correo")
			return
		}
		if existente != nil {
			responderError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "El correo ya está registrado por otro usuario")
			return
		}

		hash, err := hashPassword(in.PasswordTemporal)
		if err != nil {
			log.Printf("auth: error al hashear password con argon2: %v", err)
			responderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al procesar la contraseña")
			return
		}

		nuevoUsuario := &Usuario{
			ID:           generarUUID(),
			Nombre:       in.Nombre,
			Email:        in.Email,
			PasswordHash: hash,
			Rol:          in.Rol,
			Activo:       true,
			CreadoEn:     time.Now(),
		}

		if err := authCrearUsuario(r.Context(), db, nuevoUsuario); err != nil {
			log.Printf("auth: error al guardar usuario: %v", err)
			responderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al crear el usuario")
			return
		}

		responderJSON(w, http.StatusCreated, LoginRespuesta{
			User: nuevoUsuario.ConvertirAPublico(true),
		})
	}
}

// crearUsuarioHandler maneja POST /usuarios (solo admin)
// Regla: la contraseña se hashea con argon2 antes de guardar; nunca se devuelve en ninguna respuesta
func crearUsuarioHandler(db *sql.DB) http.HandlerFunc {
	return requerirRol(db, []string{RolAdmin}, crearUsuarioCore(db))
}

// actualizarUsuarioCore ejecuta la lógica de actualización tras la autorización.
func actualizarUsuarioCore(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := obtenerParametroID(r)
		if id == "" {
			responderError(w, http.StatusBadRequest, "INVALID_ID", "ID de usuario requerido")
			return
		}

		adminActual := usuarioDesdeContexto(r.Context())
		if adminActual == nil {
			responderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "No autenticado")
			return
		}

		// Reglas de auto-protección del administrador
		if id == adminActual.ID {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<14)
			var in UsuarioActualizarEntrada
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil && !errors.Is(err, io.EOF) {
				responderError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido o mal formado")
				return
			}
			if in.Activo != nil && !*in.Activo {
				responderError(w, http.StatusBadRequest, "CANNOT_DEACTIVATE_SELF", "No puedes desactivar tu propia cuenta de administrador")
				return
			}
			if in.Rol != nil && *in.Rol != RolAdmin {
				responderError(w, http.StatusBadRequest, "CANNOT_DEMOTE_SELF", "No puedes revocar tu propio rol de administrador")
				return
			}
		}

		// Comprobar existencia
		existente, err := authBuscarUsuarioPorID(r.Context(), db, id)
		if err != nil {
			log.Printf("auth: error al buscar usuario para actualizar: %v", err)
			responderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al buscar el usuario")
			return
		}
		if existente == nil {
			responderError(w, http.StatusNotFound, "USER_NOT_FOUND", "Usuario no encontrado")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<14)
		var in UsuarioActualizarEntrada
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil && !errors.Is(err, io.EOF) {
			responderError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido o mal formado")
			return
		}

		if in.Nombre != nil {
			nombreLimpio := strings.TrimSpace(*in.Nombre)
			if utf8.RuneCountInString(nombreLimpio) < 2 || utf8.RuneCountInString(nombreLimpio) > 150 {
				responderError(w, http.StatusBadRequest, "VALIDATION_ERROR", "El nombre debe tener entre 2 y 150 caracteres")
				return
			}
			in.Nombre = &nombreLimpio
		}

		if in.Rol != nil {
			rolLimpio := strings.ToLower(strings.TrimSpace(*in.Rol))
			if rolLimpio != RolAdmin && rolLimpio != RolEditor && rolLimpio != RolEmpleado {
				responderError(w, http.StatusBadRequest, "VALIDATION_ERROR", "El rol debe ser admin, editor o empleado")
				return
			}
			in.Rol = &rolLimpio
		}

		// Si se desactiva una cuenta, revocar todas sus sesiones activas de inmediato
		if in.Activo != nil && !*in.Activo {
			_ = authEliminarSesionesDeUsuario(r.Context(), db, id)
		}

		actualizado, err := authActualizarUsuario(r.Context(), db, id, in.Nombre, in.Rol, in.Activo)
		if err != nil {
			log.Printf("auth: error al actualizar usuario: %v", err)
			responderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al actualizar usuario")
			return
		}

		responderJSON(w, http.StatusOK, LoginRespuesta{
			User: actualizado.ConvertirAPublico(true),
		})
	}
}

// actualizarUsuarioHandler maneja PATCH /usuarios/:id (solo admin)
func actualizarUsuarioHandler(db *sql.DB) http.HandlerFunc {
	return requerirRol(db, []string{RolAdmin}, actualizarUsuarioCore(db))
}

// eliminarUsuarioCore ejecuta la eliminación validando que no sea auto-eliminación.
func eliminarUsuarioCore(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := obtenerParametroID(r)
		if id == "" {
			responderError(w, http.StatusBadRequest, "INVALID_ID", "ID de usuario requerido")
			return
		}

		adminActual := usuarioDesdeContexto(r.Context())
		if adminActual == nil {
			responderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "No autenticado")
			return
		}

		// Regla contractual: no permite que un admin se auto-elimine
		if id == adminActual.ID {
			responderError(w, http.StatusBadRequest, "CANNOT_DELETE_SELF", "No se permite que un administrador se auto-elimine")
			return
		}

		existente, err := authBuscarUsuarioPorID(r.Context(), db, id)
		if err != nil {
			log.Printf("auth: error al verificar usuario para eliminar: %v", err)
			responderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al verificar usuario")
			return
		}
		if existente == nil {
			responderError(w, http.StatusNotFound, "USER_NOT_FOUND", "Usuario no encontrado")
			return
		}

		_ = authEliminarSesionesDeUsuario(r.Context(), db, id)
		if err := authEliminarUsuario(r.Context(), db, id); err != nil {
			log.Printf("auth: error al eliminar usuario: %v", err)
			responderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al eliminar usuario")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// eliminarUsuarioHandler maneja DELETE /usuarios/:id (solo admin)
// Regla: no permite que un admin se auto-elimine
func eliminarUsuarioHandler(db *sql.DB) http.HandlerFunc {
	return requerirRol(db, []string{RolAdmin}, eliminarUsuarioCore(db))
}

// registrarPanel registra todas las rutas de autenticación y gestión de usuarios (RF-04).
func registrarPanel(mux *http.ServeMux, db *sql.DB) {
	// Auto-asegurar tablas en la base de datos
	go func() {
		if err := authCrearTablasSiNoExisten(context.Background(), db); err != nil {
			log.Printf("auth: advertencia al verificar tablas: %v", err)
		}
	}()

	limitador := nuevoAuthLimitador()

	// 1. Autenticación
	hLogin := loginHandler(db, limitador)
	hLogout := logoutHandler(db)
	hMe := meHandler(db)

	// Rutas bajo /api/v1 (Base URL oficial) y rutas directas /auth/... para máxima compatibilidad
	mux.HandleFunc("POST /api/v1/auth/login", hLogin)
	mux.HandleFunc("POST /auth/login", hLogin)

	mux.HandleFunc("POST /api/v1/auth/logout", hLogout)
	mux.HandleFunc("POST /auth/logout", hLogout)

	mux.HandleFunc("GET /api/v1/auth/me", hMe)
	mux.HandleFunc("GET /auth/me", hMe)

	// 2. Gestión de usuarios (solo admin)
	hListar := listarUsuariosHandler(db)
	hCrear := crearUsuarioHandler(db)
	hActualizar := actualizarUsuarioHandler(db)
	hEliminar := eliminarUsuarioHandler(db)

	mux.HandleFunc("GET /api/v1/usuarios", hListar)
	mux.HandleFunc("GET /usuarios", hListar)

	mux.HandleFunc("POST /api/v1/usuarios", hCrear)
	mux.HandleFunc("POST /usuarios", hCrear)

	mux.HandleFunc("PATCH /api/v1/usuarios/{id}", hActualizar)
	mux.HandleFunc("PATCH /usuarios/{id}", hActualizar)

	mux.HandleFunc("DELETE /api/v1/usuarios/{id}", hEliminar)
	mux.HandleFunc("DELETE /usuarios/{id}", hEliminar)
}
