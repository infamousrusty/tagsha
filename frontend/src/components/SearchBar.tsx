import { useRef, useCallback, useState, useEffect } from 'react'
import type { Registry } from '../types'
import type { Toast } from '../App'
import type { JSX } from 'react';

const EXAMPLES: Record<Registry, string[]> = {
  ghcr:      ['ghcr.io/infamousrusty/tagsha:latest', 'ghcr.io/cli/cli:v2.45.0'],
  dockerhub: ['docker.io/library/nginx:1.25', 'docker.io/library/redis:7'],
  other:     [],
}

const REGISTRY_LABELS: { id: Registry; label: string; icon: JSX.Element }[] = [
  {
    id: 'ghcr',
    label: 'GHCR',
    icon: (
      <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
        <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/>
      </svg>
    ),
  },
  {
    id: 'dockerhub',
    label: 'Docker Hub',
    icon: (
      <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
        <path d="M13.983 11.078h2.119a.186.186 0 00.186-.185V9.006a.186.186 0 00-.186-.186h-2.119a.185.185 0 00-.185.185v1.888c0 .102.083.185.185.185m-2.954-5.43h2.118a.186.186 0 00.186-.186V3.574a.186.186 0 00-.186-.185h-2.118a.185.185 0 00-.185.185v1.888c0 .102.082.185.185.186m0 2.716h2.118a.187.187 0 00.186-.186V6.29a.186.186 0 00-.186-.185h-2.118a.185.185 0 00-.185.185v1.887c0 .102.082.185.185.186m-2.93 0h2.12a.186.186 0 00.184-.186V6.29a.185.185 0 00-.185-.185H8.1a.185.185 0 00-.185.185v1.887c0 .102.083.185.185.186m-2.964 0h2.119a.186.186 0 00.185-.186V6.29a.185.185 0 00-.185-.185H5.136a.186.186 0 00-.186.185v1.887c0 .102.084.185.186.186m5.893 2.715h2.118a.186.186 0 00.186-.185V9.006a.186.186 0 00-.186-.186h-2.118a.185.185 0 00-.185.185v1.888c0 .102.082.185.185.185m-2.93 0h2.12a.185.185 0 00.184-.185V9.006a.185.185 0 00-.185-.186H8.1a.185.185 0 00-.185.185v1.888c0 .102.083.185.185.185m-2.964 0h2.119a.185.185 0 00.185-.185V9.006a.185.185 0 00-.185-.186H5.136a.186.186 0 00-.186.185v1.888c0 .102.084.185.186.185m-2.92 0h2.12a.185.185 0 00.184-.185V9.006a.185.185 0 00-.184-.186H2.217a.186.186 0 00-.185.185v1.888c0 .102.083.185.185.185M23.763 9.89c-.065-.051-.672-.51-1.954-.51-.338.001-.676.03-1.01.087-.248-1.7-1.653-2.53-1.716-2.566l-.344-.199-.226.327c-.284.438-.49.922-.612 1.43-.23.97-.09 1.882.403 2.661-.595.332-1.55.413-1.744.42H.751a.751.751 0 00-.75.748 11.376 11.376 0 00.692 4.062c.545 1.428 1.355 2.48 2.41 3.124 1.18.723 3.1 1.137 5.275 1.137.983.003 1.963-.086 2.93-.266a12.248 12.248 0 003.823-1.389c.98-.567 1.86-1.288 2.61-2.136 1.252-1.418 1.998-2.997 2.553-4.4h.221c1.372 0 2.215-.549 2.68-1.009.309-.293.55-.65.707-1.046l.098-.288z"/>
      </svg>
    ),
  },
]

interface Props {
  onSearch: (query: string) => void
  isLoading: boolean
  addToast: (msg: string, kind?: Toast['kind']) => void
}

export function SearchBar({ onSearch, isLoading, addToast }: Props) {
  const [value, setValue] = useState('')
  const [registry, setRegistry] = useState<Registry>('ghcr')
  const inputRef = useRef<HTMLInputElement>(null)

  // Register global `/` shortcut to focus the search input
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === '/' && !['INPUT', 'TEXTAREA'].includes((e.target as HTMLElement).tagName)) {
        e.preventDefault()
        inputRef.current?.focus()
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  const handleSubmit = useCallback((e: React.FormEvent) => {
    e.preventDefault()
    const trimmed = value.trim()
    if (!trimmed) { addToast('Please enter an image reference', 'error'); return }
    onSearch(trimmed)
  }, [value, onSearch, addToast])

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') handleSubmit(e as unknown as React.FormEvent)
  }

  const fillExample = (ex: string) => {
    setValue(ex)
    inputRef.current?.focus()
  }

  const examples = EXAMPLES[registry]
  const placeholder = registry === 'ghcr'
    ? 'ghcr.io/owner/image:tag or owner/repo'
    : 'docker.io/library/image:tag or image:tag'

  return (
    <section className="search-section" aria-label="Image search">
      <div className="registry-tabs" role="tablist" aria-label="Select registry">
        {REGISTRY_LABELS.map(r => (
          <button
            key={r.id}
            role="tab"
            aria-selected={registry === r.id}
            className={`registry-tab${registry === r.id ? ' active' : ''}`}
            onClick={() => setRegistry(r.id)}
          >
            {r.icon}
            {r.label}
          </button>
        ))}
      </div>

      <div className="search-wrapper">
        <form className="search-form" onSubmit={handleSubmit} role="search">
          <label htmlFor="image-search" className="sr-only">Image reference</label>
          <input
            id="image-search"
            ref={inputRef}
            className="search-input"
            type="text"
            value={value}
            onChange={e => setValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={isLoading}
            autoComplete="off"
            spellCheck={false}
            aria-label="Image reference"
          />
          <button className="search-btn" type="submit" disabled={isLoading || !value.trim()}>
            {isLoading ? (
              <><span className="spinner" aria-hidden="true" style={{ width: 15, height: 15, borderWidth: 2 }} />Searching…</>
            ) : (
              <>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                  <circle cx="11" cy="11" r="8"/>
                  <line x1="21" y1="21" x2="16.65" y2="16.65"/>
                </svg>
                Search
              </>
            )}
          </button>
        </form>
      </div>

      <div className="search-meta">
        {examples.length > 0 && (
          <div className="search-examples">
            <span>Try:</span>
            {examples.map(ex => (
              <button key={ex} className="example-btn" type="button" onClick={() => fillExample(ex)}>{ex}</button>
            ))}
          </div>
        )}
        <span className="kbd-hint">Press <kbd className="kbd">/</kbd> to focus</span>
      </div>
    </section>
  )
}
