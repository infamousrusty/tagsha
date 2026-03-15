interface Props { onClose: () => void }

const shortcuts = [
  { key: '/', desc: 'Focus search bar' },
  { key: 'Esc', desc: 'Close drawer / modal' },
  { key: '↑ / ↓', desc: 'Navigate tag rows' },
  { key: 'Enter', desc: 'Open tag details' },
  { key: '?', desc: 'Toggle this help dialog' },
]

export function HelpModal({ onClose }: Props) {
  return (
    <div
      className="modal-overlay"
      role="dialog" aria-modal="true" aria-label="Keyboard shortcuts"
      onClick={e => { if (e.target === e.currentTarget) onClose() }}
    >
      <div className="modal">
        <h2>Keyboard shortcuts</h2>
        <ul className="shortcut-list">
          {shortcuts.map(s => (
            <li key={s.key}>
              <span>{s.desc}</span>
              <span className="shortcut-key"><kbd className="kbd">{s.key}</kbd></span>
            </li>
          ))}
        </ul>
        <button className="modal-close-btn" onClick={onClose}>Close</button>
      </div>
    </div>
  )
}
