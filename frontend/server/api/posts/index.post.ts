import { createError, getRequestIP, readBody, setResponseStatus } from 'h3'
import { createPost } from '../../utils/posts'

const lastSubmission = new Map<string, number>()
export default defineEventHandler(async (event) => {
  // Límite básico local para reducir envíos accidentales y spam.
  const ip = getRequestIP(event) || 'unknown'
  const now = Date.now()
  const previous = lastSubmission.get(ip) || 0
  if (now - previous < 12_000) {
    throw createError({ statusCode: 429, statusMessage: 'Espera unos segundos antes de publicar otro artículo.' })
  }
  if (lastSubmission.size > 1000) {
    for (const [key, time] of lastSubmission) if (now - time > 600_000) lastSubmission.delete(key)
  }
  try {
    const post = await createPost(await readBody(event))
    lastSubmission.set(ip, now)
    setResponseStatus(event, 201)
    return post
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : 'No se pudo publicar el artículo.'
    if (/EACCES|EROFS|EPERM/.test(message)) {
      throw createError({ statusCode: 503, statusMessage: 'El servidor no tiene permiso para guardar publicaciones. Configura REMA_POSTS_FILE.' })
    }
    throw createError({ statusCode: 400, statusMessage: message })
  }
})
