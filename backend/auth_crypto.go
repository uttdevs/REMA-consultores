package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Configuración recomendada por OWASP para Argon2id.
type argon2Config struct {
	tiempo      uint32 // Iteraciones
	memoria     uint32 // KiB
	hilos       uint8  // Paralelismo
	longitudSal uint32 // Longitud de la sal
	longitudClave uint32 // Longitud del hash resultante
}

var argon2Parametros = argon2Config{
	tiempo:        1,
	memoria:       64 * 1024, // 64 MB
	hilos:         2,
	longitudSal:   16,
	longitudClave: 32,
}

// hashPassword genera un hash seguro de la contraseña usando Argon2id.
// El formato resultante es: $argon2id$v=19$m=65536,t=1,p=2$<sal_base64>$<hash_base64>
func hashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("la contraseña no puede estar vacía")
	}

	sal := make([]byte, argon2Parametros.longitudSal)
	if _, err := rand.Read(sal); err != nil {
		return "", fmt.Errorf("error al generar sal aleatoria: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		sal,
		argon2Parametros.tiempo,
		argon2Parametros.memoria,
		argon2Parametros.hilos,
		argon2Parametros.longitudClave,
	)

	salB64 := base64.RawStdEncoding.EncodeToString(sal)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	formateado := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Parametros.memoria,
		argon2Parametros.tiempo,
		argon2Parametros.hilos,
		salB64,
		hashB64,
	)

	return formateado, nil
}

// verificarPassword valida la contraseña en texto plano contra el hash almacenado en tiempo constante.
func verificarPassword(password, hashAlmacenado string) (bool, error) {
	partes := strings.Split(hashAlmacenado, "$")
	if len(partes) != 6 {
		return false, errors.New("el hash no tiene un formato Argon2 válido")
	}

	tipo := partes[1]
	if tipo != "argon2id" && tipo != "argon2i" {
		return false, fmt.Errorf("variante de argon2 no soportada: %s", tipo)
	}

	var version int
	if _, err := fmt.Sscanf(partes[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("versión inválida en hash: %w", err)
	}
	if version != argon2.Version {
		return false, errors.New("versión incompatible de Argon2")
	}

	var memoria uint32
	var tiempo uint32
	var hilos uint8
	if _, err := fmt.Sscanf(partes[3], "m=%d,t=%d,p=%d", &memoria, &tiempo, &hilos); err != nil {
		return false, fmt.Errorf("parámetros inválidos en hash: %w", err)
	}

	sal, err := base64.RawStdEncoding.DecodeString(partes[4])
	if err != nil {
		return false, fmt.Errorf("sal base64 inválida: %w", err)
	}

	hashEsperado, err := base64.RawStdEncoding.DecodeString(partes[5])
	if err != nil {
		return false, fmt.Errorf("hash base64 inválido: %w", err)
	}

	longitud := uint32(len(hashEsperado))
	var hashGenerado []byte
	if tipo == "argon2id" {
		hashGenerado = argon2.IDKey([]byte(password), sal, tiempo, memoria, hilos, longitud)
	} else {
		hashGenerado = argon2.Key([]byte(password), sal, tiempo, memoria, hilos, longitud)
	}

	if subtle.ConstantTimeCompare(hashEsperado, hashGenerado) == 1 {
		return true, nil
	}
	return false, nil
}

// generarUUID crea un UUID v4 compatible con RFC 4122 usando números aleatorios criptográficos.
func generarUUID() string {
	var u [16]byte
	if _, err := rand.Read(u[:]); err != nil {
		// En el rarísimo caso de fallo de entropía, no panic
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
	}
	u[6] = (u[6] & 0x0f) | 0x40 // Versión 4
	u[8] = (u[8] & 0x3f) | 0x80 // Variante RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

// generarTokenSesion crea un identificador de sesión seguro de 256 bits (64 caracteres hex).
func generarTokenSesion() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
