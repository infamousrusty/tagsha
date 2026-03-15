export interface Tag {
  name: string
  sha: string
  message: string
  author_name: string
  author_email: string
  date: string
  commit_url: string
}

export interface TagsResponse {
  owner: string
  repo: string
  total_count: number
  truncated: boolean
  tags: Tag[]
  cached_at: string
  github_rate_limit_remaining: number
}

export interface ResolveResponse {
  owner: string
  repo: string
  redirect_url: string
}

export interface APIError {
  error: string
  message: string
  request_id: string
}

export type LoadState = 'idle' | 'loading' | 'success' | 'error'

/** Normalised tag shape used throughout the UI */
export interface TagEntry {
  name: string
  sha?: string
  message?: string
  author?: string
  date?: string
  commitUrl?: string
  /** Full image ref prefix, e.g. ghcr.io/org/image */
  repo?: string
  digest?: string
  size?: string
}

export type Registry = 'ghcr' | 'dockerhub' | 'other'
