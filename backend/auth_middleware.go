package main

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const (
	contextKeyUsuario contextKey = "auth_usuario"
	contextKeyToken   contextKey = "auth_token"
	duracionSesion               = 8 * time.Hour
	nombreCookie                 = "rema_session"
)

// usuarioDesdeContexto recupera el usuario autenticado almacenado por el middleware.
func usuarioDesdeContexto(ctx context.Context) *Usuario {
	if val := ctx.Value(contextKeyUsuario); val != nil {
		if u, ok := val.(*Usuario); ok {
			return u
		}
	}
	return nil
}

// tokenSesionDesdeContexto recupera el token de sesión almacenado por el middleware.
func tokenSesionDesdeContexto(ctx context.Context) string {
	if val := ctx.Value(contextKeyToken); val != nil {
		if t, ok := val.(string); ok {
			return t
		}
	}
	return ""
}

// extraerTokenSesion obtiene el token desde la cookie rema_session o desde el header Authorization Bearer.
func extraerTokenSesion(r *http.Request) string {
	if cookie, err := r.Cookie(nombreCookie); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}

	return ""
}

// fijarCookieSesion emite la cookie HttpOnly y Secure según contrato.
func fijarCookieSesion(w http.ResponseWriter, r *http.Request, token string, maxAge int) {
	cookie := &http.Cookie{
		Name:     nombreCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   esPeticionSegura(r),
		SameSite: http.SameSiteLaxMode,
	}
	if maxAge > 0 {
		cookie.Expires = time.Now().Add(time.Duration(maxAge) * time.Second)
	} else {
		cookie.Expires = time.Unix(0, 0)
	}
	http.SetCookie(w, cookie)
}

// esPeticionSegura determina si se debe marcar la cookie como Secure.
func esPeticionSegura(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		return true
	}
	if env("COOKIE_SECURE", "") == "true" {
		return true
	}
	return false
}

// requerirAutenticado valida que exista una sesión activa válida en base de datos.
func requerirAutenticado(db *sql.DB, siguiente http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := extraerTokenSesion(r)
		if token == "" {
			responderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Se requiere inicio de sesión")
			return
		}

		sesion, usuario, err := authBuscarSesionYUsuario(r.Context(), db, token)
		if err != nil {
			responderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error interno al verificar la sesión")
			return
		}
		if sesion == nil || usuario == nil {
			// Sesión inexistente o eliminada: limpiar cookie
			fijarCookieSesion(w, r, "", -1)
			responderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesión inválida o expirada")
			return
		}

		ahora := time.Now()
		if sesion.ExpiraEn.Before(ahora) {
			_ = authEliminarSesion(r.Context(), db, token)
			fijarCookieSesion(w, r, "", -1)
			responderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sesión expirada")
			return
		}

		if !usuario.Activo {
			_ = authEliminarSesionesDeUsuario(r.Context(), db, usuario.ID)
			fijarCookieSesion(w, r, "", -1)
			responderError(w, http.StatusUnauthorized, "ACCOUNT_DISABLED", "Esta cuenta ha sido desactivada")
			return
		}

		// Deslizamiento de expiración: 8 horas de inactividad
		_ = authRenovarSesion(r.Context(), db, token, duracionSesion)
		fijarCookieSesion(w, r, token, int(duracionSesion.Seconds()))

		ctx := context.WithValue(r.Context(), contextKeyUsuario, usuario)
		ctx = context.WithValue(ctx, contextKeyToken, token)
		siguiente.ServeHTTP(w, r.WithContext(ctx))
	}
}

// requerirRol valida que el usuario autenticado posea uno de los roles autorizados.
func requerirRol(db *sql.DB, rolesPermitidos []string, siguiente http.HandlerFunc) http.HandlerFunc {
	return requerirAutenticado(db, func(w http.ResponseWriter, r *http.Request) {
		usuario := usuarioDesdeContexto(r.Context())
		if usuario == nil {
			responderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuario no autenticado")
			return
		}

		autorizado := false
		for _, rol := range rolesPermitidos {
			if usuario.Rol == rol {
				autorizado = true
				break
			}
		}

		if !autorizado {
			responderError(w, http.StatusForbidden, "FORBIDDEN", "No tienes permisos suficientes para realizar esta acción")
			return
		}

		siguiente.ServeHTTP(w, r)
	})
}
