// 图片上传封装：调用 POST /images，返回契约类型 ImageInfo。
import { client, errorMessage } from '@/api/client'

import type { components } from '@/api/schema'

export type ImageInfo = components['schemas']['ImageInfo']

/** 上传单个图片文件，成功返回 ImageInfo，失败返回 null（错误信息由调用方提示）。 */
export async function uploadImage(file: File): Promise<ImageInfo | null> {
  const fd = new FormData()
  fd.append('file', file)
  // 契约声明为 multipart 的 { file: binary }；运行时以 FormData 传输，
  // 此处类型断言仅用于适配 openapi-fetch 对 binary 的类型表示。
  const { data, error } = await client.POST('/images', {
    body: fd as unknown as { file: string },
    bodySerializer: (body) => body as unknown as FormData,
  })
  if (error || !data) {
    throw new Error(errorMessage(error))
  }
  return data
}