<template>
  <footer class="footer">
    <div class="container footer-grid">
      <NuxtLink to="/" aria-label="Volver al inicio" class="footer-logo">
        <img src="/images/logo-rema-original.png" alt="REMA Consultores" width="95" height="90">
      </NuxtLink>
      <p>Transformamos el futuro de los negocios a través del conocimiento y la asesoría estratégica.</p>
      <div class="footer-col">
        <strong>Contacto</strong>
        <a href="mailto:contacto@remaconsultores.com">contacto@remaconsultores.com</a>
        <span>229 387 8119 · 558 852 7869</span>
        <span>Veracruz · CDMX · Monterrey</span>
      </div>
      <nav class="footer-col" aria-label="Enlaces del pie">
        <strong>Enlaces</strong>
        <NuxtLink to="/">Inicio</NuxtLink>
        <NuxtLink to="/#servicios">Servicios</NuxtLink>
        <NuxtLink to="/#blog">Blog</NuxtLink>
        <NuxtLink to="/#contacto">Contacto</NuxtLink>
      </nav>
    </div>
    <div class="container footer-bottom">© {{ new Date().getFullYear() }} REMA Consultores. Todos los derechos reservados.</div>
  </footer>
</template>
