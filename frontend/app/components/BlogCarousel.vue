<script setup lang="ts">
import type { BlogPost, NewBlogPost } from '../../shared/types/blog'

const { data: posts, pending, error: fetchError, refresh } = await useFetch<BlogPost[]>('/api/posts', { key: 'rema-blog-posts', default: () => [] })
const current = ref(0)
const visibleCount = ref(3)
const modalOpen = ref(false)
const publishing = ref(false)
const publishError = ref('')
const notice = ref('')
const categories = [
  'Finanzas', 'Fiscalidad', 'Derecho corporativo', 'Contabilidad',
  'Inversión extranjera', 'Propiedad industrial', 'Concursos mercantiles',
  'Capacitación empresarial', 'Otros'
]
const blank = (): NewBlogPost => ({ author: '', title: '', category: 'Finanzas', summary: '', content: '', website: '' })
const form = reactive<NewBlogPost>(blank())
const count = computed(() => posts.value?.length || 0)
const shown = computed(() => {
  const list = posts.value || []
  if (!list.length) return []
  const n = Math.min(visibleCount.value, list.length)
  return Array.from({ length: n }, (_, i) => list[(current.value + i) % list.length]!)
})
function updateWidth() {
  visibleCount.value = window.innerWidth < 650 ? 1 : window.innerWidth < 1000 ? 2 : 3
}
onMounted(() => { updateWidth(); window.addEventListener('resize', updateWidth) })
onBeforeUnmount(() => window.removeEventListener('resize', updateWidth))
function rotate(delta: number) {
  if (!count.value) return
  current.value = (current.value + delta + count.value) % count.value
}
let touchStartX = 0
function onTouchStart(event: TouchEvent) { touchStartX = event.changedTouches[0]?.clientX ?? 0 }
function onTouchEnd(event: TouchEvent) {
  const delta = (event.changedTouches[0]?.clientX ?? 0) - touchStartX
  if (Math.abs(delta) > 40) rotate(delta < 0 ? 1 : -1)
}
function openModal() { publishError.value = ''; notice.value = ''; modalOpen.value = true }
function closeModal() { if (!publishing.value) modalOpen.value = false }
function dateText(date: string) {
  return new Intl.DateTimeFormat('es-MX', { day: 'numeric', month: 'short', year: 'numeric', timeZone: 'UTC' }).format(new Date(date))
}
async function publishPost() {
  if (publishing.value) return
  publishing.value = true
  publishError.value = ''
  try {
    const newPost = await $fetch<BlogPost>('/api/posts', { method: 'POST', body: { ...form } })
    await refresh()
    // Garantiza ver el artículo nuevo incluso tras reordenarse.
    const foundIndex = (posts.value || []).findIndex(item => item.id === newPost.id)
    current.value = Math.max(0, foundIndex)
    Object.assign(form, blank())
    modalOpen.value = false
    notice.value = 'Tu artículo ya aparece en el blog. ¡Gracias por compartirlo!'
  } catch (error: any) {
    publishError.value = error?.data?.statusMessage || error?.statusMessage || 'No se pudo publicar. Verifica los datos e inténtalo de nuevo.'
  } finally {
    publishing.value = false
  }
}
</script>

<template>
  <section id="blog" class="section section-blog" aria-labelledby="blog-title">
    <div class="container">
      <div class="blog-topline">
        <div>
          <p class="eyebrow">NUESTRO BLOG</p>
          <h2 id="blog-title">Ideas que <span>hacen crecer</span></h2>
          <p class="blog-intro">Conocimientos, experiencias y artículos compartidos por la comunidad.</p>
        </div>
        <button class="button button-primary blog-create-button" type="button" @click="openModal">
          <span class="plus-icon" aria-hidden="true">＋</span> Publicar un artículo
        </button>
      </div>

      <div class="blog-carousel-toolbar">
        <p class="carousel-caption">Descubre las últimas publicaciones</p>
        <div class="carousel-controls">
          <span v-if="count" class="carousel-count">{{ current + 1 }} / {{ count }}</span>
          <button type="button" aria-label="Ver artículos anteriores" :disabled="count < 2" @click="rotate(-1)">←</button>
          <button type="button" aria-label="Ver artículos siguientes" :disabled="count < 2" @click="rotate(1)">→</button>
        </div>
      </div>

      <p v-if="notice" class="blog-notice" role="status">{{ notice }}</p>
      <p v-if="fetchError" class="blog-error" role="alert">No se pudo cargar el blog. Comprueba que el servidor Nuxt esté iniciado.</p>
      <div v-if="pending" class="blog-loading" role="status">Cargando publicaciones…</div>
      <div v-else-if="!count" class="blog-empty">Todavía no hay publicaciones. ¡Comparte la primera!</div>

      <div v-else class="carousel-stage" tabindex="0" aria-label="Carrusel de artículos, usa las flechas del teclado" @keydown.left.prevent="rotate(-1)" @keydown.right.prevent="rotate(1)" @touchstart.passive="onTouchStart" @touchend.passive="onTouchEnd">
        <div class="blog-carousel-grid" :style="{ '--visible': shown.length }" aria-live="polite">
          <article v-for="post in shown" :key="post.id" class="blog-card">
            <NuxtLink :to="`/blog/${post.slug}`" class="blog-cover" :aria-label="`Leer ${post.title}`">
              <img :src="post.image" :alt="`Imagen de ${post.category}`" loading="lazy">
            </NuxtLink>
            <div class="blog-copy">
              <span class="tag">{{ post.category }}</span>
              <p class="blog-byline">Por {{ post.author }} · {{ dateText(post.createdAt) }}</p>
              <h3><NuxtLink :to="`/blog/${post.slug}`">{{ post.title }}</NuxtLink></h3>
              <p class="blog-summary">{{ post.summary }}</p>
              <NuxtLink class="blog-read-more" :to="`/blog/${post.slug}`">Leer artículo <span aria-hidden="true">→</span></NuxtLink>
            </div>
          </article>
        </div>
      </div>

      <div v-if="count > 1" class="carousel-dots" aria-label="Seleccionar artículo inicial">
        <button v-for="(post, index) in posts" :key="post.id" :class="{ active: current === index }" type="button" :aria-label="`Mostrar desde el artículo ${index + 1}`" :aria-current="current === index ? 'true' : undefined" @click="current = index" />
      </div>
      <p class="blog-footnote">Los artículos de la comunidad son informativos y no sustituyen asesoría profesional.</p>
    </div>

    <Teleport to="body">
      <div v-if="modalOpen" class="modal-backdrop" @click.self="closeModal" @keydown.esc="closeModal">
        <section class="blog-modal" role="dialog" aria-modal="true" aria-labelledby="new-post-title" tabindex="-1">
          <div class="modal-head">
            <div><p class="eyebrow">COMPARTE TUS IDEAS</p><h2 id="new-post-title">Publica tu <span>artículo</span></h2></div>
            <button type="button" class="modal-close" :disabled="publishing" aria-label="Cerrar formulario" @click="closeModal">×</button>
          </div>
          <p class="modal-description">Tu publicación aparecerá en la rueda del blog y tendrá su propia página. No incluyas datos personales sensibles.</p>
          <form class="publish-form" @submit.prevent="publishPost">
            <div class="publish-row">
              <label>Tu nombre <span>*</span><input v-model.trim="form.author" required minlength="2" maxlength="60" placeholder="Nombre del autor" autocomplete="name"></label>
              <label>Categoría <span>*</span><select v-model="form.category" required><option v-for="category in categories" :key="category" :value="category">{{ category }}</option></select></label>
            </div>
            <label>Título <span>*</span><input v-model.trim="form.title" required minlength="8" maxlength="110" placeholder="¿De qué trata tu artículo?"></label>
            <label>Resumen <span>*</span><textarea v-model.trim="form.summary" required minlength="20" maxlength="260" rows="2" placeholder="Explica brevemente el tema (20 a 260 caracteres)"></textarea></label>
            <label>Contenido completo <span>*</span><textarea v-model.trim="form.content" required minlength="80" maxlength="10000" rows="8" placeholder="Escribe tu artículo, separando párrafos con una línea en blanco…"></textarea></label>
            <input v-model="form.website" type="text" name="website" class="sr-only" tabindex="-1" autocomplete="off" aria-hidden="true">
            <p v-if="publishError" class="form-error" role="alert">{{ publishError }}</p>
            <div class="publish-actions">
              <button class="button button-quiet" type="button" :disabled="publishing" @click="closeModal">Cancelar</button>
              <button class="button button-primary" type="submit" :disabled="publishing">{{ publishing ? 'Publicando…' : 'Publicar artículo →' }}</button>
            </div>
          </form>
        </section>
      </div>
    </Teleport>
  </section>
</template>
