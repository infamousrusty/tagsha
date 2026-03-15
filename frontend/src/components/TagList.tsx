import React, { useState, useMemo } from 'react'
import type { TagsResponse } from '../types'
import { TagRow } from './TagRow'

interface Props {
  response: TagsResponse
  cacheStatus: string
}

export const TagList: React.FC<Props> = ({ response, cacheStatus }) => {
  const [filter, setFilter] = useState('')

  const filtered = useMemo(
    () =>
      filter
        ? response.tags.filter((t) =>
            t.name.toLowerCase().includes(filter.toLowerCase())
          )
        : response.tags,
    [response.tags, filter]
  )

  const cacheClass = cacheStatus === 'HIT' ? 'cache-hit'
    : cacheStatus === 'STALE' ? 'cache-stale'
    : 'cache-miss'

  return (
    <section className="tag-list-section" aria-label={`Tags for ${response.owner}/${response.repo}`}>
      <div className="tag-list-header">
        <div className="tag-list-title">
          <h2>
            <a
              href={`https://github.com/${response.owner}/${response.repo}`}
              target="_blank"
              rel="noopener noreferrer"
              className="repo-link"
            >
              {response.owner}/{response.repo}
            </a>
          </h2>
          <span className="tag-count">{response.total_count} tags</span>
          {response.truncated && (
            <span className="truncated-badge" title="Results are capped at the configured page limit">
              Truncated
            </span>
          )}
        </div>
        <div className="cache-info">
          <span className={`cache-badge ${cacheClass}`} title={`Data cached at: ${response.cached_at}`}>
            {cacheStatus}
          </span>
          {response.github_rate_limit_remaining > 0 && (
            <span className="rate-limit" title="GitHub API requests remaining">
              API: {response.github_rate_limit_remaining} remaining
            </span>
          )}
        </div>
      </div>

      <div className="tag-filter-row">
        <label htmlFor="tag-filter" className="sr-only">Filter tags</label>
        <input
          id="tag-filter"
          type="search"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          placeholder="Filter tags by name…"
          className="tag-filter-input"
          aria-controls="tag-table"
        />
        <span className="filter-count" aria-live="polite">
          Showing {filtered.length} of {response.total_count}
        </span>
      </div>

      <div className="table-scroll">
        <table id="tag-table" className="tag-table" aria-label="Repository tags">
          <thead>
            <tr>
              <th scope="col">Tag</th>
              <th scope="col">Commit SHA</th>
              <th scope="col">Author</th>
              <th scope="col">Date</th>
              <th scope="col">Message</th>
              <th scope="col"><span className="sr-only">Actions</span></th>
            </tr>
          </thead>
          <tbody>
            {filtered.length > 0 ? (
              filtered.map((tag) => <TagRow key={tag.name} tag={tag} />)
            ) : (
              <tr>
                <td colSpan={6} className="no-results">
                  {filter ? `No tags matching "${filter}"` : 'No tags found'}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </section>
  )
}
