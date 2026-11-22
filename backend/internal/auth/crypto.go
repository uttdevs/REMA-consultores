package auth

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
	tiempo        uint32 // Iteraciones
	memoria       uint32 // KiB
	hilos         uint8  // Paralelismo
	longitudSal   uint32 // Longitud de la sal
	longitudClave uint32 // Longitud del hash resultante
}

var argon2Parametros = argon2Config{
	tiempo:        1,
	memoria:       64 * 1024, // 64 MB
	hilos:         4,
	longitudSal:   16,
	longitudClave: 32,
}

var (
	errHashInvalido           = errors.New("formato de hash argon2id inválido")
	errVersionIncompatible    = errors.New("versión incompatible de argon2")
	errParametrosIncompatibles = errors.New("parámetros de hash no coinciden")
)

// HashPassword genera un hash Argon2id a partir de una contraseña en texto plano.
// Formato generado: $argon2id$v=19$m=65536,t=1,p=4$<sal_b64>$<hash_b64>
func HashPassword(password string) (string, error) {
	sal := make([]byte, argon2Parametros.longitudSal)
	if _, err := rand.Read(sal); err != nil {
		return "", fmt.Errorf("error al generar sal criptográfica: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		sal,
		argon2Parametros.tiempo,
		argon2Parametros.memoria,
		argon2Parametros.hilos,
		argon2Parametros.longitudClave,
	)

	b64Sal := base64.RawStdEncoding.EncodeToString(sal)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	formato := "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
	resultado := fmt.Sprintf(
		formato,
		argon2.Version,
		argon2Parametros.memoria,
		argon2Parametros.tiempo,
		argon2Parametros.hilos,
		b64Sal,
		b64Hash,
	)

	return resultado, nil
}

var hashPassword = HashPassword

// VerificarPassword compara de forma constante en tiempo una contraseña plana contra el hash almacenado.
func VerificarPassword(password, hashAlmacenado string) (bool, error) {
	partes := strings.Split(hashAlmacenado, "$")
	if len(partes) != 6 {
		return false, errHashInvalido
	}

	if partes[1] != "argon2id" {
		return false, errHashInvalido
	}

	var version int
	_, err := fmt.Sscanf(partes[2], "v=%d", &version)
	if err != nil || version != argon2.Version {
		return false, errVersionIncompatible
	}

	var memoria, tiempo uint32
	var hilos uint8
	_, err = fmt.Sscanf(partes[3], "m=%d,t=%d,p=%d", &memoria, &tiempo, &hilos)
	if err != nil {
		return false, errParametrosIncompatibles
	}

	sal, err := base64.RawStdEncoding.DecodeString(partes[4])
	if err != nil {
		return false, errHashInvalido
	}

	hashEsperado, err := base64.RawStdEncoding.DecodeString(partes[5])
	if err != nil {
		return false, errHashInvalido
	}

	hashCalculado := argon2.IDKey(
		[]byte(password),
		sal,
		tiempo,
		memoria,
		hilos,
		uint32(len(hashEsperado)),
	)

	if subtle.ConstantTimeCompare(hashEsperado, hashCalculado) == 1 {
		return true, nil
	}

	return false, nil
}

var verificarPassword = VerificarPassword

// GenerarUUID produce un identificador UUID v4 estándar usando números aleatorios criptográficos.
func GenerarUUID() string {
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		panic(fmt.Sprintf("fallo crítico en rand.Read al generar UUID: %v", err))
	}

	// RFC 4122 v4
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // versión 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // variante RFC 4122

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:16],
	)
}

var generarUUID = GenerarUUID

// GenerarTokenSesion genera un token seguro de 32 bytes aleatorios codificado en hexadecimal (64 caracteres).
func GenerarTokenSesion() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("fallo al generar token de sesión: %v", err))
	}
	return hex.EncodeToString(b)
}

var generarTokenSesion = GenerarTokenSesion
