# REMA Consultores — Landing Vue 3 + Nuxt

Proyecto convertido desde la landing HTML/CSS con chatbot, a **Vue 3 + Nuxt 4** (la generación vigente de Nuxt y basada en Vue 3).

## Cambios incorporados

- **Menú superior fijo:** permanece visible al recorrer la página. Los enlaces Inicio, Quiénes somos, Servicios, Blog y Contacto llevan a la sección elegida, incluso desde una página de un artículo.
- **Blog en rueda/carrusel:** muestra 3 tarjetas en escritorio, 2 en tableta y 1 en celular; flechas, indicadores y gesto de deslizar en pantallas táctiles. El carrusel gira y puede mostrar todas las publicaciones que se agreguen.
- **Publicaciones por visitantes:** el botón **Publicar un artículo** abre un formulario con nombre, categoría, título, resumen y contenido. Cada publicación se agrega al inicio del carrusel y puede abrirse en una URL propia (`/blog/slug`).
- **Contenido persistente:** se guarda en `data/posts.json` mediante rutas internas de Nuxt `/api/posts` (GET/POST). Otros visitantes que accedan al **mismo servidor** verán las nuevas publicaciones después de actualizar.
- **Chatbot de preguntas frecuentes** con balanza de la justicia, adaptado a Vue.
- **Logo original e imágenes** incluidos en `public/images`. Se mantiene el fondo `hero-integrado.webp`.

## Cómo abrir en Visual Studio Code (Windows)

1. Descomprime el ZIP y abre la carpeta `REMA_Nuxt_Blog_Carrusel` desde **Archivo → Abrir carpeta**.
2. Instala **Node.js 22 LTS o posterior** si aún no lo tienes. Comprueba la versión en la terminal con `node -v`.
3. Abre la terminal integrada en VS Code (`Terminal → Nueva terminal`) y ejecuta:

   ```bash
   npm install
   npm run dev
   ```

4. Abre **http://localhost:3000** en tu navegador (o la URL local indicada en la terminal).
5. Entra a **Blog → Publicar un artículo**, completa el formulario y comprueba que aparece en el carrusel. Recarga la página: el artículo debe seguir visible.
6. Para detener el servidor, presiona `Ctrl + C` en la terminal.

**Importante:** ya no es un proyecto HTML estático. **No uses Live Server** para abrir `index.html`; Nuxt necesita iniciarse con `npm run dev` para que funcionen el blog y su API.

## Compilar para producción

```bash
npm run build
npm run preview
```

Para guardar los artículos de forma estable en un servidor con Node y disco persistente, establece la variable de entorno `REMA_POSTS_FILE` con una ruta escribible que sobreviva a las actualizaciones (ejemplo: `/var/lib/rema/posts.json`). El proceso necesita permisos de lectura/escritura. Si se configura una ruta distinta, copia primero `data/posts.json` a esa ubicación para conservar los artículos de ejemplo.

### Consideraciones antes de publicar la web en Internet

Este paquete es un **prototipo funcional** de publicación comunitaria, pensado para usarse en local o en un servidor Node que tenga almacenamiento persistente. No incluye registro de usuarios ni moderación editorial: cualquier visitante puede publicar contenido (con límites básicos de tamaño y frecuencia). Antes de abrirlo al público se recomienda agregar **autenticación, revisión/moderación, protección antispam más completa y una base de datos compartida**. El archivo JSON es adecuado para pruebas y despliegues sencillos de una sola instancia, no para funciones serverless, múltiples instancias ni tráfico alto. Las imágenes de los artículos son predefinidas por categoría; no hay carga de archivos.

El chatbot sigue siendo de **preguntas frecuentes**, no inteligencia artificial conectada. El formulario de contacto prepara un mensaje en la aplicación de correo (`mailto:`), no lo envía desde el servidor.

## Estructura principal

```text
REMA_Nuxt_Blog_Carrusel/
├── app/
│   ├── app.vue
│   ├── assets/css/main.css
│   ├── components/
│   │   ├── SiteHeader.vue
│   │   ├── SiteFooter.vue
│   │   ├── BlogCarousel.vue
│   │   └── ChatWidget.vue
│   └── pages/
│       ├── index.vue
│       └── blog/[slug].vue
├── server/
│   ├── api/posts/index.get.ts
│   ├── api/posts/index.post.ts
│   ├── api/posts/[slug].get.ts
│   └── utils/posts.ts
├── shared/types/blog.ts
├── data/posts.json
├── public/images/ (logo, hero, iconos y fotos)
├── package.json
├── nuxt.config.ts
└── README.md
```

## Edición rápida

- Colores, sombras y espaciados: `app/assets/css/main.css`
- Texto y servicios: `app/pages/index.vue`
- Carrusel y formulario de publicación: `app/components/BlogCarousel.vue`
- Chatbot de balanza: `app/components/ChatWidget.vue`
- Navegación fija: `app/components/SiteHeader.vue`
- Artículos iniciales: `data/posts.json`
- Validación y persistencia: `server/utils/posts.ts`
