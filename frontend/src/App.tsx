import { useState, useCallback } from 'react'
import { SearchBar } from './components/SearchBar'
import { TagList } from './components/TagList'
import { fetchTags, resolveQuery } from './api/client'
import type { TagsResponse, LoadState } from './types'
import './styles.css'

interface State {
  loadState: LoadState
  response: TagsResponse | null
  cacheStatus: string
  error: string | null
}

export default function App() {
  const [state, setState] = useState<State>({
    loadState: 'idle',
    response: null,
    cacheStatus: '',
    error: null,
  })

  const handleSearch = useCallback(async (query: string) => {
    setState({ loadState: 'loading', response: null, cacheStatus: '', error: null })

    try {
      const resolved = await resolveQuery(query)
      const { data, cacheStatus } = await fetchTags(resolved.owner, resolved.repo)

      setState({
        loadState: 'success',
        response: data,
        cacheStatus,
        error: null,
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'An unexpected error occurred.'
      setState({ loadState: 'error', response: null, cacheStatus: '', error: message })
    }
  }, [])

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-inner">
          <a href="/" className="logo" aria-label="TagSha home">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true" focusable="false">
              <path d="M20.59 13.41l-7.17 7.17a2 2 0 01-2.83 0L2 12V2h10l8.59 8.59a2 2 0 010 2.82z" />
              <line x1="7" y1="7" x2="7.01" y2="7" />
            </svg>
            TagSha
          </a>
          <p className="tagline">Resolve GitHub tags to exact commit SHAs</p>
          <a
            href="https://github.com/infamousrusty/tagsha"
            target="_blank"
            rel="noopener noreferrer"
            className="github-link"
            aria-label="View TagSha on GitHub"
          >
            <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" focusable="false">
              <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
            </svg>
          </a>
        </div>
      </header>

      <main className="app-main">
        <SearchBar onSearch={handleSearch} isLoading={state.loadState === 'loading'} />

        {state.loadState === 'loading' && (
          <div className="status-msg loading" role="status" aria-live="polite">
            <span className="spinner" aria-hidden="true" />
            Fetching tags from GitHub…
          </div>
        )}

        {state.loadState === 'error' && (
          <div className="status-msg error" role="alert">
            <strong>Error:</strong> {state.error}
          </div>
        )}

        {state.loadState === 'success' && state.response && (
          <TagList response={state.response} cacheStatus={state.cacheStatus} />
        )}

        {state.loadState === 'idle' && (
          <div className="idle-msg">
            <p>Enter a GitHub repository above to browse its version tags and resolved commit SHAs.</p>
            <p className="idle-hint">Supports <code>owner/repo</code> format and full <code>github.com</code> URLs.</p>
          </div>
        )}
      </main>

      <footer className="app-footer">
        <p>
          TagSha — open source &middot;{' '}
          <a href="https://github.com/infamousrusty/tagsha" target="_blank" rel="noopener noreferrer">GitHub</a>
          {' '}&middot;{' '}
          <a href="/health">Health</a>
        </p>
      </footer>
    </div>
  )
}
