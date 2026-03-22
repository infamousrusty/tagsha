// SPDX-License-Identifier: LGPL-3.0-or-later
// Copyright (C) 2026 infamousrusty

import type { TagsResponse, ResolveResponse } from '../types'

const BASE = '/api/v1'

export interface FetchTagsResult {
  data: TagsResponse
  cacheStatus: string
}

export async function fetchTags(
  owner: string,
  repo: string
): Promise<FetchTagsResult> {
  const url = `${BASE}/tags/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}`
  const res = await fetch(url)
  const cacheStatus = res.headers.get('X-Cache') ?? 'UNKNOWN'

  if (!res.ok) {
    const err = await res.json()
    throw new Error(err.message ?? `HTTP ${res.status}`)
  }

  const data: TagsResponse = await res.json()
  return { data, cacheStatus }
}

export async function resolveQuery(query: string): Promise<ResolveResponse> {
  const res = await fetch(`${BASE}/resolve`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query }),
  })

  if (!res.ok) {
    const err = await res.json()
    throw new Error(err.message ?? `HTTP ${res.status}`)
  }

  return res.json()
}
