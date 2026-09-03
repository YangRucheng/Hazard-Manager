import { describe, expect, it } from 'vitest'

import { normalizeBaseUrl, resolveApiBaseUrl } from './config'

describe('normalizeBaseUrl', () => {
  it('去尾部斜杠', () => {
    expect(normalizeBaseUrl('https://a.com/api/v1/')).toBe('https://a.com/api/v1')
  })

  it('根保留', () => {
    expect(normalizeBaseUrl('/')).toBe('/')
  })
})

describe('resolveApiBaseUrl', () => {
  it('未配置回退相对 /api/v1', () => {
    expect(resolveApiBaseUrl(undefined)).toBe('/api/v1')
    expect(resolveApiBaseUrl('')).toBe('/api/v1')
    expect(resolveApiBaseUrl('   ')).toBe('/api/v1')
  })

  it('裸域名补 /api/v1', () => {
    expect(resolveApiBaseUrl('https://hazard-manager.qcloud.19890605.xyz')).toBe(
      'https://hazard-manager.qcloud.19890605.xyz/api/v1',
    )
    expect(resolveApiBaseUrl('https://example.com/')).toBe('https://example.com/api/v1')
    expect(resolveApiBaseUrl('http://localhost:8090')).toBe('http://localhost:8090/api/v1')
  })

  it('已含路径则直接使用', () => {
    expect(resolveApiBaseUrl('https://example.com/api/v1')).toBe('https://example.com/api/v1')
    expect(resolveApiBaseUrl('https://example.com/gateway/api/v1/')).toBe(
      'https://example.com/gateway/api/v1',
    )
  })

  it('相对路径直接使用', () => {
    expect(resolveApiBaseUrl('/api/v1')).toBe('/api/v1')
    expect(resolveApiBaseUrl('/api')).toBe('/api')
  })

  it('非法协议/URL 回退相对', () => {
    expect(resolveApiBaseUrl('ftp://example.com')).toBe('/api/v1')
    expect(resolveApiBaseUrl('not a url')).toBe('/api/v1')
  })
})