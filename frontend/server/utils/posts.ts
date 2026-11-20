import { randomUUID } from 'node:crypto'
import { promises as fs } from 'node:fs'
import { dirname, isAbsolute, resolve } from 'node:path'
import type { BlogPost, NewBlogPost } from '../../shared/types/blog'

const allowedCategories: Record<string, string> = {
  'Finanzas': '/images/blog-finanzas.webp',
  'Fiscalidad': '/images/blog-fiscal.webp',
  'Derecho corporativo': '/images/blog-corporativo.webp',
  'Contabilidad': '/images/blog-finanzas.webp',
  'Inversión extranjera': '/images/blog-corporativo.webp',
  'Propiedad industrial': '/images/blog-corporativo.webp',
  'Concursos mercantiles': '/images/blog-corporativo.webp',
  'Capacitación empresarial': '/images/blog-finanzas.webp',
  'Otros': '/images/blog-finanzas.webp'
}

const configuredFile = process.env.REMA_POSTS_FILE
const file = configuredFile
  ? (isAbsolute(configuredFile) ? configuredFile : resolve(process.cwd(), configuredFile))
  : resolve(process.cwd(), 'data', 'posts.json')

// Serializamos las escrituras de este proceso para no perder entradas simultáneas.
let queue: Promise<unknown> = Promise.resolve()

export async function readPosts(): Promise<BlogPost[]> {
  try {
    const raw = await fs.readFile(file, 'utf8')
    const posts: BlogPost[] = JSON.parse(raw)
    return Array.isArray(posts) ? posts.sort((a, b) => b.createdAt.localeCompare(a.createdAt)) : []
  } catch (err: unknown) {
    if ((err as NodeJS.ErrnoException).code === 'ENOENT') return []
    throw err
  }
}

function validate(payload: unknown): NewBlogPost {
  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) throw new Error('Datos inválidos.')
  const b = payload as Record<string, unknown>
  const val = (key: string) => typeof b[key] === 'string' ? (b[key] as string).trim() : ''
  const title = val('title')
  const author = val('author')
  const summary = val('summary')
  const content = val('content')
  const category = val('category')
  const website = val('website')
  if (website) throw new Error('La publicación no se pudo procesar.') // Campo anti-bots.
  if (title.length < 8 || title.length > 110) throw new Error('El título debe tener entre 8 y 110 caracteres.')
  if (author.length < 2 || author.length > 60) throw new Error('El nombre debe tener entre 2 y 60 caracteres.')
  if (summary.length < 20 || summary.length > 260) throw new Error('El resumen debe tener entre 20 y 260 caracteres.')
  if (content.length < 80 || content.length > 10000) throw new Error('El artículo debe tener entre 80 y 10 000 caracteres.')
  if (!Object.hasOwn(allowedCategories, category)) throw new Error('Selecciona una categoría válida.')
  // No se permiten etiquetas HTML ni scripts: la vista además utiliza escape de Vue.
  if (/<\s*\/?\s*\w/i.test([title, author, summary, content].join(' '))) throw new Error('Escribe solo texto, sin etiquetas HTML.')
  return { title, author, summary, content, category, website: '' }
}

function slugify(text: string) {
  return text.toLowerCase().normalize('NFD').replace(/[\u0300-\u036f]/g, '')
    .replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '').slice(0, 58)
}

export function createPost(payload: unknown): Promise<BlogPost> {
  const task = async (): Promise<BlogPost> => {
    const input = validate(payload)
    const posts = await readPosts()
    if (posts.length >= 500) throw new Error('Se alcanzó la capacidad de publicaciones de esta instalación.')
    const id = randomUUID()
    const post: BlogPost = {
      id,
      slug: `${slugify(input.title) || 'articulo'}-${id.slice(0, 8)}`,
      title: input.title,
      author: input.author,
      summary: input.summary,
      content: input.content,
      category: input.category,
      image: allowedCategories[input.category]!,
      createdAt: new Date().toISOString()
    }
    await fs.mkdir(dirname(file), { recursive: true })
    const tmp = `${file}.${id}.tmp`
    await fs.writeFile(tmp, JSON.stringify([post, ...posts], null, 2), 'utf8')
    await fs.rename(tmp, file)
    return post
  }
  const result = queue.then(task)
  queue = result.then(() => undefined, () => undefined)
  return result
}
