package database

// BASE COMPARTIDA: conexión y pool a MySQL/MariaDB.

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"rema-backend/internal/config"
)

// Abrir inicializa y valida el pool de conexiones a la base de datos MySQL/MariaDB.
func Abrir() (*sql.DB, error) {
	cfg := mysql.NewConfig()
	cfg.User = config.Get("DB_USER", "root")
	cfg.Passwd = os.Getenv("DB_PASS")
	cfg.Net = "tcp"
	cfg.Addr = config.Get("DB_HOST", "127.0.0.1") + ":" + config.Get("DB_PORT", "3306")
	cfg.DBName = config.Get("DB_NAME", "rema_consultores")
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
