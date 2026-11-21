# Contrato de API

## Convenciones generales

- Base URL: `/api/v1`
- Autenticación: cookie de sesión `HttpOnly`, `Secure`, enviada automáticamente por el navegador
- Formato: JSON en request y response
- Errores, formato estándar:

```json
{ "error": { "code": "string", "message": "string" } }
```

- Roles: `admin`, `editor`, `empleado`

## 1. Autenticación y roles (RF-04)

### POST /auth/login

- Rol requerido: ninguno (público)
- Request:

```json
{ "email": "string", "password": "string" }
```

- Response 200:

```json
{ "user": { "id": "uuid", "nombre": "string", "email": "string", "rol": "admin|editor|empleado" } }
```

- Response 401: credenciales inválidas
- Reglas:
  - Límite de intentos: 5 fallidos → bloqueo temporal de 15 min por IP+email
  - La cookie de sesión se emite en el header `Set-Cookie`, expira a las 8h de inactividad

### POST /auth/logout

- Rol requerido: cualquier usuario autenticado
- Response 204

### GET /auth/me

- Rol requerido: cualquier usuario autenticado
- Response 200: mismo objeto `user` que login
- Uso: frontend lo llama al cargar el panel para saber si hay sesión activa y qué rol tiene

### Gestión de usuarios (solo admin)

#### GET /usuarios

- Rol: admin
- Response 200: lista paginada `{ "items": [...], "total": n, "page": n }`
- Cada item: `{ id, nombre, email, rol, activo, creado_en }`

#### POST /usuarios

- Rol: admin
- Request: `{ nombre, email, rol, password_temporal }`
- Response 201: usuario creado
- Regla: la contraseña se hashea con argon2 antes de guardar; nunca se devuelve en ninguna respuesta

#### PATCH /usuarios/:id

- Rol: admin
- Request: campos a modificar (`rol`, `activo`, `nombre`)
- Response 200

#### DELETE /usuarios/:id

- Rol: admin
- Response 204
- Regla: no permite que un admin se auto-elimine

## 2. Blog / Artículos (RF-05, RF-06, RF-07)

### Modelo de artículo

```json
{
  "id": "uuid",
  "titulo": "string",
  "slug": "string",
  "cuerpo": "string (HTML sanitizado)",
  "extracto": "string",
  "imagen_portada": "url | null",
  "meta_titulo": "string",
  "meta_descripcion": "string",
  "estado": "borrador | programado | publicado",
  "fecha_publicacion": "ISO8601 | null",
  "autor_id": "uuid",
  "creado_en": "ISO8601",
  "actualizado_en": "ISO8601"
}
```

### GET /articulos

- Rol: editor, admin
- Query params: `estado`, `autor_id`, `page`, `q` (búsqueda por título)
- Response 200: lista paginada de artículos (modelo completo)

### GET /articulos/:id

- Rol: editor, admin
- Response 200: artículo completo

### POST /articulos

- Rol: editor, admin
- Request: `{ titulo, cuerpo, extracto, imagen_portada, meta_titulo, meta_descripcion, estado, fecha_publicacion }`
- Response 201: artículo creado
- Reglas:
  - `slug` se autogenera desde `titulo` si no se envía; debe ser único (si choca, se agrega un sufijo numérico)
  - `cuerpo` se sanitiza en el backend con `bluemonday` antes de guardar
  - Si `estado = programado`, `fecha_publicacion` es obligatoria y debe ser futura

### PATCH /articulos/:id

- Rol: editor (solo sus propios artículos) o admin (cualquiera)
- Request: campos a modificar
- Response 200

### DELETE /articulos/:id

- Rol: admin
- Response 204

### POST /articulos/importar-docx

- Rol: editor, admin
- Request: `multipart/form-data` con el archivo `.docx`
- Response 200:

```json
{ "cuerpo_html": "string", "imagenes_subidas": ["url", "url"] }
```

- Reglas:
  - Conversión con Mammoth o Pandoc
  - Las imágenes embebidas se extraen y se suben al bucket; las URLs resultantes se insertan en el HTML
  - El resultado se devuelve para edición/revisión en el editor — **no se guarda automáticamente como artículo**

### POST /articulos/:id/imagen-portada

- Rol: editor, admin
- Request: referencia a un archivo ya subido vía el flujo de almacenamiento (ver sección 3)
- Response 200: `{ "imagen_portada": "url" }`

## 3. Almacenamiento privado (RF-08, RF-09, RF-10)

### Modelo de archivo/carpeta

```json
{
  "id": "uuid",
  "nombre": "string",
  "tipo": "archivo | carpeta",
  "carpeta_padre_id": "uuid | null",
  "tamano_bytes": "number | null",
  "mime_type": "string | null",
  "subido_por": "uuid",
  "creado_en": "ISO8601"
}
```

### GET /storage/carpetas/:id/contenido

- Rol: cualquier usuario autenticado (RF-09)
- Response 200: `{ "carpetas": [...], "archivos": [...] }` dentro de esa carpeta
- `:id = raiz` para el nivel superior

### POST /storage/carpetas

- Rol: cualquier usuario autenticado
- Request: `{ nombre, carpeta_padre_id }`
- Response 201

### POST /storage/archivos/solicitar-subida

- Rol: cualquier usuario autenticado
- Request: `{ nombre, carpeta_padre_id, mime_type, tamano_bytes }`
- Response 200:

```json
{ "url_subida": "string (prefirmada, corta duración)", "archivo_id": "uuid" }
```

- Regla: el frontend sube el archivo **directo al bucket** usando esa URL; no pasa por la API Go

### POST /storage/archivos/:id/confirmar

- Rol: cualquier usuario autenticado
- Response 200
- Uso: el frontend confirma que la subida a S3 terminó; el backend marca el archivo como disponible y registra la actividad (RF-10)

### GET /storage/archivos/:id/descargar

- Rol: cualquier usuario autenticado
- Response 200: `{ "url_descarga": "string (prefirmada, corta duración)" }`
- Regla: se registra en el log de actividad (RF-10)

### PATCH /storage/archivos/:id — renombrar o mover

- Request: `{ nombre?, carpeta_padre_id? }`
- Response 200

### DELETE /storage/archivos/:id

### DELETE /storage/carpetas/:id

- Response 204
- Regla: eliminar una carpeta requiere que esté vacía, o pasar `?recursivo=true`

### GET /storage/actividad

- Rol: admin
- Query: `usuario_id`, `tipo` (subida/descarga/eliminacion), `page`
- Response 200: lista paginada `{ usuario_id, accion, archivo_id, fecha }`

## 4. Contacto y Agendamiento de Citas (RF-01, RF-02)

Este módulo unifica todas las vías por las cuales el público y los clientes se comunican con la consultoría desde el sitio institucional:
1. **Formulario de contacto general:** Envío de consultas y mensajes asíncronos sin fecha fija.
2. **Agendamiento de citas:** Reserva formal de asesorías en fecha y hora específica, en modalidad presencial (oficina) o en línea. **No requiere cuenta de usuario previa**: se agenda directamente con nombre, correo y teléfono.

---

### 4.1 Formulario de contacto general (RF-02)

Recepción y validación de mensajes del formulario "Envíanos un mensaje". Se almacenan en la tabla `mensajes_contacto` y disparan un acuse de recibo automático por correo electrónico.

#### Modelo de mensaje de contacto

```json
{
  "id": 1,
  "nombre": "string",
  "empresa": "string | null",
  "telefono": "string | null",
  "correo": "string",
  "asunto": "string",
  "mensaje": "string",
  "acepta_privacidad": true,
  "acepta_privacidad_en": "ISO8601 | null",
  "creado_en": "ISO8601"
}
```

#### POST /api/contacto (alias `/api/v1/contacto`)

- Rol requerido: ninguno (público)
- Headers: `Content-Type: application/json`
- Tamaño máximo del body: 16 KB (`16384 bytes`)
- Límite de tasa (Rate Limiting / Antispam):
  - Máximo **5 mensajes por IP** en una ventana de **10 minutos**.
  - Si se supera el límite: HTTP 429 con cabecera `Retry-After: 600`.
- Request body:

```json
{
  "nombre": "string",
  "empresa": "string (opcional)",
  "telefono": "string (opcional)",
  "correo": "string",
  "asunto": "string",
  "mensaje": "string",
  "acepta_privacidad": true
}
```

- Reglas de validación:
  - Se eliminan espacios en blanco al inicio y final (`trim`) de todas las cadenas de texto.
  - `nombre`: obligatorio, máximo 100 caracteres, sin caracteres de control (`unicode.IsControl`), UTF-8 válido.
  - `empresa`: opcional, máximo 120 caracteres, sin caracteres de control. Si no se envía o está en blanco, se persiste como `NULL`.
  - `telefono`: opcional. Si se envía, debe tener entre 7 y 20 caracteres y cumplir con la expresión regular `^[0-9+()\-. ]{7,20}$` (dígitos y símbolos `+`, `(`, `)`, `-`, `.`, espacios). Si no se envía o está en blanco, se persiste como `NULL`.
  - `correo`: obligatorio, máximo 150 caracteres, formato RFC 5322 válido (`mail.ParseAddress`), con `@` y dominio con punto (sin caracteres de control ni espacios).
  - `asunto`: obligatorio, máximo 150 caracteres, sin caracteres de control.
  - `mensaje`: obligatorio, máximo 3000 caracteres (columna `TEXT`). Permite saltos de línea y tabuladores (`\n`, `\r`, `\t`), pero ningún otro carácter de control.
  - `acepta_privacidad`: booleano (`true` | `false`). Si es `true`, en BD se guarda `acepta_privacidad = 1` y se estampa la fecha/hora en `acepta_privacidad_en = NOW()`.

- Response 201 Created:

```json
{
  "ok": true,
  "mensaje": "¡Gracias! Recibimos tu mensaje y te enviamos un correo de confirmación."
}
```

- Errores posibles:
  - **400 Bad Request**: JSON inválido o mal formado.
    ```json
    { "ok": false, "error": "JSON inválido." }
    ```
  - **405 Method Not Allowed**: Método distinto a POST.
    - Cabecera: `Allow: POST, OPTIONS`
    ```json
    { "ok": false, "error": "Método no permitido." }
    ```
  - **413 Request Entity Too Large**: El cuerpo excede los 16 KB.
    ```json
    { "ok": false, "error": "El mensaje es demasiado grande." }
    ```
  - **415 Unsupported Media Type**: No se envió cabecera `Content-Type: application/json`.
    ```json
    { "ok": false, "error": "Envía los datos como JSON." }
    ```
  - **422 Unprocessable Entity**: Errores de validación en campos individuales.
    ```json
    {
      "ok": false,
      "error": "Revisa los campos marcados.",
      "campos": {
        "nombre": "Este campo es obligatorio.",
        "correo": "Correo no válido.",
        "telefono": "Teléfono no válido (7 a 20 dígitos; puedes usar + ( ) - .)."
      }
    }
    ```
  - **429 Too Many Requests**: Límite de 5 envíos por IP en 10 minutos excedido.
    - Cabecera: `Retry-After: 600`
    ```json
    { "ok": false, "error": "Demasiados mensajes. Intenta de nuevo en unos minutos." }
    ```
  - **500 Internal Server Error**: Error en el servidor o la base de datos al guardar el mensaje.
    ```json
    { "ok": false, "error": "No pudimos guardar tu mensaje. Intenta más tarde." }
    ```

- Efectos secundarios y persistencia:
  - **Base de datos (MySQL/MariaDB)**: Inserción directa y parametrizada en la tabla `mensajes_contacto`.
  - **Confirmación por correo (asíncrona)**: Si las credenciales SMTP (`SMTP_HOST`, `SMTP_USER`, etc.) están configuradas en `.env`, se despacha en segundo plano (`goroutine`) un correo de confirmación al remitente mediante STARTTLS (puerto 587) o TLS (puerto 465) codificado en Base64. Si SMTP falla o no está configurado, el mensaje queda almacenado y la respuesta HTTP 201 no se interrumpe.

---

### 4.2 Agendamiento de citas y asesorías

Módulo para reserva de citas con disponibilidad en tiempo real y gestión desde el panel de administración.

#### Modelo de cita

```json
{
  "id": "uuid",
  "usuario_id": "uuid | null",
  "nombre": "string",
  "correo": "string",
  "telefono": "string | null",
  "servicio_id": "uuid",
  "oficina_id": "uuid | null",
  "modalidad": "presencial | en_linea",
  "fecha_hora": "ISO8601",
  "duracion_minutos": "number",
  "estado": "pendiente | confirmada | cancelada | completada",
  "notas": "string | null"
}
```

#### GET /servicios

- Rol: público
- Response 200: lista de servicios activos (para mostrar en landing y selector de reserva)

#### GET /oficinas

- Rol: público
- Response 200: lista de oficinas activas (para elegir sede si la modalidad es presencial)

#### GET /citas/disponibilidad

- Rol: público
- Query: `servicio_id`, `oficina_id` (si es presencial), `fecha`
- Response 200: horarios libres ese día para ese servicio/oficina
- Regla: calculado a partir de `duracion_minutos` del servicio y las citas ya ocupadas

#### POST /citas

- Rol: público (sin cuenta) o usuario autenticado
- Request: `{ nombre, correo, telefono, servicio_id, oficina_id, modalidad, fecha_hora }`
- Response 201: cita creada con `estado = pendiente`
- Reglas:
  - Si `modalidad = presencial`, `oficina_id` es obligatorio
  - `duracion_minutos` se toma del servicio, no lo envía el cliente
  - Se valida que el horario siga disponible antes de confirmar (evita doble reserva)

#### GET /citas

- Rol: admin, editor
- Query: `estado`, `fecha`, `servicio_id`, `oficina_id`, `page`
- Response 200: lista paginada de citas

#### PATCH /citas/:id

- Rol: admin, editor
- Request: `{ estado, notas? }`
- Response 200
- Uso: confirmar, cancelar o marcar como completada una cita

## Notas para ambos equipos

- Frontend puede empezar a maquetar pantallas contra estos contratos de inmediato, usando datos simulados (mock) con esta misma forma de JSON — no necesita esperar a que backend tenga nada implementado.
- Cualquier cambio a un contrato debe discutirse entre ambos equipos antes de implementarlo: mientras no exista una spec OpenAPI generada, este documento es la fuente de verdad.
- Pendiente de decidir: formalizar esto en OpenAPI para generar un mock server y tipos TypeScript automáticamente, y así evitar que el documento se desincronice del código real.

## Nota sobre roles (actualizado)

`usuarios.rol` es solo **admin** y **editor** — las dos personas que entran al panel. No hay rol "cliente": quien agenda una cita lo hace sin necesidad de cuenta (ver sección 4). No hay registro público; las cuentas las crea un admin (ver sección 1).
