export interface BlogPost {
  id: string
  slug: string
  title: string
  author: string
  category: string
  summary: string
  content: string
  image: string
  createdAt: string
}
export interface NewBlogPost {
  title: string
  author: string
  category: string
  summary: string
  content: string
  website?: string
}
