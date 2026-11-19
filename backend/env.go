package main

// BASE COMPARTIDA: carga de variables de entorno desde .env (sin dependencias).

import (
	"bufio"
	"os"
	"strings"
)

// cargarEnv lee un archivo .env (CLAVE=valor). Las variables que ya existen en
// el entorno no se sobreescriben. Si el archivo no existe, no pasa nada.
func cargarEnv(ruta string) {
	f, err := os.Open(ruta)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		linea := strings.TrimSpace(sc.Text())
		if linea == "" || strings.HasPrefix(linea, "#") {
			continue
		}
		clave, valor, ok := strings.Cut(linea, "=")
		if !ok {
			continue
		}
		clave = strings.TrimSpace(clave)
		valor = strings.Trim(strings.TrimSpace(valor), `"'`)
		if _, existe := os.LookupEnv(clave); !existe {
			os.Setenv(clave, valor)
		}
	}
}

// env devuelve la variable o un valor por defecto si está vacía.
func env(clave, defecto string) string {
	if v := strings.TrimSpace(os.Getenv(clave)); v != "" {
		return v
	}
	return defecto
}