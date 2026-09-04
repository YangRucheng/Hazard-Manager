// 类型化 API 客户端：契约来自 openapi.yaml 生成的 schema.d.ts，全链路无 any。
import createClient from 'openapi-fetch'

import { useAuth } from '@/stores/auth'

import { API_BASE } from './config'
import type { paths } from './schema'

export { API_BASE }

export const client = createClient<paths>({
  baseUrl: API_BASE,
})

// 鉴权中间件：请求注入 Bearer 令牌；401 响应清除本地会话并回到登录页。
client.use({
  async onRequest({ request }) {
    const { token } = useAuth()
    if (token.value) {
      request.headers.set('Authorization', `Bearer ${token.value}`)
    }
    return request
  },
  async onResponse({ response }) {
    if (response.status === 401 && !window.location.pathname.startsWith('/login')) {
      const { clearAuth } = useAuth()
      clearAuth()
      window.location.href = '/login'
      return response
    }
    return response
  },
})

/** 响应错误结构（契约 Error schema）。 */
export interface ApiError {
  code: string
  message: string
}

/** 从 openapi-fetch 错误对象安全提取消息。 */
export function errorMessage(err: unknown): string {
  if (err && typeof err === 'object' && 'message' in err && typeof err.message === 'string') {
    return err.message
  }
  return '请求失败，请稍后重试'
}

/** 由图片 uuid 拼原图 URL（base 已含 /api/v1）。 */
export function imageUrl(id: string): string {
  return `${API_BASE}/images/${id}`
}

/** 由图片 uuid 拼缩略图 URL（列表预览用）。 */
export function thumbnailUrl(id: string): string {
  return `${API_BASE}/images/${id}/thumbnail`
}

// ---- 鉴权图片加载 ----
// 全部接口（含图片二进制）都要求 Bearer 头，而浏览器 <img> / window.open 无法
// 携带请求头，因此先经 fetch 带令牌取回二进制，再转为 objectURL 供页面渲染。

const blobUrlCache = new Map<string, Promise<string>>()

async function fetchBlobUrl(url: string): Promise<string> {
  const { token } = useAuth()
  const headers = new Headers()
  if (token.value) {
    headers.set('Authorization', `Bearer ${token.value}`)
  }
  const res = await fetch(url, { headers })
  if (!res.ok) {
    throw new Error(
      res.status === 401 ? '登录已过期，请重新登录' : `图片加载失败（HTTP ${res.status}）`,
    )
  }
  return URL.createObjectURL(await res.blob())
}

/**
 * 带 Bearer 拉取图片资源并返回 objectURL。
 * 会话级缓存（objectURL 不主动回收，同图去重）、并发去重；失败不缓存、可重试。
 */
export function authedBlobUrl(url: string): Promise<string> {
  let cached = blobUrlCache.get(url)
  if (!cached) {
    cached = fetchBlobUrl(url).catch((err: unknown) => {
      blobUrlCache.delete(url)
      throw err
    })
    blobUrlCache.set(url, cached)
  }
  return cached
}