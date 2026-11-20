<script setup lang="ts">
useSeoMeta({ title: 'REMA Consultores | Asesoría integral para tu negocio' })

const services = [
  { name: 'Contabilidad', image: '/images/contabilidad.png', points: ['Estructuración contable', 'Estados financieros confiables', 'Capacitación en cifras clave'] },
  { name: 'Finanzas', image: '/images/finanzas.png', points: ['Análisis financiero integral', 'Presupuestos y flujo de efectivo', 'Planeación patrimonial'] },
  { name: 'Fiscalidad', image: '/images/fiscalidad.png', points: ['Planeación y cumplimiento fiscal', 'Litigios y devoluciones ante el SAT', 'Dictámenes y auditorías fiscales'] },
  { name: 'Derecho Corporativo', image: '/images/derecho.png', points: ['Constitución de sociedades', 'Contratos y gobierno corporativo', 'Auditoría legal y reestructuras'] },
  { name: 'Inversión Extranjera', image: '/images/inversion.png', points: ['Registro y renovación RNIE', 'Informes corporativos', 'Asesoría regulatoria internacional'] },
  { name: 'Propiedad Industrial', image: '/images/propiedad.png', points: ['Registro de marcas y patentes', 'Contratos de franquicia y tecnología', 'Protección de propiedad intelectual'] },
  { name: 'Concursos Mercantiles', image: '/images/concursos.png', points: ['Acompañamiento legal en insolvencia', 'Reestructuración de obligaciones', 'Defensa técnica ante jueces'] },
  { name: 'Capacitación Empresarial', image: '/images/capacitacion.png', points: ['Formación en finanzas y fiscalidad', 'Desarrollo de liderazgo empresarial', 'Talleres estratégicos a medida'] }
]
const form = reactive({ nombre: '', empresa: '', telefono: '', correo: '', asunto: '', mensaje: '' })
function prepareEmail() {
  const text = [
    `Nombre: ${form.nombre}`, `Empresa: ${form.empresa || '-'}`,
    `Teléfono: ${form.telefono || '-'}`, `Correo: ${form.correo}`,
    '', 'Mensaje:', form.mensaje
  ].join('\n')
  window.location.href = `mailto:contacto@remaconsultores.com?subject=${encodeURIComponent(form.asunto)}&body=${encodeURIComponent(text)}`
}
</script>

<template>
  <main>
    <section class="hero" aria-labelledby="hero-title">
      <div class="container hero-content">
        <div class="hero-text">
          <p class="eyebrow hero-label">ESTRATEGIA <span>·</span> CONSULTORÍA <span>·</span> <b>RESULTADOS</b></p>
          <h1 id="hero-title">Asesoría Integral<br>para <span>tu Negocio</span></h1>
          <p class="hero-subtitle">Contabilidad, Finanzas, Derecho Corporativo y más.</p>
          <p class="hero-description">Impulsamos líderes y empresas con soluciones estratégicas.</p>
          <NuxtLink class="button button-primary" to="/#contacto">Agenda tu consulta <span aria-hidden="true">→</span></NuxtLink>
          <div class="hero-features">
            <div class="feature"><span class="icon-circle"><span aria-hidden="true">◇</span></span><span>Soluciones<br>a la medida</span></div>
            <div class="feature"><span class="icon-circle"><span aria-hidden="true">♙</span></span><span>Acompañamiento<br>especializado</span></div>
            <div class="feature"><span class="icon-circle"><span aria-hidden="true">♢</span></span><span>Experiencia<br>profesional</span></div>
          </div>
        </div>
      </div>
    </section>

    <section id="nosotros" class="section section-about">
      <div class="container about-grid">
        <div>
          <p class="eyebrow">SOMOS REMA CONSULTORES</p>
          <h2>Transformamos el futuro de <span>emprendedores y empresas</span></h2>
          <p class="body-copy">Somos un equipo de profesionales comprometidos con transformar el futuro de los negocios a través de asesoría, entrenamiento y educación especializada para emprendedores, líderes empresariales y los profesionales del mañana.</p>
          <NuxtLink class="button button-outline" to="/#contacto">Conoce más sobre nosotros <span aria-hidden="true">→</span></NuxtLink>
        </div>
        <div class="about-photo-wrap">
          <img class="about-photo" src="/images/equipo-consultores.webp" alt="Equipo de consultores conversando en una oficina" loading="lazy">
          <div class="values-box">
            <div><span class="icon-circle" aria-hidden="true">♙</span><p><strong>Liderazgo</strong><small>empresarial</small></p></div>
            <div><span class="icon-circle" aria-hidden="true">↗</span><p><strong>Crecimiento</strong><small>con propósito</small></p></div>
            <div><span class="icon-circle" aria-hidden="true">♜</span><p><strong>Profesionalismo</strong><small>en cada paso</small></p></div>
          </div>
        </div>
      </div>
    </section>

    <section id="servicios" class="section section-services">
      <div class="container">
        <div class="section-heading">
          <p class="eyebrow">NUESTROS SERVICIOS</p>
          <h2>Servicios</h2>
          <p>Soluciones profesionales y estratégicas para cada etapa de tu empresa.</p>
        </div>
        <div class="services-grid">
          <article v-for="service in services" :key="service.name" class="service-card">
            <img :src="service.image" alt="" loading="lazy" width="92" height="92">
            <h3>{{ service.name }}</h3>
            <ul><li v-for="point in service.points" :key="point">{{ point }}</li></ul>
          </article>
        </div>
      </div>
    </section>

    <BlogCarousel />

    <section id="contacto" class="section section-contact">
      <div class="container contact-grid">
        <div>
          <p class="eyebrow">CONTÁCTANOS</p>
          <h2>Hablemos de <span>tu próximo proyecto</span></h2>
          <p class="body-copy">Estamos listos para asesorarte. Escríbenos y uno de nuestros especialistas se pondrá en contacto contigo.</p>
          <address class="contact-data">
            <div><span class="icon-circle" aria-hidden="true">⌖</span><p><strong>Oficinas</strong>Veracruz, CDMX y Monterrey</p></div>
            <div><span class="icon-circle" aria-hidden="true">✉</span><p><strong>Email</strong><a href="mailto:contacto@remaconsultores.com">contacto@remaconsultores.com</a></p></div>
            <div><span class="icon-circle" aria-hidden="true">☎</span><p><strong>Teléfonos</strong><a href="tel:+522293878119">229 387 8119</a> · <a href="tel:+525588527869">558 852 7869</a></p></div>
          </address>
        </div>
        <div class="form-card">
          <h3>Envíanos un mensaje</h3>
          <p class="form-help">El botón abrirá tu correo con los datos del mensaje preparados.</p>
          <form id="contact-form" @submit.prevent="prepareEmail">
            <div class="form-field"><label for="nombre">Nombre *</label><input id="nombre" v-model.trim="form.nombre" name="nombre" placeholder="Tu nombre" autocomplete="name" required></div>
            <div class="form-field"><label for="empresa">Empresa</label><input id="empresa" v-model.trim="form.empresa" name="empresa" placeholder="Tu empresa" autocomplete="organization"></div>
            <div class="form-field"><label for="telefono">Teléfono</label><input id="telefono" v-model.trim="form.telefono" name="telefono" type="tel" placeholder="Tu teléfono" autocomplete="tel"></div>
            <div class="form-field"><label for="correo">Correo *</label><input id="correo" v-model.trim="form.correo" name="correo" type="email" placeholder="correo@ejemplo.com" autocomplete="email" required></div>
            <div class="form-field full"><label for="asunto">Asunto *</label><input id="asunto" v-model.trim="form.asunto" name="asunto" placeholder="¿En qué podemos ayudarte?" required></div>
            <div class="form-field full"><label for="mensaje">Mensaje *</label><textarea id="mensaje" v-model.trim="form.mensaje" name="mensaje" rows="4" placeholder="Cuéntanos sobre tu proyecto" required></textarea></div>
            <button class="button button-primary full" type="submit">Preparar mensaje <span aria-hidden="true">→</span></button>
          </form>
        </div>
      </div>
    </section>
  </main>
</template>
