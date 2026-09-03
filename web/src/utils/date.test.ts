import { describe, expect, it } from 'vitest'

import { addDays, formatDate, isOverdue, parseDate, today } from './date'

describe('formatDate / parseDate', () => {
  it('格式化为 YYYY-MM-DD', () => {
    expect(formatDate(new Date(2026, 8, 3))).toBe('2026-09-03')
  })

  it('解析 YYYY-MM-DD 为本地日期', () => {
    const d = parseDate('2026-09-03')
    expect(d.getFullYear()).toBe(2026)
    expect(d.getMonth()).toBe(8)
    expect(d.getDate()).toBe(3)
  })

  it('today 返回今天', () => {
    expect(today()).toBe(formatDate(new Date()))
  })
})

describe('addDays', () => {
  it('加 7 天', () => {
    expect(addDays('2026-09-03', 7)).toBe('2026-09-10')
  })

  it('跨月加 7 天', () => {
    expect(addDays('2026-09-29', 7)).toBe('2026-10-06')
  })

  it('跨年加 7 天', () => {
    expect(addDays('2026-12-29', 7)).toBe('2027-01-05')
  })
})

describe('isOverdue', () => {
  it('已整改不算逾期', () => {
    expect(isOverdue('2020-01-01', '已整改')).toBe(false)
  })

  it('未过期不算逾期', () => {
    const future = addDays(today(), 3)
    expect(isOverdue(future, '待整改')).toBe(false)
  })

  it('待整改且已过期算逾期', () => {
    expect(isOverdue('2000-01-01', '待整改')).toBe(true)
  })

  it('整改受阻且已过期算逾期', () => {
    expect(isOverdue('2000-01-01', '整改受阻')).toBe(true)
  })
})