import { readPosts } from '../../utils/posts'
export default defineEventHandler(async () => readPosts())
