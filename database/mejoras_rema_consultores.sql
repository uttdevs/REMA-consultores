-- Mejoras propuestas para `rema_consultores`
-- Todas son aditivas o correcciones seguras; no eliminan datos existentes.
-- Revisar y aplicar en un entorno de pruebas antes de producción.

-- ---------------------------------------------------------
-- 1. publicaciones: soportar programación y SEO completo (RF-05, RF-07)
-- ---------------------------------------------------------

ALTER TABLE `publicaciones`
  MODIFY `estado` ENUM('borrador','programado','publicado') NOT NULL DEFAULT 'borrador';

ALTER TABLE `publicaciones`
  ADD COLUMN `meta_titulo` VARCHAR(200) DEFAULT NULL AFTER `descripcion`,
  ADD COLUMN `meta_descripcion` VARCHAR(300) DEFAULT NULL AFTER `meta_titulo`;

-- imagen_portada como NOT NULL obliga a tenerla desde la creación;
-- se vuelve opcional para permitir guardar un borrador sin imagen todavía.
ALTER TABLE `publicaciones`
  MODIFY `imagen_portada` VARCHAR(255) DEFAULT NULL;

-- ---------------------------------------------------------
-- 2. citas: duración y consistencia presencial/en línea
-- ---------------------------------------------------------

ALTER TABLE `citas`
  ADD COLUMN `duracion_minutos` INT(11) NOT NULL DEFAULT 60 AFTER `fecha_hora`;

-- Si la cita es presencial, debe tener oficina asignada.
-- (MariaDB 10.4 sí aplica CHECK constraints, no solo las parsea.)
ALTER TABLE `citas`
  ADD CONSTRAINT `chk_citas_oficina_presencial`
  CHECK (`modalidad` <> 'presencial' OR `oficina_id` IS NOT NULL);

-- ---------------------------------------------------------
-- 3. Tablas de almacenamiento privado (RF-08, RF-09) — no existían
-- ---------------------------------------------------------

CREATE TABLE `carpetas` (
  `id` INT(11) NOT NULL AUTO_INCREMENT,
  `nombre` VARCHAR(150) NOT NULL,
  `carpeta_padre_id` INT(11) DEFAULT NULL,
  `creado_por` INT(11) NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP(),
  PRIMARY KEY (`id`),
  KEY `idx_carpetas_padre` (`carpeta_padre_id`),
  KEY `idx_carpetas_creado_por` (`creado_por`),
  CONSTRAINT `fk_carpetas_padre` FOREIGN KEY (`carpeta_padre_id`) REFERENCES `carpetas` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_carpetas_creado_por` FOREIGN KEY (`creado_por`) REFERENCES `usuarios` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `archivos` (
  `id` INT(11) NOT NULL AUTO_INCREMENT,
  `nombre` VARCHAR(255) NOT NULL,
  `carpeta_id` INT(11) DEFAULT NULL,
  `ruta_bucket` VARCHAR(500) NOT NULL,
  `mime_type` VARCHAR(120) DEFAULT NULL,
  `tamano_bytes` BIGINT(20) DEFAULT NULL,
  `subido_por` INT(11) NOT NULL,
  -- disponible=0 mientras se espera la confirmación de subida a S3 (ver contrato de API)
  `disponible` TINYINT(1) NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP(),
  PRIMARY KEY (`id`),
  KEY `idx_archivos_carpeta` (`carpeta_id`),
  KEY `idx_archivos_subido_por` (`subido_por`),
  CONSTRAINT `fk_archivos_carpeta` FOREIGN KEY (`carpeta_id`) REFERENCES `carpetas` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_archivos_subido_por` FOREIGN KEY (`subido_por`) REFERENCES `usuarios` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- La tabla `bitacora` ya existente cubre RF-10 (quién subió/descargó/eliminó),
-- usando tabla='archivos' o tabla='carpetas' y registro_id. No se necesita tabla aparte.

-- ---------------------------------------------------------
-- 4. usuarios: simplificar roles (decisión confirmada)
-- ---------------------------------------------------------
-- "cliente" era redundante: `citas` ya admite reservas sin cuenta
-- (usuario_id nullable + nombre/correo/telefono en la propia fila).
-- usuarios.rol queda solo para quien entra al panel: admin y editor.
-- No hay registro público; las cuentas las crea un admin (ver RF-04).

ALTER TABLE `usuarios`
  MODIFY `rol` ENUM('admin','editor') NOT NULL DEFAULT 'editor';
