<script setup lang="ts">
const route = useRoute()
const menuOpen = ref(false)
watch(() => route.fullPath, () => { menuOpen.value = false })
const links = [
  { to: '/#nosotros', label: 'Quiénes somos' },
  { to: '/#servicios', label: 'Servicios' },
  { to: '/#blog', label: 'Blog' },
  { to: '/#contacto', label: 'Contacto' }
]
</script>

<template>
  <header class="site-header" id="inicio">
    <div class="container header-content">
      <NuxtLink to="/" class="logo-link" aria-label="REMA Consultores, inicio">
        <img src="/images/logo-rema-original.png" alt="REMA Consultores" width="112" height="92">
      </NuxtLink>
      <nav class="nav-desktop" aria-label="Navegación principal">
        <NuxtLink v-for="item in links" :key="item.to" :to="item.to">{{ item.label }}</NuxtLink>
      </nav>
      <div class="mobile-nav-wrap">
        <button class="nav-toggle" type="button" :aria-expanded="menuOpen" aria-controls="mobile-navigation" @click="menuOpen = !menuOpen">
          <span aria-hidden="true">☰</span> Menú
        </button>
        <nav v-if="menuOpen" id="mobile-navigation" class="nav-mobile" aria-label="Navegación móvil">
          <NuxtLink v-for="item in links" :key="item.to" :to="item.to" @click="menuOpen = false">{{ item.label }}</NuxtLink>
        </nav>
      </div>
    </div>
  </header>
</template>
