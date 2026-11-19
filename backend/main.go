package main

// BASE COMPARTIDA: servidor HTTP, conexión a MySQL/MariaDB, CORS y cabeceras.
// Cada área del equipo se conecta con UNA línea: registrarXxx(mux, db).

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-sql-driver/mysql"
)

func abrirBD() (*sql.DB, error) {
	cfg := mysql.NewConfig()
	cfg.User = env("DB_USER", "root")
	cfg.Passwd = os.Getenv("DB_PASS")
	cfg.Net = "tcp"
	cfg.Addr = env("DB_HOST", "127.0.0.1") + ":" + env("DB_PORT", "3306")
	cfg.DBName = env("DB_NAME", "rema_consultores")
	cfg.Params = map[string]string{"charset": "utf8mb4"}
	cfg.Collation = "utf8mb4_unicode_ci"
	cfg.ParseTime = true
	cfg.Timeout = 5 * time.Second
	cfg.ReadTimeout = 10 * time.Second
	cfg.WriteTimeout = 10 * time.Second
	// Importante: NO activar multiStatements (evita ejecutar varias sentencias).

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// middleware aplica CORS y cabeceras de seguridad a todas las rutas.
func middleware(siguiente http.Handler) http.Handler {
	permitidos := map[string]bool{}
	for _, o := range strings.Split(env("CORS_ORIGINS", "http://127.0.0.1:5500,http://localhost:5500,http://localhost:5173,http://127.0.0.1:5173"), ",") {
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

func main() {
	cargarEnv(".env")

	db, err := abrirBD()
	if err != nil {
		log.Fatalf("no se pudo conectar a la base de datos: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/salud", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"ok":true}`))
	})

	// --- Aquí se registra cada área del equipo ---
	registrarContacto(mux, db)
	// registrarChatbot(mux, db)
	registrarPanel(mux, db)
	// registrarBlog(mux, db)
	// registrarNube(mux, db)

	srv := &http.Server{
		Addr:              ":" + env("PORT", "8080"),
		Handler:           middleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 14,
	}

	salir := make(chan os.Signal, 1)
	signal.Notify(salir, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-salir
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	cert, key := os.Getenv("TLS_CERT"), os.Getenv("TLS_KEY")
	if cert != "" && key != "" {
		log.Printf("API en https://localhost%s", srv.Addr)
		err = srv.ListenAndServeTLS(cert, key)
	} else {
		log.Printf("API en http://localhost%s (sin HTTPS: en producción usa TLS_CERT/TLS_KEY o un proxy con HTTPS)", srv.Addr)
		err = srv.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}