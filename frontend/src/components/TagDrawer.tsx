import { useEffect, useRef } from 'react'
import type { TagEntry } from '../types'
import type { Toast } from '../App'

interface Props {
  tag: TagEntry
  onClose: () => void
  addToast: (msg: string, kind?: Toast['kind']) => void
}

function CopyBtn({ value, label }: { value: string; label: string }) {
  const copy = async () => {
    await navigator.clipboard.writeText(value)
  }
  return (
    <button className="copy-btn" onClick={copy} aria-label={`Copy ${label}`}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
        <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
      </svg>
      Copy
    </button>
  )
}

export function TagDrawer({ tag, onClose, addToast }: Props) {
  const closeRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    closeRef.current?.focus()
    document.body.style.overflow = 'hidden'
    return () => { document.body.style.overflow = '' }
  }, [])

  const fullRef = tag.repo ? `${tag.repo}:${tag.name}` : tag.name
  const cliCmd = `tagsha ${fullRef}`

  const copyValue = async (value: string, label: string) => {
    try {
      await navigator.clipboard.writeText(value)
      addToast(`${label} copied!`, 'success')
    } catch {
      addToast('Copy failed', 'error')
    }
  }

  return (
    <>
      <div
        className="drawer-overlay"
        onClick={onClose}
        aria-hidden="true"
      />
      <aside
        className="drawer"
        role="dialog"
        aria-modal="true"
        aria-label={`Tag details: ${tag.name}`}
      >
        <div className="drawer-header">
          <div>
            {tag.repo && <div className="drawer-image-ref">{tag.repo}</div>}
            <div className="drawer-title">{tag.name}</div>
          </div>
          <button ref={closeRef} className="drawer-close" onClick={onClose} aria-label="Close tag details">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>

        <div className="drawer-body">
          {/* Commit SHA */}
          <div className="drawer-section">
            <div className="drawer-section-title">Commit SHA</div>
            {tag.sha ? (
              <div className="sha-block">
                <code className="sha-value">{tag.sha}</code>
                <button
                  className="copy-btn"
                  onClick={() => copyValue(tag.sha!, 'SHA')}
                  aria-label="Copy commit SHA"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                    <rect x="9" y="9" width="13" height="13" rx="2"/>
                    <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
                  </svg>
                  Copy
                </button>
              </div>
            ) : (
              <p style={{ color: 'var(--text-muted)', fontSize: '0.88rem' }}>SHA not resolved for this tag.</p>
            )}
          </div>

          {/* Commit metadata */}
          {(tag.message || tag.author || tag.date || tag.commitUrl) && (
            <div className="drawer-section">
              <div className="drawer-section-title">Commit details</div>
              <ul className="meta-list">
                {tag.message && (
                  <li><span className="meta-key">Message</span><span className="meta-val">{tag.message}</span></li>
                )}
                {tag.author && (
                  <li><span className="meta-key">Author</span><span className="meta-val">{tag.author}</span></li>
                )}
                {tag.date && (
                  <li><span className="meta-key">Date</span><span className="meta-val">{new Date(tag.date).toLocaleString()}</span></li>
                )}
                {tag.commitUrl && (
                  <li>
                    <span className="meta-key">Commit</span>
                    <span className="meta-val">
                      <a href={tag.commitUrl} target="_blank" rel="noopener noreferrer">View on GitHub ↗</a>
                    </span>
                  </li>
                )}
              </ul>
            </div>
          )}

          {/* Image metadata */}
          {(tag.digest || tag.size) && (
            <div className="drawer-section">
              <div className="drawer-section-title">Image metadata</div>
              <ul className="meta-list">
                {tag.digest && (
                  <li><span className="meta-key">Digest</span><span className="meta-val"><code>{tag.digest}</code></span></li>
                )}
                {tag.size && (
                  <li><span className="meta-key">Size</span><span className="meta-val">{tag.size}</span></li>
                )}
              </ul>
            </div>
          )}

          {/* CLI equivalent */}
          <div className="drawer-section">
            <div className="drawer-section-title">CLI equivalent</div>
            <div className="cli-snippet-block">
              <code className="cli-snippet-cmd">
                <span className="cli-prefix">$ </span>{cliCmd}
              </code>
              <button
                className="copy-btn"
                onClick={() => copyValue(cliCmd, 'CLI command')}
                aria-label="Copy CLI command"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                  <rect x="9" y="9" width="13" height="13" rx="2"/>
                  <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
                </svg>
                Copy
              </button>
            </div>
          </div>
        </div>

        <div className="drawer-actions">
          {tag.sha && (
            <button className="btn btn-primary" onClick={() => copyValue(tag.sha!, 'SHA')}>
              Copy SHA
            </button>
          )}
          <button className="btn btn-secondary" onClick={() => copyValue(fullRef, 'image reference')}>
            Copy ref
          </button>
          {tag.commitUrl && (
            <a href={tag.commitUrl} target="_blank" rel="noopener noreferrer" className="btn btn-secondary">
              View on GitHub ↗
            </a>
          )}
          <button className="btn btn-secondary" onClick={onClose} style={{ marginLeft: 'auto' }}>
            Close
          </button>
        </div>
      </aside>
    </>
  )
}
