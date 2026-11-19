-- =====================================================================
-- REMA Consultores - Esquema de Base de Datos (MySQL / MariaDB)
-- Módulos:
--   - Contacto (RF-02)
--   - Autenticación y roles / Panel admin (RF-04)
-- =====================================================================

CREATE DATABASE IF NOT EXISTS rema_consultores
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE rema_consultores;

-- ---------------------------------------------------------------------
-- 1. Tabla: mensajes_contacto (RF-02)
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mensajes_contacto (
    id INT AUTO_INCREMENT PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    empresa VARCHAR(120) NULL,
    telefono VARCHAR(20) NULL,
    correo VARCHAR(150) NOT NULL,
    asunto VARCHAR(150) NOT NULL,
    mensaje TEXT NOT NULL,
    acepta_privacidad TINYINT(1) NOT NULL DEFAULT 0,
    acepta_privacidad_en DATETIME NULL,
    creado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_contacto_correo (correo),
    INDEX idx_contacto_creado (creado_en)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
-- 2. Tabla: usuarios (RF-04)
-- Roles: admin, editor, empleado
-- Las contraseñas se almacenan únicamente como hash Argon2id ($argon2id$...)
-- ---------------------------------------------------------------------
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
    INDEX idx_usuarios_activo (activo),
    INDEX idx_usuarios_creado (creado_en)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
-- 3. Tabla: sesiones (RF-04)
-- Maneja cookies de sesión HttpOnly con 8 horas de inactividad
-- ---------------------------------------------------------------------
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
-- 4. Semilla: Usuario Administrador Inicial
-- Correo: admin@remaconsultores.com
-- Contraseña temporal: Admin1234#
-- Hash generado con Argon2id (m=65536, t=1, p=2)
-- NOTA: El backend también creará automáticamente este administrador si la
-- tabla 'usuarios' se encuentra vacía al iniciar.
-- ---------------------------------------------------------------------
INSERT IGNORE INTO usuarios (id, nombre, email, password_hash, rol, activo, creado_en, actualizado_en)
VALUES (
    'a0000000-0000-4000-8000-000000000001',
    'Administrador REMA',
    'admin@remaconsultores.com',
    '$argon2id$v=19$m=65536,t=1,p=2$B5dYn4hK1PzJ8wQ2tLv8mA$w8E6sU2jL5P9rT3vX1bN8qM4yK0hF7eD2aZ9cW3vS1u',
    'admin',
    1,
    NOW(),
    NOW()
);
