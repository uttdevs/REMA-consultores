package main

// SERVIDOR HTTP PRINCIPAL (Punto de entrada de la aplicación)
// Inicializa la configuración, la base de datos, los middlewares globales
// y registra los módulos de la aplicación (contacto, auth, etc.).

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rema-backend/internal/auth"
	"rema-backend/internal/config"
	"rema-backend/internal/contacto"
	"rema-backend/internal/database"
	"rema-backend/internal/middleware"
)

func main() {
	config.Cargar(".env", "../.env")

	db, err := database.Abrir()
	if err != nil {
		log.Fatalf("no se pudo conectar a la base de datos: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/salud", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"ok":true}`))
	})

	// --- Registro de módulos de la aplicación ---
	contacto.RegistrarRutas(mux, db)
	auth.RegistrarRutas(mux, db)
	// chatbot.RegistrarRutas(mux, db)
	// blog.RegistrarRutas(mux, db)
	// nube.RegistrarRutas(mux, db)

	srv := &http.Server{
		Addr:              ":" + config.Get("PORT", "8080"),
		Handler:           middleware.CORSYSeguridad(mux),
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
