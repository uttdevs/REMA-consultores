package contacto

// MÓDULO CONTACTO: correo de confirmación por SMTP (solo librería estándar).
// Puerto 587 usa STARTTLS; puerto 465 usa TLS directo. Nunca se envía en claro.

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"rema-backend/internal/config"
)

type correoConfig struct {
	host, puerto, usuario, clave, de, nombreDe string
}

func correoLeerConfig() (correoConfig, bool) {
	c := correoConfig{
		host:     config.Get("SMTP_HOST", ""),
		puerto:   config.Get("SMTP_PORT", "587"),
		usuario:  config.Get("SMTP_USER", ""),
		clave:    config.Get("SMTP_PASS", ""),
		nombreDe: config.Get("SMTP_FROM_NAME", "REMA Consultores"),
	}
	c.de = config.Get("SMTP_FROM", c.usuario)
	return c, c.host != "" && c.usuario != "" && c.clave != "" && c.de != ""
}

// contactoEnviarConfirmacion manda el acuse al correo que escribió la persona.
// Los datos del usuario (nombre, asunto) solo van en el CUERPO, que se codifica
// en base64; las cabeceras (To, From, Subject) no llevan texto del usuario, así
// que no hay inyección de cabeceras. El destinatario ya fue validado.
func contactoEnviarConfirmacion(nombre, destino, asunto string) error {
	cfg, ok := correoLeerConfig()
	if !ok {
		log.Print("contacto: SMTP no configurado (SMTP_HOST/USER/PASS); se omite el correo de confirmación")
		return nil
	}
	msg, err := correoArmarMensaje(cfg, nombre, destino, asunto)
	if err != nil {
		return err
	}
	return correoEnviarSMTP(cfg, destino, msg)
}

func correoArmarMensaje(cfg correoConfig, nombre, destino, asunto string) ([]byte, error) {
	cuerpo := "Hola " + nombre + ",\r\n\r\n" +
		"Gracias por escribirnos. Recibimos tu mensaje con el asunto «" + asunto + "» " +
		"y uno de nuestros especialistas se pondrá en contacto contigo lo antes posible.\r\n\r\n" +
		"Si necesitas atención inmediata, llámanos al 229 387 8119 o al 558 852 7869.\r\n\r\n" +
		"REMA Consultores\r\ncontacto@remaconsultores.com\r\n\r\n" +
		"Este es un mensaje automático de confirmación.\r\n"

	idAleatorio := make([]byte, 12)
	if _, err := rand.Read(idAleatorio); err != nil {
		return nil, err
	}
	_, dominio, _ := strings.Cut(cfg.de, "@")

	cabeceras := []string{
		"From: " + (&mail.Address{Name: cfg.nombreDe, Address: cfg.de}).String(),
		"To: " + (&mail.Address{Address: destino}).String(),
		"Subject: " + mime.QEncoding.Encode("UTF-8", "Recibimos tu mensaje - REMA Consultores"),
		"Date: " + time.Now().Format(time.RFC1123Z),
		"Message-ID: <" + hex.EncodeToString(idAleatorio) + "@" + dominio + ">",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: base64",
	}
	return []byte(strings.Join(cabeceras, "\r\n") + "\r\n\r\n" + correoBase64(cuerpo)), nil
}

// correoBase64 codifica en base64 con líneas de 76 caracteres (RFC 2045).
func correoBase64(s string) string {
	b := base64.StdEncoding.EncodeToString([]byte(s))
	var sb strings.Builder
	for len(b) > 76 {
		sb.WriteString(b[:76] + "\r\n")
		b = b[76:]
	}
	sb.WriteString(b + "\r\n")
	return sb.String()
}

func correoEnviarSMTP(cfg correoConfig, para string, msg []byte) error {
	dir := net.JoinHostPort(cfg.host, cfg.puerto)
	d := net.Dialer{Timeout: 10 * time.Second}
	tlsCfg := &tls.Config{ServerName: cfg.host, MinVersion: tls.VersionTLS12}

	var conn net.Conn
	var err error
	if cfg.puerto == "465" {
		conn, err = tls.DialWithDialer(&d, "tcp", dir, tlsCfg)
	} else {
		conn, err = d.Dial("tcp", dir)
	}
	if err != nil {
		return err
	}
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	c, err := smtp.NewClient(conn, cfg.host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()

	if cfg.puerto != "465" {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return errors.New("el servidor SMTP no ofrece STARTTLS; no se envía en claro")
		}
		if err := c.StartTLS(tlsCfg); err != nil {
			return err
		}
	}
	if err := c.Auth(smtp.PlainAuth("", cfg.usuario, cfg.clave, cfg.host)); err != nil {
		return fmt.Errorf("autenticación SMTP: %w", err)
	}
	if err := c.Mail(cfg.de); err != nil {
		return err
	}
	if err := c.Rcpt(para); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}
