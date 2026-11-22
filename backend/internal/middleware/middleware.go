package middleware

// BASE COMPARTIDA: cabeceras de seguridad y CORS.

import (
	"net/http"
	"strings"

	"rema-backend/internal/config"
)

// CORSYSeguridad aplica cabeceras de seguridad y CORS a todas las peticiones entrantes.
func CORSYSeguridad(siguiente http.Handler) http.Handler {
	permitidos := map[string]bool{}
	for _, o := range strings.Split(config.Get("CORS_ORIGINS", "http://127.0.0.1:5500,http://localhost:5500,http://localhost:5173,http://127.0.0.1:5173"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			permitidos[o] = true
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Cache-Control", "no-store")
		h.Set("Referrer-Policy", "no-referrer")

		if origen := r.Header.Get("Origin"); origen != "" && permitidos[origen] {
			h.Set("Access-Control-Allow-Origin", origen)
			h.Set("Vary", "Origin")
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			h.Set("Access-Control-Max-Age", "600")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		siguiente.ServeHTTP(w, r)
	})
}
