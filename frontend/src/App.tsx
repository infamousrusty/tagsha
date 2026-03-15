import { useState, useCallback, useEffect, useRef } from 'react'
import { SearchBar } from './components/SearchBar'
import { TagList } from './components/TagList'
import { TagDrawer } from './components/TagDrawer'
import { ToastContainer } from './components/Toast'
import { HelpModal } from './components/HelpModal'
import { fetchTags, resolveQuery } from './api/client'
import type { TagsResponse, LoadState, TagEntry } from './types'
import './styles.css'

const IS_SELF_HOSTED = import.meta.env.VITE_SELF_HOSTED === 'true'

interface State {
  loadState: LoadState
  response: TagsResponse | null
  cacheStatus: string
  error: string | null
  errorKind: 'not_found' | 'rate_limit' | 'server' | 'network' | null
  currentQuery: string
}

export interface Toast {
  id: number
  message: string
  kind: 'success' | 'error' | 'info'
}

export default function App() {
  const [state, setState] = useState<State>({
    loadState: 'idle', response: null,
    cacheStatus: '', error: null, errorKind: null, currentQuery: '',
  })
  const [selectedTag, setSelectedTag] = useState<TagEntry | null>(null)
  const [toasts, setToasts] = useState<Toast[]>([])
  const [showHelp, setShowHelp] = useState(false)
  const [theme, setTheme] = useState<'dark' | 'light'>(() => {
    const stored = localStorage.getItem('tagsha-theme')
    return (stored === 'light' ? 'light' : 'dark')
  })
  const toastId = useRef(0)

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem('tagsha-theme', theme)
  }, [theme])

  // Global keyboard shortcuts
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === '?' && !['INPUT','TEXTAREA'].includes((e.target as HTMLElement).tagName)) {
        setShowHelp(h => !h)
      }
      if (e.key === 'Escape') {
        setSelectedTag(null)
        setShowHelp(false)
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  const addToast = useCallback((message: string, kind: Toast['kind'] = 'info') => {
    const id = ++toastId.current
    setToasts(t => [...t, { id, message, kind }])
    setTimeout(() => setToasts(t => t.filter(x => x.id !== id)), 3500)
  }, [])

  const handleSearch = useCallback(async (query: string) => {
    setState(s => ({ ...s, loadState: 'loading', response: null, error: null, errorKind: null, currentQuery: query }))
    try {
      const resolved = await resolveQuery(query)
      const { data, cacheStatus } = await fetchTags(resolved.owner, resolved.repo)
      setState(s => ({ ...s, loadState: 'success', response: data, cacheStatus }))
    } catch (err) {
      const message = err instanceof Error ? err.message : 'An unexpected error occurred.'
      const lower = message.toLowerCase()
      const errorKind =
        lower.includes('not found') ? 'not_found' :
        lower.includes('rate') ? 'rate_limit' :
        lower.includes('network') || lower.includes('fetch') ? 'network' : 'server'
      setState(s => ({ ...s, loadState: 'error', error: message, errorKind }))
    }
  }, [])

  const handleRetry = useCallback(() => {
    if (state.currentQuery) handleSearch(state.currentQuery)
  }, [state.currentQuery, handleSearch])

  const errorMeta = {
    not_found:  { icon: '🔍', title: 'Repository not found', hint: 'Check the owner/repo name and try again.' },
    rate_limit: { icon: '⏳', title: 'Rate limit reached', hint: 'GitHub is throttling requests. Wait a moment and retry.' },
    network:    { icon: '🌐', title: 'Network error', hint: 'Check your connection and try again.' },
    server:     { icon: '⚠️', title: 'Server error', hint: 'Something went wrong on our end. Please retry.' },
  }

  return (
    <>
      <a href="#main-content" className="skip-link">Skip to content</a>

      {IS_SELF_HOSTED && (
        <div className="self-hosted-banner" role="banner">
          🏠 Self-hosted Tagsha instance
        </div>
      )}

      <div className="app">
        <header className="app-header">
          <div className="header-inner">
            <a href="/" className="logo" aria-label="TagSha home">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                <path d="M20.59 13.41l-7.17 7.17a2 2 0 01-2.83 0L2 12V2h10l8.59 8.59a2 2 0 010 2.82z" />
                <line x1="7" y1="7" x2="7.01" y2="7" />
              </svg>
              TagSha
            </a>
            <p className="tagline">Resolve Docker tags to exact commit SHAs</p>

            <div className="header-actions">
              <button
                className="icon-btn"
                onClick={() => setTheme(t => t === 'dark' ? 'light' : 'dark')}
                aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
                title={`Toggle ${theme === 'dark' ? 'light' : 'dark'} mode`}
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                  {theme === 'dark'
                    ? <><circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></>
                    : <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z"/>}
                </svg>
              </button>
              <button
                className="icon-btn"
                onClick={() => setShowHelp(true)}
                aria-label="Keyboard shortcuts"
                title="Keyboard shortcuts (?)"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                  <circle cx="12" cy="12" r="10"/>
                  <path d="M9.09 9a3 3 0 015.83 1c0 2-3 3-3 3"/>
                  <line x1="12" y1="17" x2="12.01" y2="17"/>
                </svg>
              </button>
              <a
                href="https://github.com/infamousrusty/tagsha"
                target="_blank" rel="noopener noreferrer"
                className="github-link" aria-label="View TagSha on GitHub"
              >
                <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                  <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
                </svg>
              </a>
            </div>
          </div>
        </header>

        <main className="app-main" id="main-content">
          {state.loadState === 'idle' && (
            <section className="hero" aria-label="Product overview">
              <div className="hero-badge"><span className="dot" aria-hidden="true"/>Open Source · Free to use</div>
              <h1>Resolve Docker Tags<br/>to Git Commits</h1>
              <p>Instantly find the exact commit SHA behind any GHCR or Docker Hub image tag. Verify builds, audit images, and maintain reproducibility.</p>
            </section>
          )}

          <SearchBar
            onSearch={handleSearch}
            isLoading={state.loadState === 'loading'}
            addToast={addToast}
          />

          {state.loadState === 'idle' && (
            <div className="how-it-works" aria-label="How it works">
              {[
                { n: '1', label: 'Select registry' },
                { n: '2', label: 'Enter image ref' },
                { n: '3', label: 'Browse tags' },
                { n: '4', label: 'Copy SHA' },
              ].map((step, i, arr) => (
                <>
                  <div key={step.n} className="hiw-step">
                    <span className="step-num" aria-hidden="true">{step.n}</span>
                    {step.label}
                  </div>
                  {i < arr.length - 1 && <span className="hiw-arrow" aria-hidden="true">→</span>}
                </>
              ))}
            </div>
          )}

          {state.loadState === 'loading' && (
            <div className="status-msg loading" role="status" aria-live="polite">
              <span className="spinner" aria-hidden="true" />
              Fetching tags…
            </div>
          )}

          {state.loadState === 'error' && (() => {
            const meta = errorMeta[state.errorKind ?? 'server']
            return (
              <div className="error-card" role="alert">
                <div className="error-card-icon" aria-hidden="true">{meta.icon}</div>
                <h3>{meta.title}</h3>
                <p>{meta.hint}</p>
                {state.error && <p className="error-corr">Detail: {state.error}</p>}
                <div className="error-actions">
                  <button className="btn btn-primary" onClick={handleRetry}>Retry</button>
                  <button className="btn btn-secondary" onClick={() => setState(s => ({ ...s, loadState: 'idle', error: null }))}>Clear</button>
                </div>
              </div>
            )
          })()}

          {state.loadState === 'success' && state.response && (
            <TagList
              response={state.response}
              cacheStatus={state.cacheStatus}
              onSelectTag={setSelectedTag}
              addToast={addToast}
            />
          )}
        </main>

        <footer className="app-footer">
          <p>
            TagSha — open source ·{' '}
            <a href="https://github.com/infamousrusty/tagsha" target="_blank" rel="noopener noreferrer">GitHub</a>
            {' '}·{' '}
            <a href="/health">Health</a>
            {IS_SELF_HOSTED && <>{' '}· <a href="/settings">Settings</a></>}
          </p>
        </footer>
      </div>

      {selectedTag && (
        <TagDrawer
          tag={selectedTag}
          onClose={() => setSelectedTag(null)}
          addToast={addToast}
        />
      )}

      {showHelp && <HelpModal onClose={() => setShowHelp(false)} />}

      <ToastContainer toasts={toasts} />
    </>
  )
}
