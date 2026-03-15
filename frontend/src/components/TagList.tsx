import { useState, useCallback, useRef } from 'react'
import type { TagsResponse, TagEntry } from '../types'
import type { Toast } from '../App'

interface Props {
  response: TagsResponse
  cacheStatus: string
  onSelectTag: (tag: TagEntry) => void
  addToast: (msg: string, kind?: Toast['kind']) => void
}

function formatDate(iso: string) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}

type SortKey = 'default' | 'date_desc' | 'date_asc' | 'semver' | 'alpha'

function sortTags(tags: TagEntry[], key: SortKey): TagEntry[] {
  const copy = [...tags]
  if (key === 'date_desc') return copy.sort((a, b) => (b.date ?? '').localeCompare(a.date ?? ''))
  if (key === 'date_asc')  return copy.sort((a, b) => (a.date ?? '').localeCompare(b.date ?? ''))
  if (key === 'alpha')     return copy.sort((a, b) => a.name.localeCompare(b.name))
  return copy
}

export function TagList({ response, cacheStatus, onSelectTag, addToast }: Props) {
  const [filter, setFilter] = useState('')
  const [sort, setSort] = useState<SortKey>('default')
  const rowRefs = useRef<(HTMLTableRowElement | null)[]>([])

  const repoUrl = `https://github.com/${response.owner}/${response.repo}`
  const cs = cacheStatus.toLowerCase()
  const cacheClass = cs.includes('hit') ? 'cache-hit' : cs.includes('stale') ? 'cache-stale' : 'cache-miss'

  const rawEntries: TagEntry[] = response.tags.map(t => ({
    name: t.name,
    sha: t.sha,
    message: t.message,
    author: t.author_name,
    date: t.date,
    commitUrl: t.commit_url,
    repo: `github.com/${response.owner}/${response.repo}`,
  }))

  const filtered = sortTags(
    rawEntries.filter(t => t.name.toLowerCase().includes(filter.toLowerCase())),
    sort,
  )

  const copyValue = useCallback(async (value: string, label: string, e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      await navigator.clipboard.writeText(value)
      addToast(`${label} copied!`, 'success')
    } catch {
      addToast('Copy failed', 'error')
    }
  }, [addToast])

  // Keyboard navigation for rows
  const handleRowKey = useCallback((e: React.KeyboardEvent, tag: TagEntry, idx: number) => {
    if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelectTag(tag) }
    if (e.key === 'ArrowDown') { e.preventDefault(); rowRefs.current[idx + 1]?.focus() }
    if (e.key === 'ArrowUp')   { e.preventDefault(); rowRefs.current[idx - 1]?.focus() }
  }, [onSelectTag])

  return (
    <section className="tag-list-section" aria-label="Tag results">
      <div className="tag-list-header">
        <div className="tag-list-title">
          <h2>
            <a href={repoUrl} target="_blank" rel="noopener noreferrer" className="repo-link">
              {response.owner}/{response.repo}
            </a>
          </h2>
          <span className="tag-count">{response.total_count} tags</span>
          {response.truncated && <span className="truncated-badge">Truncated</span>}
        </div>
        <div className="cache-info">
          {cacheStatus && <span className={`cache-badge ${cacheClass}`}>{cacheStatus}</span>}
          {response.github_rate_limit_remaining != null && (
            <span className="rate-limit">Rate limit: {response.github_rate_limit_remaining} remaining</span>
          )}
        </div>
      </div>

      <div className="tag-filter-row">
        <label htmlFor="tag-filter" className="sr-only">Filter tags</label>
        <input
          id="tag-filter"
          className="tag-filter-input"
          type="text"
          placeholder="Filter tags…"
          value={filter}
          onChange={e => setFilter(e.target.value)}
          aria-label="Filter tags"
        />
        <label htmlFor="tag-sort" className="sr-only">Sort tags</label>
        <select
          id="tag-sort"
          className="sort-select"
          value={sort}
          onChange={e => setSort(e.target.value as SortKey)}
          aria-label="Sort tags"
        >
          <option value="default">Default order</option>
          <option value="date_desc">Newest first</option>
          <option value="date_asc">Oldest first</option>
          <option value="alpha">A → Z</option>
        </select>
        <span className="filter-count" aria-live="polite">
          {filter ? `${filtered.length} of ${rawEntries.length}` : ''}
        </span>
      </div>

      <div className="table-scroll">
        <table className="tag-table" role="grid" aria-label="Tag list">
          <thead>
            <tr>
              <th scope="col">Tag</th>
              <th scope="col">Commit SHA</th>
              <th scope="col" className="tag-message">Message</th>
              <th scope="col" className="tag-author">Author</th>
              <th scope="col">Date</th>
              <th scope="col"><span className="sr-only">Actions</span></th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr><td colSpan={6} className="no-results">No tags match your filter.</td></tr>
            ) : filtered.map((tag, idx) => (
              <tr
                key={tag.name}
                ref={el => { rowRefs.current[idx] = el }}
                tabIndex={0}
                onClick={() => onSelectTag(tag)}
                onKeyDown={e => handleRowKey(e, tag, idx)}
                aria-label={`Tag ${tag.name}, SHA ${tag.sha ? tag.sha.slice(0, 7) : 'unknown'}`}
              >
                <td><span className="tag-badge">{tag.name}</span></td>
                <td>
                  {tag.sha ? (
                    <>
                      <code className="sha-code">{tag.sha.slice(0, 7)}</code>
                      <button
                        className="copy-btn"
                        onClick={e => copyValue(tag.sha!, 'SHA', e)}
                        aria-label={`Copy SHA for ${tag.name}`}
                      >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                          <rect x="9" y="9" width="13" height="13" rx="2"/>
                          <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
                        </svg>
                      </button>
                    </>
                  ) : <span style={{ color: 'var(--text-faint)' }}>—</span>}
                </td>
                <td className="tag-message">{tag.message || '—'}</td>
                <td className="tag-author">{tag.author || '—'}</td>
                <td className="tag-date">{formatDate(tag.date ?? '')}</td>
                <td>
                  <span className="open-drawer-hint">View →</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
