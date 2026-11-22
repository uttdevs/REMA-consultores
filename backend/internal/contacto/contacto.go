package contacto

// MÓDULO CONTACTO: POST /api/contacto
// Recibe el formulario "Envíanos un mensaje", lo valida, lo guarda en
// mensajes_contacto y manda el correo de confirmación.

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

// Longitudes máximas = tamaños de las columnas de mensajes_contacto.
const (
	contactoMaxNombre   = 100
	contactoMaxEmpresa  = 120
	contactoMaxTelefono = 20
	contactoMaxCorreo   = 150
	contactoMaxAsunto   = 150
	contactoMaxMensaje  = 3000 // la columna es TEXT; este tope evita abusos
	contactoMaxBody     = 16 << 10
)

var contactoRegexTelefono = regexp.MustCompile(`^[0-9+()\-. ]{7,20}$`)

type contactoEntrada struct {
	Nombre           string `json:"nombre"`
	Empresa          string `json:"empresa"`
	Telefono         string `json:"telefono"`
	Correo           string `json:"correo"`
	Asunto           string `json:"asunto"`
	Mensaje          string `json:"mensaje"`
	AceptaPrivacidad bool   `json:"acepta_privacidad"`
}

// RegistrarRutas registra las rutas del módulo de contacto en el mux HTTP.
func RegistrarRutas(mux *http.ServeMux, db *sql.DB) {
	h := contactoHandler(db)
	mux.HandleFunc("/api/contacto", h)
	mux.HandleFunc("/api/v1/contacto", h)
}

// Registrar es un alias conveniente de RegistrarRutas.
var Registrar = RegistrarRutas

func contactoHandler(db *sql.DB) http.HandlerFunc {
	limitador := &contactoLimitador{
		visitas: map[string][]time.Time{},
		max:     5,
		ventana: 10 * time.Minute,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST, OPTIONS")
			contactoError(w, http.StatusMethodNotAllowed, "Método no permitido.", nil)
			return
		}
		if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
			contactoError(w, http.StatusUnsupportedMediaType, "Envía los datos como JSON.", nil)
			return
		}
		if !limitador.permitir(contactoIP(r)) {
			w.Header().Set("Retry-After", "600")
			contactoError(w, http.StatusTooManyRequests, "Demasiados mensajes. Intenta de nuevo en unos minutos.", nil)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, contactoMaxBody)
		dec := json.NewDecoder(r.Body)
		var in contactoEntrada
		if err := dec.Decode(&in); err != nil {
			var tope *http.MaxBytesError
			if errors.As(err, &tope) {
				contactoError(w, http.StatusRequestEntityTooLarge, "El mensaje es demasiado grande.", nil)
				return
			}
			contactoError(w, http.StatusBadRequest, "JSON inválido.", nil)
			return
		}
		if err := dec.Decode(&struct{}{}); err != io.EOF {
			contactoError(w, http.StatusBadRequest, "JSON inválido.", nil)
			return
		}

		if errs := contactoValidar(&in); len(errs) > 0 {
			contactoError(w, http.StatusUnprocessableEntity, "Revisa los campos marcados.", errs)
			return
		}

		if _, err := contactoGuardar(r.Context(), db, &in); err != nil {
			// El detalle solo va al log, nunca al cliente.
			log.Printf("contacto: error al guardar en BD: %v", err)
			contactoError(w, http.StatusInternalServerError, "No pudimos guardar tu mensaje. Intenta más tarde.", nil)
			return
		}

		go func(nombre, destino, asunto string) {
			if err := contactoEnviarConfirmacion(nombre, destino, asunto); err != nil {
				log.Printf("contacto: no se pudo enviar correo a %s: %v", destino, err)
			}
		}(in.Nombre, in.Correo, in.Asunto)

		contactoResponder(w, http.StatusOK, map[string]any{
			"ok":      true,
			"mensaje": "Mensaje recibido. Te responderemos pronto.",
		})
	}
}

func contactoValidar(in *contactoEntrada) map[string]string {
	errs := map[string]string{}

	in.Nombre = strings.TrimSpace(in.Nombre)
	in.Empresa = strings.TrimSpace(in.Empresa)
	in.Telefono = strings.TrimSpace(in.Telefono)
	in.Correo = strings.TrimSpace(in.Correo)
	in.Asunto = strings.TrimSpace(in.Asunto)
	in.Mensaje = strings.TrimSpace(in.Mensaje)

	contactoLinea(errs, "nombre", in.Nombre, contactoMaxNombre, true)
	contactoLinea(errs, "empresa", in.Empresa, contactoMaxEmpresa, false)

	if in.Telefono != "" {
		contactoLinea(errs, "telefono", in.Telefono, contactoMaxTelefono, false)
		if _, ok := errs["telefono"]; !ok && !contactoRegexTelefono.MatchString(in.Telefono) {
			errs["telefono"] = "Teléfono inválido (ejemplo: +52 55 1234 5678)."
		}
	}

	if in.Correo == "" {
		errs["correo"] = "El correo es obligatorio."
	} else if utf8.RuneCountInString(in.Correo) > contactoMaxCorreo {
		errs["correo"] = "El correo no puede exceder 150 caracteres."
	} else if !contactoCorreoValido(in.Correo) {
		errs["correo"] = "Escribe un correo válido (ejemplo: nombre@empresa.com)."
	}

	contactoLinea(errs, "asunto", in.Asunto, contactoMaxAsunto, true)

	if in.Mensaje == "" {
		errs["mensaje"] = "El mensaje es obligatorio."
	} else if utf8.RuneCountInString(in.Mensaje) > contactoMaxMensaje {
		errs["mensaje"] = "El mensaje no puede exceder 3 000 caracteres."
	} else if contactoTieneControl(in.Mensaje, true) {
		errs["mensaje"] = "El mensaje contiene caracteres no permitidos."
	}

	if !in.AceptaPrivacidad {
		errs["acepta_privacidad"] = "Debes aceptar el aviso de privacidad para continuar."
	}

	return errs
}

func contactoLinea(errs map[string]string, campo, valor string, max int, obligatorio bool) {
	if valor == "" {
		if obligatorio {
			errs[campo] = "Este campo es obligatorio."
		}
		return
	}
	if utf8.RuneCountInString(valor) > max {
		errs[campo] = "No puede exceder los caracteres permitidos."
		return
	}
	if contactoTieneControl(valor, false) {
		errs[campo] = "No se permiten saltos de línea ni caracteres de control."
	}
}

func contactoTieneControl(s string, permitirSaltos bool) bool {
	for _, r := range s {
		if permitirSaltos && (r == '\n' || r == '\r' || r == '\t') {
			continue
		}
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func contactoCorreoValido(c string) bool {
	if strings.ContainsAny(c, " \t\r\n<>\"") || contactoTieneControl(c, false) {
		return false
	}
	a, err := mail.ParseAddress(c)
	if err != nil || a.Address != c {
		return false
	}
	_, dominio, ok := strings.Cut(c, "@")
	return ok && strings.Contains(dominio, ".") && !strings.HasSuffix(dominio, ".")
}

func contactoResponder(w http.ResponseWriter, estado int, cuerpo any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(cuerpo)
}

func contactoError(w http.ResponseWriter, estado int, mensaje string, campos map[string]string) {
	cuerpo := map[string]any{"ok": false, "error": mensaje}
	if len(campos) > 0 {
		cuerpo["campos"] = campos
	}
	contactoResponder(w, estado, cuerpo)
}

func contactoIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// contactoLimitador: máximo N mensajes por IP en una ventana de tiempo.
// Protege contra spam y contra usar el formulario para mandar correos masivos.
type contactoLimitador struct {
	mu      sync.Mutex
	visitas map[string][]time.Time
	max     int
	ventana time.Duration
}

func (l *contactoLimitador) permitir(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	ahora := time.Now()
	corte := ahora.Add(-l.ventana)

	if len(l.visitas) > 10000 { // limpieza para que el mapa no crezca sin fin
		for k, v := range l.visitas {
			if len(v) == 0 || v[len(v)-1].Before(corte) {
				delete(l.visitas, k)
			}
		}
	}

	recientes := l.visitas[ip][:0]
	for _, t := range l.visitas[ip] {
		if t.After(corte) {
			recientes = append(recientes, t)
		}
	}
	if len(recientes) >= l.max {
		l.visitas[ip] = recientes
		return false
	}
	l.visitas[ip] = append(recientes, ahora)
	return true
}
