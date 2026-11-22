package contacto

// MÓDULO CONTACTO: guardado en la tabla mensajes_contacto.
//
// Anti inyección SQL: la consulta es FIJA y los datos del usuario viajan solo
// como parámetros (?). El driver los envía por separado (prepared statement),
// así que nunca se interpretan como SQL, sin importar lo que escriban.

import (
	"context"
	"database/sql"
	"time"
)

const contactoSQLInsertar = `
INSERT INTO mensajes_contacto
  (nombre, empresa, telefono, correo, asunto, mensaje, acepta_privacidad, acepta_privacidad_en)
VALUES
  (?, ?, ?, ?, ?, ?, ?, IF(? = 1, NOW(), NULL))`

func contactoGuardar(ctx context.Context, db *sql.DB, in *contactoEntrada) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	acepta := 0
	if in.AceptaPrivacidad {
		acepta = 1
	}

	res, err := db.ExecContext(ctx, contactoSQLInsertar,
		in.Nombre,
		contactoNulo(in.Empresa),
		contactoNulo(in.Telefono),
		in.Correo,
		in.Asunto,
		in.Mensaje,
		acepta,
		acepta,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// contactoNulo guarda NULL (y no cadena vacía) en los campos opcionales.
func contactoNulo(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
