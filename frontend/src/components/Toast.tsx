import type { Toast } from '../App'

interface Props {
  toasts: Toast[]
}

const icons: Record<Toast['kind'], string> = {
  success: '✓',
  error: '✕',
  info: 'ℹ',
}

export function ToastContainer({ toasts }: Props) {
  if (toasts.length === 0) return null
  return (
    <div className="toast-container" aria-live="polite" aria-atomic="false">
      {toasts.map(t => (
        <div key={t.id} className={`toast ${t.kind}`} role="status">
          <span className="toast-icon" aria-hidden="true">{icons[t.kind]}</span>
          {t.message}
        </div>
      ))}
    </div>
  )
}
