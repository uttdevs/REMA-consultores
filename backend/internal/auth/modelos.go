package auth

import "time"

// Roles válidos para el panel de administración.
const (
	RolAdmin    = "admin"
	RolEditor   = "editor"
	RolEmpleado = "empleado"
)

// Usuario representa un registro en la tabla usuarios.
type Usuario struct {
	ID           string    `json:"id"`
	Nombre       string    `json:"nombre"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Rol          string    `json:"rol"`
	Activo       bool      `json:"activo"`
	CreadoEn     time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"-"`
}

// UsuarioPublico representa la estructura de usuario enviada al frontend (sin datos sensibles).
type UsuarioPublico struct {
	ID       string `json:"id"`
	Nombre   string `json:"nombre"`
	Email    string `json:"email"`
	Rol      string `json:"rol"`
	Activo   *bool  `json:"activo,omitempty"`
	CreadoEn *time.Time `json:"creado_en,omitempty"`
}

// ConvertirAPublico crea una vista segura del usuario.
func (u *Usuario) ConvertirAPublico(incluirMeta bool) UsuarioPublico {
	pub := UsuarioPublico{
		ID:     u.ID,
		Nombre: u.Nombre,
		Email:  u.Email,
		Rol:    u.Rol,
	}
	if incluirMeta {
		activo := u.Activo
		pub.Activo = &activo
		creado := u.CreadoEn
		pub.CreadoEn = &creado
	}
	return pub
}

// Sesion representa una sesión activa en la tabla sesiones.
type Sesion struct {
	ID           string    `json:"id"`
	UsuarioID    string    `json:"usuario_id"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent"`
	CreadoEn     time.Time `json:"creado_en"`
	UltimoAcceso time.Time `json:"ultimo_acceso"`
	ExpiraEn     time.Time `json:"expira_en"`
}

// LoginEntrada es el payload esperado en POST /auth/login.
type LoginEntrada struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRespuesta es el cuerpo devuelto en 200 OK tras iniciar sesión o consultar /auth/me.
type LoginRespuesta struct {
	User UsuarioPublico `json:"user"`
}

// UsuarioCrearEntrada es el payload para POST /usuarios.
type UsuarioCrearEntrada struct {
	Nombre           string `json:"nombre"`
	Email            string `json:"email"`
	Rol              string `json:"rol"`
	PasswordTemporal string `json:"password_temporal"`
}

// UsuarioActualizarEntrada es el payload parcial para PATCH /usuarios/:id.
type UsuarioActualizarEntrada struct {
	Nombre *string `json:"nombre"`
	Rol    *string `json:"rol"`
	Activo *bool   `json:"activo"`
}

// ListaPaginada es el formato estándar para respuestas de listado según contrato.
type ListaPaginada[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
}

// ErrorRespuesta implementa el formato estándar de error del contrato:
// { "error": { "code": "string", "message": "string" } }
type ErrorRespuesta struct {
	Error DetalleError `json:"error"`
}

type DetalleError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
