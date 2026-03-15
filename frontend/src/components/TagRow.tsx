import React from 'react'
import type { Tag } from '../types'
import { CopyButton } from './CopyButton'

interface Props {
  tag: Tag
}

function formatDate(iso: string): string {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleDateString('en-GB', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  } catch {
    return iso
  }
}

export const TagRow: React.FC<Props> = ({ tag }) => {
  const shortSHA = tag.sha ? tag.sha.slice(0, 12) : '—'

  return (
    <tr className="tag-row">
      <td className="tag-name">
        <span className="tag-badge">{tag.name}</span>
      </td>
      <td className="tag-sha">
        <code className="sha-code" title={tag.sha}>{shortSHA}</code>
        {tag.sha && (
          <CopyButton text={tag.sha} label={shortSHA} />
        )}
      </td>
      <td className="tag-author">
        {/* Rendered as plain text, never as HTML — XSS safe */}
        {tag.author_name || '—'}
      </td>
      <td className="tag-date">{formatDate(tag.date)}</td>
      <td className="tag-message">
        {tag.message
          ? <span title={tag.message}>{tag.message.slice(0, 72)}{tag.message.length > 72 ? '…' : ''}</span>
          : '—'
        }
      </td>
      <td className="tag-actions">
        {tag.commit_url && (
          <a
            href={tag.commit_url}
            target="_blank"
            rel="noopener noreferrer"
            className="view-link"
            aria-label={`View commit for tag ${tag.name} on GitHub`}
          >
            GitHub ↗
          </a>
        )}
      </td>
    </tr>
  )
}
