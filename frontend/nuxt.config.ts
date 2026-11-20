export default defineNuxtConfig({
  compatibilityDate: '2025-07-01',
  devtools: { enabled: true },
  css: ['~/assets/css/main.css'],
  app: {
    head: {
      htmlAttrs: { lang: 'es-MX' },
      title: 'REMA Consultores | Asesoría integral para tu negocio',
      meta: [
        { name: 'description', content: 'Contabilidad, finanzas, fiscalidad, derecho corporativo y formación empresarial en REMA Consultores.' },
        { name: 'theme-color', content: '#FBFAFD' }
      ],
      link: [
        { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico?v=2' },
        { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-32x32.png?v=2' },
        { rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-touch-icon.png?v=2' },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700&family=Playfair+Display:wght@500;600;700&display=swap' }
      ]
    }
  }
})
