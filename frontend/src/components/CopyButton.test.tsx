import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, act } from '@testing-library/react'
import { CopyButton } from './CopyButton'

describe('CopyButton', () => {
  beforeEach(() => {
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: vi.fn().mockResolvedValue(undefined) },
      writable: true,
      configurable: true,
    })
  })

  it('renders with default label', () => {
    render(<CopyButton text="abc123" />)
    expect(screen.getByRole('button', { name: /copy/i })).toBeDefined()
  })

  it('shows copied state after click', async () => {
    render(<CopyButton text="abc123def456" label="abc123" />)
    const btn = screen.getByRole('button')
    await act(async () => { fireEvent.click(btn) })
    expect(btn.textContent).toContain('Copied')
  })
})
