CREATE DATABASE IF NOT EXISTS `correccion_rema` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `correccion_rema`;

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

CREATE TABLE `autores` (
  `id` int(11) NOT NULL,
  `nombre` varchar(120) NOT NULL,
  `tipo` enum('persona','empresa') NOT NULL DEFAULT 'persona'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `bitacora` (
  `id` bigint(20) NOT NULL,
  `usuario_id` int(11) DEFAULT NULL,
  `accion` varchar(50) NOT NULL,
  `tabla` varchar(60) DEFAULT NULL,
  `registro_id` int(11) DEFAULT NULL,
  `detalle` json DEFAULT NULL,
  `ip` varchar(45) DEFAULT NULL,
  `user_agent` varchar(255) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `categorias_blog` (
  `id` int(11) NOT NULL,
  `nombre` varchar(80) NOT NULL,
  `slug` varchar(100) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `citas` (
  `id` int(11) NOT NULL,
  `usuario_id` int(11) DEFAULT NULL,
  `nombre` varbinary(512) NOT NULL,
  `correo` varbinary(768) NOT NULL,
  `correo_hash` char(64) NOT NULL,
  `telefono` varbinary(128) DEFAULT NULL,
  `servicio_id` int(11) NOT NULL,
  `oficina_id` int(11) DEFAULT NULL,
  `modalidad` enum('presencial','en_linea') NOT NULL,
  `fecha_hora` datetime NOT NULL,
  `duracion_minutos` int(11) NOT NULL DEFAULT 60,
  `fecha_fin` datetime GENERATED ALWAYS AS (`fecha_hora` + interval `duracion_minutos` minute) PERSISTENT,
  `estado` enum('pendiente','confirmada','cancelada','completada') NOT NULL DEFAULT 'pendiente',
  `notas` mediumblob DEFAULT NULL,
  `acepta_privacidad` tinyint(1) NOT NULL DEFAULT 0,
  `acepta_privacidad_en` datetime DEFAULT NULL,
  `slot_key` varchar(40) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  `updated_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  CONSTRAINT `chk_citas_oficina_presencial` CHECK (`modalidad` <> 'presencial' OR `oficina_id` IS NOT NULL),
  CONSTRAINT `chk_citas_duracion` CHECK (`duracion_minutos` BETWEEN 15 AND 480),
  CONSTRAINT `chk_citas_privacidad` CHECK (`acepta_privacidad` = 0 OR `acepta_privacidad_en` IS NOT NULL)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `configuracion_sitio` (
  `id` int(11) NOT NULL,
  `clave` varchar(100) NOT NULL,
  `valor` text DEFAULT NULL,
  `descripcion` varchar(200) DEFAULT NULL,
  `actualizado_por` int(11) DEFAULT NULL,
  `updated_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `mensajes_contacto` (
  `id` int(11) NOT NULL,
  `usuario_id` int(11) DEFAULT NULL,
  `nombre` varbinary(512) NOT NULL,
  `empresa` varbinary(512) DEFAULT NULL,
  `telefono` varbinary(128) DEFAULT NULL,
  `correo` varbinary(768) NOT NULL,
  `correo_hash` char(64) NOT NULL,
  `asunto` varbinary(768) NOT NULL,
  `mensaje` mediumblob NOT NULL,
  `acepta_privacidad` tinyint(1) NOT NULL DEFAULT 0,
  `acepta_privacidad_en` datetime DEFAULT NULL,
  `estado` enum('nuevo','leido','atendido') NOT NULL DEFAULT 'nuevo',
  `servicio_id` int(11) DEFAULT NULL,
  `atendido_por` int(11) DEFAULT NULL,
  `atendido_en` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  CONSTRAINT `chk_mensajes_privacidad` CHECK (`acepta_privacidad` = 0 OR `acepta_privacidad_en` IS NOT NULL)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `oficinas` (
  `id` int(11) NOT NULL,
  `ciudad` varchar(80) NOT NULL,
  `direccion` varchar(255) DEFAULT NULL,
  `telefono` varchar(20) DEFAULT NULL,
  `correo` varchar(150) DEFAULT NULL,
  `horario` varchar(150) DEFAULT NULL,
  `mapa_url` varchar(255) DEFAULT NULL,
  `activo` tinyint(1) NOT NULL DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `publicaciones` (
  `id` int(11) NOT NULL,
  `titulo` varchar(200) NOT NULL,
  `slug` varchar(220) NOT NULL,
  `descripcion` varchar(300) NOT NULL,
  `meta_titulo` varchar(200) DEFAULT NULL,
  `meta_descripcion` varchar(300) DEFAULT NULL,
  `contenido` mediumtext NOT NULL,
  `imagen_portada` varchar(255) DEFAULT NULL,
  `autor_id` int(11) NOT NULL,
  `creado_por` int(11) NOT NULL,
  `categoria_id` int(11) NOT NULL,
  `estado` enum('borrador','programado','publicado') NOT NULL DEFAULT 'borrador',
  `publicado_en` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  `updated_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  CONSTRAINT `chk_publicaciones_fecha` CHECK (`estado` = 'borrador' OR `publicado_en` IS NOT NULL)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `publicacion_imagenes` (
  `id` int(11) NOT NULL,
  `publicacion_id` int(11) NOT NULL,
  `ruta` varchar(255) NOT NULL,
  `descripcion` varchar(150) DEFAULT NULL,
  `orden` int(11) NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `redes_sociales` (
  `id` int(11) NOT NULL,
  `nombre` varchar(50) NOT NULL,
  `url` varchar(255) NOT NULL,
  `icono` varchar(80) DEFAULT NULL,
  `orden` int(11) NOT NULL DEFAULT 0,
  `activo` tinyint(1) NOT NULL DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `servicios` (
  `id` int(11) NOT NULL,
  `nombre` varchar(100) NOT NULL,
  `slug` varchar(120) NOT NULL,
  `descripcion` text DEFAULT NULL,
  `icono` varchar(80) DEFAULT NULL,
  `orden` int(11) NOT NULL DEFAULT 0,
  `activo` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  `updated_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `servicio_items` (
  `id` int(11) NOT NULL,
  `servicio_id` int(11) NOT NULL,
  `texto` varchar(200) NOT NULL,
  `orden` int(11) NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `tokens_usuario` (
  `id` int(11) NOT NULL,
  `usuario_id` int(11) NOT NULL,
  `tipo` enum('recuperar_password') NOT NULL DEFAULT 'recuperar_password',
  `token_hash` char(64) NOT NULL,
  `expira_en` datetime NOT NULL,
  `usado_en` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `usuarios` (
  `id` int(11) NOT NULL,
  `nombre` varchar(100) NOT NULL,
  `correo` varchar(150) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `rol` enum('admin','editor') NOT NULL DEFAULT 'editor',
  `activo` tinyint(1) NOT NULL DEFAULT 1,
  `intentos_fallidos` tinyint(3) UNSIGNED NOT NULL DEFAULT 0,
  `bloqueado_hasta` datetime DEFAULT NULL,
  `ultimo_acceso` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  `updated_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `valores_agregados` (
  `id` int(11) NOT NULL,
  `titulo` varchar(120) NOT NULL,
  `descripcion` text DEFAULT NULL,
  `orden` int(11) NOT NULL DEFAULT 0,
  `activo` tinyint(1) NOT NULL DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `carpetas` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `nombre` varchar(150) NOT NULL,
  `carpeta_padre_id` int(11) DEFAULT NULL,
  `creado_por` int(11) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `idx_carpetas_padre` (`carpeta_padre_id`),
  KEY `idx_carpetas_creado_por` (`creado_por`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `archivos` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `nombre` varchar(255) NOT NULL,
  `carpeta_id` int(11) DEFAULT NULL,
  `ruta_bucket` varchar(500) NOT NULL,
  `mime_type` varchar(120) DEFAULT NULL,
  `tamano_bytes` bigint(20) DEFAULT NULL,
  `sha256` char(64) DEFAULT NULL,
  `subido_por` int(11) NOT NULL,
  `disponible` tinyint(1) NOT NULL DEFAULT 0,
  `eliminado_en` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `idx_archivos_carpeta` (`carpeta_id`),
  KEY `idx_archivos_subido_por` (`subido_por`),
  KEY `idx_archivos_eliminado` (`eliminado_en`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `solicitudes_privacidad` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tipo` enum('acceso','rectificacion','cancelacion','oposicion') NOT NULL,
  `nombre` varbinary(512) NOT NULL,
  `correo` varbinary(768) NOT NULL,
  `correo_hash` char(64) NOT NULL,
  `detalle` mediumblob DEFAULT NULL,
  `estado` enum('recibida','en_proceso','resuelta','rechazada') NOT NULL DEFAULT 'recibida',
  `atendido_por` int(11) DEFAULT NULL,
  `atendido_en` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `idx_solicitudes_estado` (`estado`,`created_at`),
  KEY `idx_solicitudes_correo_hash` (`correo_hash`),
  KEY `idx_solicitudes_atendido_por` (`atendido_por`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE `autores`
  ADD PRIMARY KEY (`id`);

ALTER TABLE `bitacora`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_bitacora_usuario` (`usuario_id`),
  ADD KEY `idx_bitacora_fecha` (`created_at`),
  ADD KEY `idx_bitacora_tabla_registro` (`tabla`,`registro_id`);

ALTER TABLE `categorias_blog`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_categorias_slug` (`slug`);

ALTER TABLE `citas`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_citas_slot` (`slot_key`),
  ADD KEY `idx_citas_fecha` (`fecha_hora`),
  ADD KEY `idx_citas_rango` (`fecha_hora`,`fecha_fin`),
  ADD KEY `idx_citas_estado_fecha` (`estado`,`fecha_hora`),
  ADD KEY `idx_citas_correo_hash` (`correo_hash`),
  ADD KEY `idx_citas_usuario` (`usuario_id`),
  ADD KEY `idx_citas_servicio` (`servicio_id`),
  ADD KEY `idx_citas_oficina` (`oficina_id`);

ALTER TABLE `configuracion_sitio`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_config_clave` (`clave`),
  ADD KEY `idx_config_actualizado_por` (`actualizado_por`);

ALTER TABLE `mensajes_contacto`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_mensajes_estado` (`estado`,`created_at`),
  ADD KEY `idx_mensajes_correo_hash` (`correo_hash`),
  ADD KEY `idx_mensajes_usuario` (`usuario_id`),
  ADD KEY `idx_mensajes_servicio` (`servicio_id`),
  ADD KEY `idx_mensajes_atendido_por` (`atendido_por`);

ALTER TABLE `oficinas`
  ADD PRIMARY KEY (`id`);

ALTER TABLE `publicaciones`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_publicaciones_slug` (`slug`),
  ADD KEY `idx_publicaciones_estado_fecha` (`estado`,`publicado_en`),
  ADD KEY `idx_publicaciones_categoria` (`categoria_id`),
  ADD KEY `idx_publicaciones_autor` (`autor_id`),
  ADD KEY `idx_publicaciones_creado_por` (`creado_por`),
  ADD FULLTEXT KEY `ft_publicaciones_busqueda` (`titulo`,`descripcion`,`contenido`);

ALTER TABLE `publicacion_imagenes`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_pub_imagenes` (`publicacion_id`,`orden`);

ALTER TABLE `redes_sociales`
  ADD PRIMARY KEY (`id`);

ALTER TABLE `servicios`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_servicios_slug` (`slug`);

ALTER TABLE `servicio_items`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_servicio_items` (`servicio_id`,`orden`);

ALTER TABLE `tokens_usuario`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_tokens_hash` (`token_hash`),
  ADD KEY `idx_tokens_usuario_tipo` (`usuario_id`,`tipo`),
  ADD KEY `idx_tokens_expira` (`expira_en`);

ALTER TABLE `usuarios`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_usuarios_correo` (`correo`);

ALTER TABLE `valores_agregados`
  ADD PRIMARY KEY (`id`);

ALTER TABLE `autores`              MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `bitacora`             MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT;
ALTER TABLE `categorias_blog`      MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `citas`                MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `configuracion_sitio`  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `mensajes_contacto`    MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `oficinas`             MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `publicaciones`        MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `publicacion_imagenes` MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `redes_sociales`       MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `servicios`            MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `servicio_items`       MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `tokens_usuario`       MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `usuarios`             MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `valores_agregados`    MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

ALTER TABLE `bitacora`
  ADD CONSTRAINT `fk_bitacora_usuario` FOREIGN KEY (`usuario_id`) REFERENCES `usuarios` (`id`) ON DELETE SET NULL;

ALTER TABLE `citas`
  ADD CONSTRAINT `fk_citas_oficina` FOREIGN KEY (`oficina_id`) REFERENCES `oficinas` (`id`),
  ADD CONSTRAINT `fk_citas_servicio` FOREIGN KEY (`servicio_id`) REFERENCES `servicios` (`id`),
  ADD CONSTRAINT `fk_citas_usuario` FOREIGN KEY (`usuario_id`) REFERENCES `usuarios` (`id`) ON DELETE SET NULL;

ALTER TABLE `configuracion_sitio`
  ADD CONSTRAINT `fk_config_actualizado_por` FOREIGN KEY (`actualizado_por`) REFERENCES `usuarios` (`id`) ON DELETE SET NULL;

ALTER TABLE `mensajes_contacto`
  ADD CONSTRAINT `fk_mensajes_atendido_por` FOREIGN KEY (`atendido_por`) REFERENCES `usuarios` (`id`) ON DELETE SET NULL,
  ADD CONSTRAINT `fk_mensajes_servicio` FOREIGN KEY (`servicio_id`) REFERENCES `servicios` (`id`) ON DELETE SET NULL,
  ADD CONSTRAINT `fk_mensajes_usuario` FOREIGN KEY (`usuario_id`) REFERENCES `usuarios` (`id`) ON DELETE SET NULL;

ALTER TABLE `publicaciones`
  ADD CONSTRAINT `fk_publicaciones_autor` FOREIGN KEY (`autor_id`) REFERENCES `autores` (`id`),
  ADD CONSTRAINT `fk_publicaciones_categoria` FOREIGN KEY (`categoria_id`) REFERENCES `categorias_blog` (`id`),
  ADD CONSTRAINT `fk_publicaciones_creado_por` FOREIGN KEY (`creado_por`) REFERENCES `usuarios` (`id`);

ALTER TABLE `publicacion_imagenes`
  ADD CONSTRAINT `fk_pub_imagenes_publicacion` FOREIGN KEY (`publicacion_id`) REFERENCES `publicaciones` (`id`) ON DELETE CASCADE;

ALTER TABLE `servicio_items`
  ADD CONSTRAINT `fk_servicio_items_servicio` FOREIGN KEY (`servicio_id`) REFERENCES `servicios` (`id`) ON DELETE CASCADE;

ALTER TABLE `tokens_usuario`
  ADD CONSTRAINT `fk_tokens_usuario` FOREIGN KEY (`usuario_id`) REFERENCES `usuarios` (`id`) ON DELETE CASCADE;

ALTER TABLE `carpetas`
  ADD CONSTRAINT `fk_carpetas_padre` FOREIGN KEY (`carpeta_padre_id`) REFERENCES `carpetas` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `fk_carpetas_creado_por` FOREIGN KEY (`creado_por`) REFERENCES `usuarios` (`id`);

ALTER TABLE `archivos`
  ADD CONSTRAINT `fk_archivos_carpeta` FOREIGN KEY (`carpeta_id`) REFERENCES `carpetas` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `fk_archivos_subido_por` FOREIGN KEY (`subido_por`) REFERENCES `usuarios` (`id`);

ALTER TABLE `solicitudes_privacidad`
  ADD CONSTRAINT `fk_solicitudes_atendido_por` FOREIGN KEY (`atendido_por`) REFERENCES `usuarios` (`id`) ON DELETE SET NULL;

COMMIT;

CREATE EVENT IF NOT EXISTS `ev_limpiar_tokens`
  ON SCHEDULE EVERY 1 DAY
  DO DELETE FROM `tokens_usuario` WHERE `expira_en` < NOW() - INTERVAL 1 DAY OR `usado_en` IS NOT NULL;

CREATE USER IF NOT EXISTS 'rema_web'@'localhost' IDENTIFIED BY 'CAMBIAR_ESTA_CLAVE';

GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`autores` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`categorias_blog` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`citas` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`configuracion_sitio` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`mensajes_contacto` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`oficinas` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`publicaciones` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`publicacion_imagenes` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`redes_sociales` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`servicios` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`servicio_items` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`tokens_usuario` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`usuarios` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`valores_agregados` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`carpetas` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`archivos` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON `correccion_rema`.`solicitudes_privacidad` TO 'rema_web'@'localhost';
GRANT SELECT, INSERT ON `correccion_rema`.`bitacora` TO 'rema_web'@'localhost';

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
