import { useState } from 'react'

interface Props {
  text: string
  label?: string
}

export const CopyButton: React.FC<Props> = ({ text, label = 'Copy' }) => {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text)
    } catch {
      // Fallback for non-secure contexts.
      const el = document.createElement('textarea')
      el.value = text
      el.style.position = 'absolute'
      el.style.left = '-9999px'
      document.body.appendChild(el)
      el.select()
      document.execCommand('copy')
      document.body.removeChild(el)
    }
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <button
      onClick={handleCopy}
      className={`copy-btn${copied ? ' copied' : ''}`}
      aria-label={copied ? 'Copied!' : `Copy ${label}`}
      title={copied ? 'Copied to clipboard' : `Copy ${label}`}
      type="button"
    >
      {copied ? '✓ Copied' : label}
    </button>
  )
}
