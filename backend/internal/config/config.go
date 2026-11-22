package config

// BASE COMPARTIDA: carga de variables de entorno desde .env y lectura con valores por defecto.

import (
	"bufio"
	"os"
	"strings"
)

// Cargar lee uno o varios archivos .env (CLAVE=valor). Las variables que ya
// existen en el entorno no se sobreescriben. Si el archivo no existe, se ignora.
// Si no se especifican rutas, busca por defecto en ".env" y "../.env".
func Cargar(rutas ...string) {
	if len(rutas) == 0 {
		rutas = []string{".env", "../.env"}
	}
	for _, ruta := range rutas {
		f, err := os.Open(ruta)
		if err != nil {
			continue
		}
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
			valor = strings.TrimSpace(valor)
			if len(valor) >= 2 {
				if (valor[0] == '"' && valor[len(valor)-1] == '"') ||
					(valor[0] == '\'' && valor[len(valor)-1] == '\'') {
					valor = valor[1 : len(valor)-1]
				}
			}
			if _, yaExiste := os.LookupEnv(clave); !yaExiste {
				_ = os.Setenv(clave, valor)
			}
		}
		_ = f.Close()
	}
}

// Get obtiene el valor de una variable de entorno o retorna defecto si está vacía.
func Get(clave, defecto string) string {
	if v := strings.TrimSpace(os.Getenv(clave)); v != "" {
		return v
	}
	return defecto
}
