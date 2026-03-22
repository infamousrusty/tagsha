// SPDX-License-Identifier: LGPL-3.0-or-later
// Copyright (C) 2026 infamousrusty

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { fetchTags, resolveQuery } from './client'

const mockFetch = vi.fn()
global.fetch = mockFetch

describe('fetchTags', () => {
  beforeEach(() => mockFetch.mockReset())

  it('returns data and cache status on success', async () => {
    const mockData = { owner: 'torvalds', repo: 'linux', tags: [], total_count: 0 }
    mockFetch.mockResolvedValueOnce({
      ok: true,
      headers: { get: (h: string) => h === 'X-Cache' ? 'HIT' : null },
      json: async () => mockData,
    })
    const result = await fetchTags('torvalds', 'linux')
    expect(result.cacheStatus).toBe('HIT')
    expect(result.data.owner).toBe('torvalds')
  })

  it('throws on HTTP error', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 404,
      headers: { get: () => null },
      json: async () => ({ message: 'Repository not found', error: 'REPOSITORY_NOT_FOUND', request_id: 'x' }),
    })
    await expect(fetchTags('invalid', 'repo')).rejects.toThrow('Repository not found')
  })
})

describe('resolveQuery', () => {
  it('parses and returns owner/repo', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ owner: 'golang', repo: 'go', redirect_url: '/api/v1/tags/golang/go' }),
    })
    const result = await resolveQuery('golang/go')
    expect(result.owner).toBe('golang')
  })
})
