package main

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

// registrarContacto es la ÚNICA función que el main.go común necesita llamar.
func registrarContacto(mux *http.ServeMux, db *sql.DB) {
	h := contactoHandler(db)
	mux.HandleFunc("/api/contacto", h)
	mux.HandleFunc("/api/v1/contacto", h)
}

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
			log.Printf("contacto: error al guardar: %v", err)
			contactoError(w, http.StatusInternalServerError, "No pudimos guardar tu mensaje. Intenta más tarde.", nil)
			return
		}

		// El correo se manda aparte: si falla, el mensaje ya quedó guardado.
		go func(nombre, correo, asunto string) {
			if err := contactoEnviarConfirmacion(nombre, correo, asunto); err != nil {
				log.Printf("contacto: correo de confirmación no enviado: %v", err)
			}
		}(in.Nombre, in.Correo, in.Asunto)

		contactoResponder(w, http.StatusCreated, map[string]any{
			"ok":      true,
			"mensaje": "¡Gracias! Recibimos tu mensaje y te enviamos un correo de confirmación.",
		})
	}
}

// contactoValidar limpia (trim) y valida. Devuelve campo -> error.
func contactoValidar(in *contactoEntrada) map[string]string {
	in.Nombre = strings.TrimSpace(in.Nombre)
	in.Empresa = strings.TrimSpace(in.Empresa)
	in.Telefono = strings.TrimSpace(in.Telefono)
	in.Correo = strings.TrimSpace(in.Correo)
	in.Asunto = strings.TrimSpace(in.Asunto)
	in.Mensaje = strings.TrimSpace(in.Mensaje)

	errs := map[string]string{}

	contactoLinea(errs, "nombre", in.Nombre, contactoMaxNombre, true)
	contactoLinea(errs, "empresa", in.Empresa, contactoMaxEmpresa, false)
	contactoLinea(errs, "asunto", in.Asunto, contactoMaxAsunto, true)

	if in.Telefono != "" && !contactoRegexTelefono.MatchString(in.Telefono) {
		errs["telefono"] = "Teléfono no válido (7 a 20 dígitos; puedes usar + ( ) - .)."
	}

	switch {
	case in.Correo == "":
		errs["correo"] = "Este campo es obligatorio."
	case utf8.RuneCountInString(in.Correo) > contactoMaxCorreo:
		errs["correo"] = "Máximo 150 caracteres."
	case !contactoCorreoValido(in.Correo):
		errs["correo"] = "Correo no válido."
	}

	switch {
	case in.Mensaje == "":
		errs["mensaje"] = "Este campo es obligatorio."
	case !utf8.ValidString(in.Mensaje) || contactoTieneControl(in.Mensaje, true):
		errs["mensaje"] = "El mensaje contiene caracteres no permitidos."
	case utf8.RuneCountInString(in.Mensaje) > contactoMaxMensaje:
		errs["mensaje"] = "Máximo 3000 caracteres."
	}
	return errs
}

func contactoLinea(errs map[string]string, campo, valor string, max int, obligatorio bool) {
	switch {
	case valor == "":
		if obligatorio {
			errs[campo] = "Este campo es obligatorio."
		}
	case !utf8.ValidString(valor) || contactoTieneControl(valor, false):
		errs[campo] = "Contiene caracteres no permitidos."
	case utf8.RuneCountInString(valor) > max:
		errs[campo] = "Demasiado largo."
	}
}

// contactoTieneControl detecta caracteres de control (NUL, saltos de línea, etc.).
// Si permitirSaltos es true, acepta \n, \r y \t (para el campo mensaje).
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
	json.NewEncoder(w).Encode(cuerpo)
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