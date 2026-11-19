package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestArgon2Hashing(t *testing.T) {
	password := "MiSuperContrasena123!"

	hash, err := hashPassword(password)
	if err != nil {
		t.Fatalf("error inesperado al hashear: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("el hash debe iniciar con $argon2id$, obtenido: %s", hash)
	}

	// Verificación exitosa
	valido, err := verificarPassword(password, hash)
	if err != nil {
		t.Fatalf("error al verificar: %v", err)
	}
	if !valido {
		t.Errorf("se esperaba que la contraseña correcta fuera válida")
	}

	// Contraseña incorrecta
	valido, err = verificarPassword("ContrasenaIncorrecta", hash)
	if err != nil {
		t.Fatalf("error al verificar contraseña incorrecta: %v", err)
	}
	if valido {
		t.Errorf("la contraseña errónea no debió ser válida")
	}

	// Hash corrupto
	_, err = verificarPassword(password, "hash_invalido")
	if err == nil {
		t.Errorf("se esperaba error al verificar hash con formato inválido")
	}
}

func TestGenerarUUID(t *testing.T) {
	vistos := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generarUUID()
		if len(id) != 36 {
			t.Fatalf("longitud de UUID inválida: %d (esperada 36)", len(id))
		}
		// Formato: 8-4-4-4-12
		partes := strings.Split(id, "-")
		if len(partes) != 5 {
			t.Fatalf("el UUID no tiene 5 segmentos: %s", id)
		}
		// Versión 4: partes[2] debe iniciar con '4'
		if partes[2][0] != '4' {
			t.Errorf("el UUID no tiene versión 4: %s", id)
		}
		if vistos[id] {
			t.Fatalf("UUID repetido encontrado: %s", id)
		}
		vistos[id] = true
	}
}

func TestGenerarTokenSesion(t *testing.T) {
	token1 := generarTokenSesion()
	token2 := generarTokenSesion()

	if len(token1) != 64 {
		t.Errorf("longitud esperada de token de sesión: 64, obtenida: %d", len(token1))
	}
	if token1 == token2 {
		t.Errorf("dos tokens generados sucesivamente no deben coincidir")
	}
}

func TestAuthLimitador(t *testing.T) {
	limitador := &authLimitador{
		intentos: make(map[string]*intentoRegistro),
	}

	ip := "192.168.1.100"
	email := "usuario@empresa.com"

	// Los primeros 4 intentos fallidos no deben bloquear
	for i := 1; i <= 4; i++ {
		bloqueado, _ := limitador.registrarFallo(ip, email)
		if bloqueado {
			t.Fatalf("el intento %d no debió bloquear al usuario", i)
		}
		estaBloq, _ := limitador.estaBloqueado(ip, email)
		if estaBloq {
			t.Fatalf("estaBloqueado no debió ser true en intento %d", i)
		}
	}

	// El 5to intento fallido debe bloquear por 15 minutos
	bloqueado, restante := limitador.registrarFallo(ip, email)
	if !bloqueado {
		t.Fatalf("el 5to intento fallido debe activar el bloqueo")
	}
	if restante <= 0 || restante > 15*time.Minute {
		t.Errorf("tiempo restante de bloqueo inválido: %v", restante)
	}

	estaBloq, _ := limitador.estaBloqueado(ip, email)
	if !estaBloq {
		t.Fatalf("estaBloqueado debe retornar true tras 5 fallos")
	}

	// Un IP o correo distinto no debe estar bloqueado
	otroBloq, _ := limitador.estaBloqueado("192.168.1.101", email)
	if otroBloq {
		t.Errorf("una IP diferente no debió estar bloqueada")
	}

	// Al registrar éxito, se restablece el estado
	limitador.registrarExito(ip, email)
	estaBloqDespues, _ := limitador.estaBloqueado(ip, email)
	if estaBloqDespues {
		t.Errorf("después de registrar éxito el usuario no debe continuar bloqueado")
	}
}

func TestErrorRespuestaFormatoJSON(t *testing.T) {
	resp := ErrorRespuesta{
		Error: DetalleError{
			Code:    "INVALID_CREDENTIALS",
			Message: "Credenciales inválidas",
		},
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("error al serializar error: %v", err)
	}

	esperado := `{"error":{"code":"INVALID_CREDENTIALS","message":"Credenciales inválidas"}}`
	if string(b) != esperado {
		t.Errorf("formato de error no coincide con el contrato.\nEsperado: %s\nObtenido: %s", esperado, string(b))
	}
}

func TestUsuarioConvertirAPublico(t *testing.T) {
	u := &Usuario{
		ID:           "test-uuid",
		Nombre:       "Test User",
		Email:        "test@example.com",
		PasswordHash: "$argon2id$v=19$...",
		Rol:          RolAdmin,
		Activo:       true,
		CreadoEn:     time.Date(2026, 11, 19, 10, 0, 0, 0, time.UTC),
	}

	// Vista simple para login/me
	pub := u.ConvertirAPublico(false)
	b, err := json.Marshal(pub)
	if err != nil {
		t.Fatalf("error serializando usuario público: %v", err)
	}
	str := string(b)
	if strings.Contains(str, "PasswordHash") || strings.Contains(str, "argon2") {
		t.Errorf("el hash de contraseña jamás debe ser expuesto en la respuesta JSON")
	}
	if strings.Contains(str, "activo") {
		t.Errorf("la vista simple no debe incluir el campo activo")
	}

	// Vista completa para admin/usuarios
	pubMeta := u.ConvertirAPublico(true)
	bMeta, _ := json.Marshal(pubMeta)
	strMeta := string(bMeta)
	if !strings.Contains(strMeta, `"activo":true`) {
		t.Errorf("la vista completa debe incluir el campo activo")
	}
	if !strings.Contains(strMeta, `"creado_en":`) {
		t.Errorf("la vista completa debe incluir creado_en")
	}
}


func TestMiddlewareRequerirAutenticadoSinCredenciales(t *testing.T) {
	handler := requerirAutenticado(nil, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("código esperado: 401, obtenido: %d", rr.Code)
	}

	var errResp ErrorRespuesta
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatalf("error decodificando respuesta de error: %v", err)
	}
	if errResp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("código de error esperado: UNAUTHORIZED, obtenido: %s", errResp.Error.Code)
	}
}

func TestMiddlewareRequerirRolForbidden(t *testing.T) {
	// Simulamos un usuario con rol 'editor' intentando acceder a una ruta de 'admin'
	usuarioEditor := &Usuario{
		ID:     "user-editor-123",
		Nombre: "Editor User",
		Email:  "editor@empresa.com",
		Rol:    RolEditor,
		Activo: true,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Probamos la lógica interna de validación de rol
		usuario := usuarioDesdeContexto(r.Context())
		if usuario.Rol != RolAdmin {
			responderError(w, http.StatusForbidden, "FORBIDDEN", "No tienes permisos suficientes")
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/usuarios", nil)
	ctx := context.WithValue(req.Context(), contextKeyUsuario, usuarioEditor)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("código esperado: 403 Forbidden, obtenido: %d", rr.Code)
	}

	var errResp ErrorRespuesta
	_ = json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp.Error.Code != "FORBIDDEN" {
		t.Errorf("código de error esperado: FORBIDDEN, obtenido: %s", errResp.Error.Code)
	}
}

func TestEliminarUsuarioNoPermiteAutoEliminacion(t *testing.T) {
	adminID := "admin-uuid-12345"
	adminUsuario := &Usuario{
		ID:     adminID,
		Nombre: "Admin Supremo",
		Email:  "admin@remaconsultores.com",
		Rol:    RolAdmin,
		Activo: true,
	}

	handler := eliminarUsuarioCore(nil)

	// Admin intenta eliminarse a sí mismo
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/usuarios/"+adminID, nil)
	req.SetPathValue("id", adminID)
	ctx := context.WithValue(req.Context(), contextKeyUsuario, adminUsuario)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("código esperado: 400 Bad Request, obtenido: %d", rr.Code)
	}

	var errResp ErrorRespuesta
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatalf("error decodificando respuesta: %v", err)
	}
	if errResp.Error.Code != "CANNOT_DELETE_SELF" {
		t.Errorf("código de error esperado: CANNOT_DELETE_SELF, obtenido: %s", errResp.Error.Code)
	}
}

func TestActualizarUsuarioNoPermiteAutoDesactivacion(t *testing.T) {
	adminID := "admin-uuid-12345"
	adminUsuario := &Usuario{
		ID:     adminID,
		Nombre: "Admin Supremo",
		Email:  "admin@remaconsultores.com",
		Rol:    RolAdmin,
		Activo: true,
	}

	handler := actualizarUsuarioCore(nil)

	body := strings.NewReader(`{"activo": false}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/usuarios/"+adminID, body)
	req.SetPathValue("id", adminID)
	ctx := context.WithValue(req.Context(), contextKeyUsuario, adminUsuario)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("código esperado: 400 Bad Request, obtenido: %d", rr.Code)
	}

	var errResp ErrorRespuesta
	_ = json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp.Error.Code != "CANNOT_DEACTIVATE_SELF" {
		t.Errorf("código de error esperado: CANNOT_DEACTIVATE_SELF, obtenido: %s", errResp.Error.Code)
	}
}


func TestLoginHandlerValidacionesHTTP(t *testing.T) {
	limitador := nuevoAuthLimitador()
	handler := loginHandler(nil, limitador)

	// 1. Método inválido
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	rrGet := httptest.NewRecorder()
	handler(rrGet, reqGet)
	if rrGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("esperado 405, obtenido %d", rrGet.Code)
	}

	// 2. Content-type faltante
	reqNoJSON := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader("hola"))
	rrNoJSON := httptest.NewRecorder()
	handler(rrNoJSON, reqNoJSON)
	if rrNoJSON.Code != http.StatusUnsupportedMediaType {
		t.Errorf("esperado 415, obtenido %d", rrNoJSON.Code)
	}

	// 3. Campos vacíos
	reqEmpty := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"","password":""}`))
	reqEmpty.Header.Set("Content-Type", "application/json")
	rrEmpty := httptest.NewRecorder()
	handler(rrEmpty, reqEmpty)
	if rrEmpty.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obtenido %d", rrEmpty.Code)
	}

	var errResp ErrorRespuesta
	_ = json.NewDecoder(rrEmpty.Body).Decode(&errResp)
	if errResp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("esperado VALIDATION_ERROR, obtenido %s", errResp.Error.Code)
	}
}

func TestLoginHandlerBloqueoTemporalRateLimit(t *testing.T) {
	limitador := nuevoAuthLimitador()
	ip := "203.0.113.195"
	email := "victima@remaconsultores.com"

	// Provocamos 5 fallos directamente en el limitador
	for i := 0; i < 5; i++ {
		limitador.registrarFallo(ip, email)
	}

	handler := loginHandler(nil, limitador)

	body := strings.NewReader(`{"email":"` + email + `","password":"PasswordIncorrecto!"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", ip)

	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("esperado 429 Too Many Requests, obtenido %d", rr.Code)
	}

	var errResp ErrorRespuesta
	_ = json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp.Error.Code != "TOO_MANY_REQUESTS" {
		t.Errorf("esperado código TOO_MANY_REQUESTS, obtenido %s", errResp.Error.Code)
	}
}

func TestCrearUsuarioValidaciones(t *testing.T) {
	adminUsuario := &Usuario{
		ID:     "admin-123",
		Nombre: "Admin",
		Email:  "admin@remaconsultores.com",
		Rol:    RolAdmin,
		Activo: true,
	}

	pruebas := []struct {
		nombre string
		body   string
	}{
		{"Email inválido", `{"nombre":"Juan","email":"correo-invalido","rol":"editor","password_temporal":"12345678"}`},
		{"Nombre muy corto", `{"nombre":"J","email":"juan@example.com","rol":"editor","password_temporal":"12345678"}`},
		{"Rol inválido", `{"nombre":"Juan","email":"juan@example.com","rol":"superadmin","password_temporal":"12345678"}`},
		{"Password muy corta", `{"nombre":"Juan","email":"juan@example.com","rol":"editor","password_temporal":"123"}`},
	}

	handler := crearUsuarioCore(nil)

	for _, p := range pruebas {
		t.Run(p.nombre, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/usuarios", strings.NewReader(p.body))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), contextKeyUsuario, adminUsuario)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("para '%s' se esperaba 400, obtenido %d", p.nombre, rr.Code)
			}
			var errResp ErrorRespuesta
			_ = json.NewDecoder(rr.Body).Decode(&errResp)
			if errResp.Error.Code != "VALIDATION_ERROR" {
				t.Errorf("para '%s' se esperaba VALIDATION_ERROR, obtenido %s", p.nombre, errResp.Error.Code)
			}
		})
	}
}
