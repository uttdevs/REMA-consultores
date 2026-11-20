<script setup lang="ts">
type ChatMessage = { text: string; who: 'bot' | 'user'; link?: { text: string; href: string } }
const open = ref(false)
const question = ref('')
const messageList = ref<HTMLElement | null>(null)
const messages = ref<ChatMessage[]>([
  { who: 'bot', text: '¡Hola! Bienvenido a REMA Consultores. Soy un asistente informativo. ¿Te ayudo con nuestros servicios, el blog o cómo contactarnos?' }
])
const contact = { text: 'Ir a contacto', href: '/#contacto' }
function normalized(value: string) {
  return value.toLowerCase().normalize('NFD').replace(/[\u0300-\u036f]/g, '').replace(/[^a-z0-9\s]/g, ' ').replace(/\s+/g, ' ').trim()
}
function answer(raw: string): ChatMessage {
  const q = normalized(raw)
  const has = (...words: string[]) => words.some(w => q.includes(w))
  if (has('blog', 'public', 'articulo', 'escribir', 'compartir')) return { who: 'bot', text: 'Puedes explorar artículos en el carrusel del blog y hacer clic en “Publicar un artículo” para compartir tus conocimientos. Los textos se guardan en el servidor.', link: { text: 'Ir al blog', href: '/#blog' } }
  if (has('agend', 'cita', 'consulta', 'reunion')) return { who: 'bot', text: 'Para solicitar una consulta, escribe a contacto@remaconsultores.com o utiliza el formulario de la página.', link: contact }
  if (has('ubic', 'donde', 'oficina', 'sucursal')) return { who: 'bot', text: 'REMA Consultores indica presencia en Veracruz, Ciudad de México y Monterrey. Para consultar una dirección exacta, contáctanos.', link: contact }
  if (has('telefono', 'llamar', 'correo', 'contact', 'whatsapp')) return { who: 'bot', text: 'Puedes llamar al 229 387 8119 o 558 852 7869, o escribir a contacto@remaconsultores.com.', link: contact }
  if (has('cost', 'precio', 'cotiz', 'tarifa')) return { who: 'bot', text: 'La cotización depende de las necesidades de tu empresa. Contacta al equipo para una propuesta personalizada.', link: contact }
  if (has('contabil', 'finanz', 'fiscal', 'derecho', 'corporativ', 'inversion', 'marca', 'patente', 'concurs', 'mercantil', 'capacit', 'servicio', 'sat')) return { who: 'bot', text: 'REMA ofrece Contabilidad, Finanzas, Fiscalidad, Derecho Corporativo, Inversión Extranjera, Propiedad Industrial, Concursos Mercantiles y Capacitación Empresarial.', link: { text: 'Ver servicios', href: '/#servicios' } }
  if (has('hola', 'buenas', 'buen dia')) return { who: 'bot', text: '¡Hola! Puedo orientarte con información sobre nuestros servicios, oficinas, blog y contacto.' }
  return { who: 'bot', text: 'Soy un asistente de preguntas frecuentes. Puedes preguntarme por servicios, el blog o cómo solicitar una consulta. Para un caso específico, contacta a un especialista.', link: contact }
}
function send(value = question.value) {
  const clean = value.trim().slice(0, 300)
  if (!clean) return
  open.value = true
  messages.value.push({ who: 'user', text: clean })
  messages.value.push(answer(clean))
  question.value = ''
  nextTick(() => { if (messageList.value) messageList.value.scrollTop = messageList.value.scrollHeight })
}
function closeChat() { open.value = false }
function onKeyDown(event: KeyboardEvent) { if (event.key === 'Escape') closeChat() }
onMounted(() => document.addEventListener('keydown', onKeyDown))
onBeforeUnmount(() => document.removeEventListener('keydown', onKeyDown))
</script>

<template>
  <aside class="rema-chat" aria-label="Asistente de REMA Consultores">
    <section v-if="open" class="rema-chat-panel" aria-label="Conversación con el asistente de REMA">
      <header class="rema-chat-header">
        <span class="rema-chat-header-icon"><svg class="rema-scale" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M32 8v44M15 19h34M27 12h10M23 53h18M17 59h30"/><path d="M15 19 6 37m9-18 9 18m25-18-9 18m9-18 9 18"/><path d="M6 37h18c-1.5 6-4.5 9-9 9s-7.5-3-9-9Zm34 0h18c-1.5 6-4.5 9-9 9s-7.5-3-9-9Z"/></svg></span>
        <div><strong>Asistente REMA</strong><small>● Respuestas automáticas</small></div>
        <button type="button" class="rema-chat-close" aria-label="Cerrar chat" @click="closeChat">×</button>
      </header>
      <div ref="messageList" class="rema-chat-messages" role="log" aria-live="polite">
        <div v-for="(message, index) in messages" :key="index" :class="['rema-chat-message', `rema-chat-message-${message.who}`]">
          <span v-if="message.who === 'bot'" class="rema-chat-message-name">Asistente REMA</span>
          <p>{{ message.text }}</p>
          <NuxtLink v-if="message.link" class="rema-chat-message-link" :to="message.link.href" @click="closeChat">{{ message.link.text }} →</NuxtLink>
        </div>
      </div>
      <div class="rema-chat-suggestions">
        <span>Consulta rápida</span>
        <div class="rema-chat-chips">
          <button v-for="chip in ['Servicios', 'Publicar blog', 'Agendar consulta', 'Contacto']" :key="chip" type="button" @click="send(chip)">{{ chip }}</button>
        </div>
      </div>
      <form class="rema-chat-form" @submit.prevent="send()">
        <label class="sr-only" for="rema-chat-input">Escribe tu pregunta</label>
        <input id="rema-chat-input" v-model="question" maxlength="300" autocomplete="off" placeholder="Escribe tu pregunta..." required>
        <button type="submit" aria-label="Enviar mensaje">➤</button>
      </form>
      <p class="rema-chat-disclaimer">Asistente informativo, no chat en vivo. Evita compartir información sensible.</p>
    </section>
    <button class="rema-chat-launcher" type="button" :aria-expanded="open" :aria-label="open ? 'Cerrar asistente REMA' : 'Abrir asistente REMA'" @click="open = !open">
      <span class="rema-chat-launcher-icon"><svg class="rema-scale" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M32 8v44M15 19h34M27 12h10M23 53h18M17 59h30"/><path d="M15 19 6 37m9-18 9 18m25-18-9 18m9-18 9 18"/><path d="M6 37h18c-1.5 6-4.5 9-9 9s-7.5-3-9-9Zm34 0h18c-1.5 6-4.5 9-9 9s-7.5-3-9-9Z"/></svg></span>
      <span>{{ open ? 'Cerrar chat' : '¿Necesitas ayuda?' }}</span>
    </button>
  </aside>
</template>
