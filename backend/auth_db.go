package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// authCrearTablasSiNoExisten asegura que las tablas usuarios y sesiones existan en MySQL/MariaDB.
func authCrearTablasSiNoExisten(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	sqlUsuarios := `
	CREATE TABLE IF NOT EXISTS usuarios (
		id CHAR(36) NOT NULL PRIMARY KEY,
		nombre VARCHAR(150) NOT NULL,
		email VARCHAR(191) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		rol ENUM('admin', 'editor', 'empleado') NOT NULL DEFAULT 'editor',
		activo TINYINT(1) NOT NULL DEFAULT 1,
		creado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		actualizado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_usuarios_rol (rol),
		INDEX idx_usuarios_activo (activo)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	if _, err := db.ExecContext(ctx, sqlUsuarios); err != nil {
		return fmt.Errorf("error al crear tabla usuarios: %w", err)
	}

	sqlSesiones := `
	CREATE TABLE IF NOT EXISTS sesiones (
		id VARCHAR(64) NOT NULL PRIMARY KEY,
		usuario_id CHAR(36) NOT NULL,
		ip VARCHAR(45) NULL,
		user_agent VARCHAR(255) NULL,
		creado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		ultimo_acceso DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expira_en DATETIME NOT NULL,
		INDEX idx_sesiones_usuario (usuario_id),
		INDEX idx_sesiones_expira (expira_en),
		CONSTRAINT fk_sesiones_usuario FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	if _, err := db.ExecContext(ctx, sqlSesiones); err != nil {
		return fmt.Errorf("error al crear tabla sesiones: %w", err)
	}

	// Semilla automática si no existe ningún usuario
	var totalUsuarios int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM usuarios").Scan(&totalUsuarios); err == nil && totalUsuarios == 0 {
		adminEmail := env("ADMIN_INITIAL_EMAIL", "admin@remaconsultores.com")
		adminPass := env("ADMIN_INITIAL_PASSWORD", "Admin1234#")
		hash, err := hashPassword(adminPass)
		if err == nil {
			id := generarUUID()
			_, err = db.ExecContext(ctx, `
				INSERT INTO usuarios (id, nombre, email, password_hash, rol, activo, creado_en, actualizado_en)
				VALUES (?, ?, ?, ?, 'admin', 1, NOW(), NOW())
			`, id, "Administrador Inicial", adminEmail, hash)
			if err == nil {
				log.Printf("auth: administrador inicial creado con éxito (%s)", adminEmail)
			}
		}
	}

	return nil
}

// authBuscarUsuarioPorEmail localiza un usuario activo o inactivo por correo electrónico.
func authBuscarUsuarioPorEmail(ctx context.Context, db *sql.DB, email string) (*Usuario, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	emailLimpio := strings.ToLower(strings.TrimSpace(email))
	query := `SELECT id, nombre, email, password_hash, rol, activo, creado_en, actualizado_en FROM usuarios WHERE email = ?`

	var u Usuario
	var activoInt int
	err := db.QueryRowContext(ctx, query, emailLimpio).Scan(
		&u.ID,
		&u.Nombre,
		&u.Email,
		&u.PasswordHash,
		&u.Rol,
		&activoInt,
		&u.CreadoEn,
		&u.ActualizadoEn,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u.Activo = activoInt == 1
	return &u, nil
}

// authBuscarUsuarioPorID localiza un usuario por su UUID.
func authBuscarUsuarioPorID(ctx context.Context, db *sql.DB, id string) (*Usuario, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, nombre, email, password_hash, rol, activo, creado_en, actualizado_en FROM usuarios WHERE id = ?`

	var u Usuario
	var activoInt int
	err := db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Nombre,
		&u.Email,
		&u.PasswordHash,
		&u.Rol,
		&activoInt,
		&u.CreadoEn,
		&u.ActualizadoEn,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u.Activo = activoInt == 1
	return &u, nil
}

// authCrearSesion guarda una nueva sesión en base de datos.
func authCrearSesion(ctx context.Context, db *sql.DB, s *Sesion) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `INSERT INTO sesiones (id, usuario_id, ip, user_agent, creado_en, ultimo_acceso, expira_en)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := db.ExecContext(ctx, query,
		s.ID,
		s.UsuarioID,
		contactoNulo(s.IP),
		contactoNulo(s.UserAgent),
		s.CreadoEn,
		s.UltimoAcceso,
		s.ExpiraEn,
	)
	return err
}

// authBuscarSesionYUsuario valida el token y retorna la sesión junto a la información del usuario.
func authBuscarSesionYUsuario(ctx context.Context, db *sql.DB, token string) (*Sesion, *Usuario, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT s.id, s.usuario_id, s.ip, s.user_agent, s.creado_en, s.ultimo_acceso, s.expira_en,
		       u.id, u.nombre, u.email, u.password_hash, u.rol, u.activo, u.creado_en, u.actualizado_en
		FROM sesiones s
		JOIN usuarios u ON s.usuario_id = u.id
		WHERE s.id = ?
	`

	var s Sesion
	var ip, ua sql.NullString
	var u Usuario
	var activoInt int

	err := db.QueryRowContext(ctx, query, token).Scan(
		&s.ID,
		&s.UsuarioID,
		&ip,
		&ua,
		&s.CreadoEn,
		&s.UltimoAcceso,
		&s.ExpiraEn,
		&u.ID,
		&u.Nombre,
		&u.Email,
		&u.PasswordHash,
		&u.Rol,
		&activoInt,
		&u.CreadoEn,
		&u.ActualizadoEn,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	if ip.Valid {
		s.IP = ip.String
	}
	if ua.Valid {
		s.UserAgent = ua.String
	}
	u.Activo = activoInt == 1

	return &s, &u, nil
}

// authRenovarSesion actualiza el último acceso y extiende la fecha de expiración por inactividad.
func authRenovarSesion(ctx context.Context, db *sql.DB, token string, duracion time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	nuevaExpiracion := time.Now().Add(duracion)
	query := `UPDATE sesiones SET ultimo_acceso = NOW(), expira_en = ? WHERE id = ?`
	_, err := db.ExecContext(ctx, query, nuevaExpiracion, token)
	return err
}

// authEliminarSesion borra la sesión correspondiente.
func authEliminarSesion(ctx context.Context, db *sql.DB, token string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `DELETE FROM sesiones WHERE id = ?`
	_, err := db.ExecContext(ctx, query, token)
	return err
}

// authEliminarSesionesDeUsuario borra todas las sesiones abiertas de un usuario (para forzar logout al desactivarlo).
func authEliminarSesionesDeUsuario(ctx context.Context, db *sql.DB, usuarioID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `DELETE FROM sesiones WHERE usuario_id = ?`
	_, err := db.ExecContext(ctx, query, usuarioID)
	return err
}

// authListarUsuarios devuelve una lista paginada de usuarios para el panel admin.
func authListarUsuarios(ctx context.Context, db *sql.DB, page, limit int) ([]UsuarioPublico, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM usuarios").Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, nombre, email, rol, activo, creado_en FROM usuarios ORDER BY creado_en DESC LIMIT ? OFFSET ?`
	rows, err := db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]UsuarioPublico, 0)
	for rows.Next() {
		var u Usuario
		var activoInt int
		if err := rows.Scan(&u.ID, &u.Nombre, &u.Email, &u.Rol, &activoInt, &u.CreadoEn); err != nil {
			return nil, 0, err
		}
		u.Activo = activoInt == 1
		items = append(items, u.ConvertirAPublico(true))
	}

	return items, total, rows.Err()
}

// authCrearUsuario inserta un nuevo usuario en la base de datos.
func authCrearUsuario(ctx context.Context, db *sql.DB, u *Usuario) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `INSERT INTO usuarios (id, nombre, email, password_hash, rol, activo, creado_en, actualizado_en)
	          VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`

	activoInt := 0
	if u.Activo {
		activoInt = 1
	}

	_, err := db.ExecContext(ctx, query,
		u.ID,
		u.Nombre,
		strings.ToLower(strings.TrimSpace(u.Email)),
		u.PasswordHash,
		u.Rol,
		activoInt,
	)
	return err
}

// authActualizarUsuario actualiza de forma dinámica los campos nombre, rol y/o activo.
func authActualizarUsuario(ctx context.Context, db *sql.DB, id string, nombre *string, rol *string, activo *bool) (*Usuario, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	setClauses := make([]string, 0)
	args := make([]any, 0)

	if nombre != nil {
		setClauses = append(setClauses, "nombre = ?")
		args = append(args, strings.TrimSpace(*nombre))
	}
	if rol != nil {
		setClauses = append(setClauses, "rol = ?")
		args = append(args, *rol)
	}
	if activo != nil {
		activoInt := 0
		if *activo {
			activoInt = 1
		}
		setClauses = append(setClauses, "activo = ?")
		args = append(args, activoInt)
	}

	if len(setClauses) == 0 {
		return authBuscarUsuarioPorID(ctx, db, id)
	}

	setClauses = append(setClauses, "actualizado_en = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE usuarios SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		// Validar si existe el usuario
		u, err := authBuscarUsuarioPorID(ctx, db, id)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, sql.ErrNoRows
		}
		return u, nil
	}

	return authBuscarUsuarioPorID(ctx, db, id)
}

// authEliminarUsuario borra un usuario por su ID.
func authEliminarUsuario(ctx context.Context, db *sql.DB, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := db.ExecContext(ctx, "DELETE FROM usuarios WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
