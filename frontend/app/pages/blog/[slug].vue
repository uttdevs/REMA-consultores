<script setup lang="ts">
import type { BlogPost } from '../../../shared/types/blog'
const route = useRoute()
const slug = String(route.params.slug || '')
const { data: post, error } = await useFetch<BlogPost>(`/api/posts/${encodeURIComponent(slug)}`, { key: `post-${slug}` })
useSeoMeta({
  title: () => post.value ? `${post.value.title} | Blog REMA Consultores` : 'Artículo | REMA Consultores',
  description: () => post.value?.summary || 'Artículos de la comunidad REMA Consultores'
})
const displayDate = computed(() => post.value
  ? new Intl.DateTimeFormat('es-MX', { day: 'numeric', month: 'long', year: 'numeric', timeZone: 'UTC' }).format(new Date(post.value.createdAt))
  : '')
const paragraphs = computed(() => (post.value?.content || '').split(/\n\s*\n/).filter(Boolean))
</script>

<template>
  <main class="article-page">
    <div class="container article-container">
      <NuxtLink class="back-link" to="/#blog">← Volver al blog</NuxtLink>
      <div v-if="error || !post" class="article-not-found">
        <h1>Artículo no encontrado</h1>
        <p>Es posible que se haya cambiado la dirección de esta publicación.</p>
        <NuxtLink class="button button-primary" to="/#blog">Ver blog</NuxtLink>
      </div>
      <article v-else>
        <header class="article-header">
          <span class="tag">{{ post.category }}</span>
          <h1>{{ post.title }}</h1>
          <p class="article-meta">Escrito por <strong>{{ post.author }}</strong> · {{ displayDate }}</p>
          <p class="article-summary">{{ post.summary }}</p>
        </header>
        <img class="article-hero" :src="post.image" :alt="`Imagen del artículo: ${post.title}`">
        <div class="article-body">
          <p v-for="(paragraph, i) in paragraphs" :key="i">{{ paragraph }}</p>
        </div>
        <div class="article-end">
          <p>Publicación de la comunidad. La información no sustituye una consulta profesional.</p>
          <NuxtLink class="button button-outline" to="/#blog">← Ver más publicaciones</NuxtLink>
        </div>
      </article>
    </div>
  </main>
</template>
