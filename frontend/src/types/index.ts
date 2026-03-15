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
