package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"rema-backend/internal/config"
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
		creado_en TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		actualizado_en TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_usuarios_rol (rol),
		INDEX idx_usuarios_activo (activo)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	if _, err := db.ExecContext(ctx, sqlUsuarios); err != nil {
		return fmt.Errorf("error al verificar/crear tabla usuarios: %w", err)
	}

	sqlSesiones := `
	CREATE TABLE IF NOT EXISTS sesiones (
		id VARCHAR(128) NOT NULL PRIMARY KEY,
		usuario_id CHAR(36) NOT NULL,
		ip VARCHAR(45) NOT NULL DEFAULT '',
		user_agent VARCHAR(500) NOT NULL DEFAULT '',
		creado_en TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		ultimo_acceso TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expira_en TIMESTAMP NOT NULL,
		INDEX idx_sesiones_usuario (usuario_id),
		INDEX idx_sesiones_expira (expira_en),
		CONSTRAINT fk_sesiones_usuario FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`

	if _, err := db.ExecContext(ctx, sqlSesiones); err != nil {
		return fmt.Errorf("error al verificar/crear tabla sesiones: %w", err)
	}

	// Sembrar usuario administrador por defecto si la tabla está vacía
	var totalUsuarios int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM usuarios").Scan(&totalUsuarios)
	if err == nil && totalUsuarios == 0 {
		adminEmail := config.Get("ADMIN_INITIAL_EMAIL", "admin@remaconsultores.com")
		adminPass := config.Get("ADMIN_INITIAL_PASSWORD", "Admin1234#")
		hash, err := HashPassword(adminPass)
		if err == nil {
			id := GenerarUUID()
			_, err = db.ExecContext(ctx,
				"INSERT INTO usuarios (id, nombre, email, password_hash, rol, activo) VALUES (?, ?, ?, ?, 'admin', 1)",
				id, "Administrador REMA", adminEmail, hash,
			)
			if err != nil {
				log.Printf("auth: no se pudo crear usuario admin inicial: %v", err)
			} else {
				log.Printf("auth: usuario admin inicial creado con éxito (%s)", adminEmail)
			}
		}
	}

	return nil
}

// authBuscarUsuarioPorEmail consulta un usuario por su correo electrónico.
func authBuscarUsuarioPorEmail(ctx context.Context, db *sql.DB, email string) (*Usuario, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	u := &Usuario{}
	query := `
	SELECT id, nombre, email, password_hash, rol, activo, creado_en, actualizado_en
	FROM usuarios
	WHERE email = ? LIMIT 1`

	row := db.QueryRowContext(ctx, query, strings.ToLower(strings.TrimSpace(email)))
	err := row.Scan(
		&u.ID,
		&u.Nombre,
		&u.Email,
		&u.PasswordHash,
		&u.Rol,
		&u.Activo,
		&u.CreadoEn,
		&u.ActualizadoEn,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al consultar usuario por email: %w", err)
	}
	return u, nil
}

// authBuscarUsuarioPorID consulta un usuario por su identificador UUID.
func authBuscarUsuarioPorID(ctx context.Context, db *sql.DB, id string) (*Usuario, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	u := &Usuario{}
	query := `
	SELECT id, nombre, email, password_hash, rol, activo, creado_en, actualizado_en
	FROM usuarios
	WHERE id = ? LIMIT 1`

	row := db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&u.ID,
		&u.Nombre,
		&u.Email,
		&u.PasswordHash,
		&u.Rol,
		&u.Activo,
		&u.CreadoEn,
		&u.ActualizadoEn,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al consultar usuario por id: %w", err)
	}
	return u, nil
}

// authCrearSesion inserta una nueva sesión en la base de datos.
func authCrearSesion(ctx context.Context, db *sql.DB, s *Sesion) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
	INSERT INTO sesiones (id, usuario_id, ip, user_agent, expira_en)
	VALUES (?, ?, ?, ?, ?)`

	_, err := db.ExecContext(ctx, query,
		s.ID,
		s.UsuarioID,
		s.IP,
		s.UserAgent,
		s.ExpiraEn,
	)
	if err != nil {
		return fmt.Errorf("error al registrar sesión: %w", err)
	}
	return nil
}

// authBuscarSesionYUsuario valida un token de sesión y retorna la sesión junto con su usuario si está activo.
func authBuscarSesionYUsuario(ctx context.Context, db *sql.DB, token string) (*Sesion, *Usuario, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
	SELECT 
		s.id, s.usuario_id, s.ip, s.user_agent, s.creado_en, s.ultimo_acceso, s.expira_en,
		u.id, u.nombre, u.email, u.password_hash, u.rol, u.activo, u.creado_en, u.actualizado_en
	FROM sesiones s
	INNER JOIN usuarios u ON u.id = s.usuario_id
	WHERE s.id = ? AND s.expira_en > NOW()
	LIMIT 1`

	s := &Sesion{}
	u := &Usuario{}

	row := db.QueryRowContext(ctx, query, token)
	err := row.Scan(
		&s.ID,
		&s.UsuarioID,
		&s.IP,
		&s.UserAgent,
		&s.CreadoEn,
		&s.UltimoAcceso,
		&s.ExpiraEn,
		&u.ID,
		&u.Nombre,
		&u.Email,
		&u.PasswordHash,
		&u.Rol,
		&u.Activo,
		&u.CreadoEn,
		&u.ActualizadoEn,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("error al validar sesión: %w", err)
	}

	// Si el usuario fue desactivado, la sesión no es válida
	if !u.Activo {
		return nil, nil, nil
	}

	// Refrescar asíncronamente el último acceso para no penalizar la latencia
	go func(tok string) {
		ctxAsync, cancelAsync := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelAsync()
		_, _ = db.ExecContext(ctxAsync, "UPDATE sesiones SET ultimo_acceso = NOW() WHERE id = ?", tok)
	}(token)

	return s, u, nil
}

// authRenovarSesion amplía el tiempo de expiración de una sesión activa (sliding session).
func authRenovarSesion(ctx context.Context, db *sql.DB, token string, duracion time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	nuevaExpiracion := time.Now().Add(duracion)
	_, err := db.ExecContext(ctx, "UPDATE sesiones SET expira_en = ? WHERE id = ?", nuevaExpiracion, token)
	return err
}

// authEliminarSesion remueve una sesión por su token (usado en logout).
func authEliminarSesion(ctx context.Context, db *sql.DB, token string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, "DELETE FROM sesiones WHERE id = ?", token)
	return err
}

// authEliminarSesionesDeUsuario revoca todas las sesiones de un usuario (para reseteo o bloqueo).
func authEliminarSesionesDeUsuario(ctx context.Context, db *sql.DB, usuarioID string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, "DELETE FROM sesiones WHERE usuario_id = ?", usuarioID)
	return err
}

// authListarUsuarios consulta la lista paginada de usuarios para el panel administrativo.
func authListarUsuarios(ctx context.Context, db *sql.DB, page, limit int) ([]UsuarioPublico, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM usuarios").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error al contar usuarios: %w", err)
	}

	query := `
	SELECT id, nombre, email, rol, activo, creado_en
	FROM usuarios
	ORDER BY creado_en DESC
	LIMIT ? OFFSET ?`

	rows, err := db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error al listar usuarios: %w", err)
	}
	defer rows.Close()

	items := make([]UsuarioPublico, 0, limit)
	for rows.Next() {
		var u Usuario
		if err := rows.Scan(&u.ID, &u.Nombre, &u.Email, &u.Rol, &u.Activo, &u.CreadoEn); err != nil {
			return nil, 0, fmt.Errorf("error al leer registro de usuario: %w", err)
		}
		items = append(items, u.ConvertirAPublico(true))
	}

	return items, total, nil
}

// authCrearUsuario inserta un nuevo usuario en la base de datos.
func authCrearUsuario(ctx context.Context, db *sql.DB, u *Usuario) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
	INSERT INTO usuarios (id, nombre, email, password_hash, rol, activo)
	VALUES (?, ?, ?, ?, ?, ?)`

	activoNum := 0
	if u.Activo {
		activoNum = 1
	}

	_, err := db.ExecContext(ctx, query,
		u.ID,
		u.Nombre,
		strings.ToLower(strings.TrimSpace(u.Email)),
		u.PasswordHash,
		u.Rol,
		activoNum,
	)
	if err != nil {
		return err
	}
	return nil
}

// authActualizarUsuario actualiza parcialmente los datos de un usuario (PATCH).
func authActualizarUsuario(ctx context.Context, db *sql.DB, id string, nombre *string, rol *string, activo *bool) (*Usuario, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	setClauses := make([]string, 0, 3)
	args := make([]any, 0, 4)

	if nombre != nil {
		setClauses = append(setClauses, "nombre = ?")
		args = append(args, strings.TrimSpace(*nombre))
	}
	if rol != nil {
		setClauses = append(setClauses, "rol = ?")
		args = append(args, *rol)
	}
	if activo != nil {
		setClauses = append(setClauses, "activo = ?")
		val := 0
		if *activo {
			val = 1
		}
		args = append(args, val)
	}

	if len(setClauses) == 0 {
		return authBuscarUsuarioPorID(ctx, db, id)
	}

	query := fmt.Sprintf("UPDATE usuarios SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, id)

	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar usuario: %w", err)
	}

	afectados, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if afectados == 0 {
		// Validar si existía
		u, errCheck := authBuscarUsuarioPorID(ctx, db, id)
		if errCheck != nil {
			return nil, errCheck
		}
		if u == nil {
			return nil, sql.ErrNoRows
		}
	}

	return authBuscarUsuarioPorID(ctx, db, id)
}

// authEliminarUsuario elimina a un usuario permanentemente.
func authEliminarUsuario(ctx context.Context, db *sql.DB, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	res, err := db.ExecContext(ctx, "DELETE FROM usuarios WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error al eliminar usuario: %w", err)
	}

	afectados, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if afectados == 0 {
		return sql.ErrNoRows
	}

	return nil
}
