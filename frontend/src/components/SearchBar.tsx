import { useState, useRef } from 'react'
import type { FC, FormEvent } from 'react'

interface Props {
  onSearch: (query: string) => void
  isLoading: boolean
}

export const SearchBar: FC<Props> = ({ onSearch, isLoading }) => {
  const [value, setValue] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    const trimmed = value.trim()
    if (trimmed) onSearch(trimmed)
  }

  const examples = [
    'torvalds/linux',
    'golang/go',
    'rust-lang/rust',
    'https://github.com/nodejs/node',
  ]

  return (
    <div className="search-wrapper">
      <form onSubmit={handleSubmit} className="search-form" role="search">
        <label htmlFor="repo-search" className="sr-only">
          GitHub repository identifier
        </label>
        <input
          id="repo-search"
          ref={inputRef}
          type="text"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder="owner/repo or https://github.com/owner/repo"
          maxLength={256}
          autoComplete="off"
          autoCorrect="off"
          autoCapitalize="off"
          spellCheck={false}
          disabled={isLoading}
          className="search-input"
          aria-label="GitHub repository identifier"
          aria-describedby="search-examples"
        />
        <button
          type="submit"
          disabled={isLoading || !value.trim()}
          className="search-btn"
          aria-busy={isLoading}
        >
          {isLoading ? 'Resolving…' : 'Lookup Tags'}
        </button>
      </form>
      <p id="search-examples" className="search-examples">
        {examples.map((ex) => (
          <button
            key={ex}
            className="example-btn"
            onClick={() => { setValue(ex); onSearch(ex) }}
            type="button"
            aria-label={`Search for ${ex}`}
          >
            {ex}
          </button>
        ))}
      </p>
    </div>
  )
}
