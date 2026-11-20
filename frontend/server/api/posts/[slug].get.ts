import { createError, getRouterParam } from 'h3'
import { readPosts } from '../../utils/posts'
export default defineEventHandler(async (event) => {
  const slug = getRouterParam(event, 'slug') || ''
  if (!/^[a-z0-9-]{1,100}$/.test(slug)) throw createError({ statusCode: 404, statusMessage: 'Artículo no encontrado.' })
  const post = (await readPosts()).find(item => item.slug === slug)
  if (!post) throw createError({ statusCode: 404, statusMessage: 'Artículo no encontrado.' })
  return post
})
