package auth

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"rema-backend/internal/config"
)

type contextKey string

const (
	contextKeyUsuario contextKey = "auth_usuario"
	contextKeyToken   contextKey = "auth_token"
	duracionSesion               = 8 * time.Hour
	nombreCookie                 = "rema_session"
)

// UsuarioDesdeContexto recupera el usuario autenticado almacenado por el middleware.
func UsuarioDesdeContexto(ctx context.Context) *Usuario {
	if val := ctx.Value(contextKeyUsuario); val != nil {
		if u, ok := val.(*Usuario); ok {
			return u
		}
	}
	return nil
}

var usuarioDesdeContexto = UsuarioDesdeContexto

// TokenSesionDesdeContexto recupera el token de sesión activo.
func TokenSesionDesdeContexto(ctx context.Context) string {
	if val := ctx.Value(contextKeyToken); val != nil {
		if tok, ok := val.(string); ok {
			return tok
		}
	}
	return ""
}

var tokenSesionDesdeContexto = TokenSesionDesdeContexto

// ExtraerTokenSesion busca el token en la cookie rema_session o en el header Authorization (Bearer <token>).
func ExtraerTokenSesion(r *http.Request) string {
	if c, err := r.Cookie(nombreCookie); err == nil && strings.TrimSpace(c.Value) != "" {
		return strings.TrimSpace(c.Value)
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}

	return ""
}

var extraerTokenSesion = ExtraerTokenSesion

// fijarCookieSesion escribe la cookie HTTP-only y segura para la sesión.
func fijarCookieSesion(w http.ResponseWriter, r *http.Request, token string, maxAge int) {
	cookie := &http.Cookie{
		Name:     nombreCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   esPeticionSegura(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	}
	if maxAge <= 0 {
		cookie.Expires = time.Unix(0, 0)
	} else {
		cookie.Expires = time.Now().Add(time.Duration(maxAge) * time.Second)
	}
	http.SetCookie(w, cookie)
}

func esPeticionSegura(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); strings.EqualFold(proto, "https") {
		return true
	}
	if config.Get("COOKIE_SECURE", "") == "true" {
		return true
	}
	return false
}

// RequerirAutenticado protege una ruta asegurando que el cliente tenga una sesión válida y activa.
func RequerirAutenticado(db *sql.DB, siguiente http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := ExtraerTokenSesion(r)
		if token == "" {
			responderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesión requerida")
			return
		}

		if db == nil {
			responderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesión inválida o expirada")
			return
		}

		sesion, usuario, err := authBuscarSesionYUsuario(r.Context(), db, token)
		if err != nil {
			responderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al verificar sesión")
			return
		}

		if sesion == nil || usuario == nil {
			// Limpiar cookie caducada
			fijarCookieSesion(w, r, "", -1)
			responderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesión inválida o expirada")
			return
		}

		// Sliding session: Si restan menos de 4 horas de las 8, renovar expiración
		if time.Until(sesion.ExpiraEn) < 4*time.Hour {
			_ = authRenovarSesion(r.Context(), db, token, duracionSesion)
			fijarCookieSesion(w, r, token, int(duracionSesion.Seconds()))
		}

		// Inyectar en el contexto de la petición
		ctx := context.WithValue(r.Context(), contextKeyUsuario, usuario)
		ctx = context.WithValue(ctx, contextKeyToken, token)

		siguiente(w, r.WithContext(ctx))
	}
}

var requerirAutenticado = RequerirAutenticado

// RequerirRol asegura que el usuario autenticado cuente con alguno de los roles permitidos.
func RequerirRol(db *sql.DB, rolesPermitidos []string, siguiente http.HandlerFunc) http.HandlerFunc {
	return RequerirAutenticado(db, func(w http.ResponseWriter, r *http.Request) {
		u := UsuarioDesdeContexto(r.Context())
		if u == nil {
			responderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesión requerida")
			return
		}

		permitido := false
		for _, rol := range rolesPermitidos {
			if u.Rol == rol {
				permitido = true
				break
			}
		}

		if !permitido {
			responderError(w, http.StatusForbidden, "FORBIDDEN", "No tienes permisos para realizar esta acción")
			return
		}

		siguiente(w, r)
	})
}

var requerirRol = RequerirRol
