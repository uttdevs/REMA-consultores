# REMA Backend

Servicio backend en Go para la plataforma de **REMA Consultores**.

## Estructura del Proyecto

Punto de entrada ejecutable (`cmd/`) de la lógica de negocio y utilidades internas (`internal/`):

```text
backend/
├── cmd/
│   └── api/
│       └── main.go          # Punto de entrada de la aplicación HTTP
├── internal/
│   ├── auth/                # Módulo de autenticación y gestión de usuarios
│   │   ├── auth_test.go     # Tests unitarios del módulo auth
│   │   ├── crypto.go        # Hasheo con Argon2id, tokens y UUIDs
│   │   ├── db.go            # Consultas SQL para usuarios y sesiones
│   │   ├── handlers.go      # Endpoints HTTP (/auth/login, /usuarios, etc.)
│   │   ├── limitador.go     # Rate limiting contra ataques de fuerza bruta
│   │   ├── middleware.go    # Autenticación por sesión y control de roles
│   │   └── modelos.go       # Estructuras de dominio y DTOs
│   ├── config/
│   │   └── config.go        # Carga de variables de entorno (.env)
│   ├── contacto/            # Módulo de contacto y prospectos
│   │   ├── contacto.go      # Endpoint POST /api/contacto y validaciones
│   │   ├── correo.go        # Notificaciones automáticas por SMTP
│   │   └── db.go            # Persistencia en mensajes_contacto
│   ├── database/
│   │   └── database.go      # Conexión y pool a MySQL/MariaDB
│   └── middleware/
│       └── middleware.go    # Middlewares globales (CORS y cabeceras de seguridad)
├── go.mod
├── go.sum
└── .env                     # Variables de entorno locales
```

## Comandos Principales

### 1. Ejecutar en desarrollo
Desde el directorio `backend/`:
```bash
go run ./cmd/api
```

### 2. Compilar el binario de producción
```bash
go build -o rema-backend ./cmd/api
```

### 3. Ejecutar los tests unitarios
```bash
go test -v ./...
```
