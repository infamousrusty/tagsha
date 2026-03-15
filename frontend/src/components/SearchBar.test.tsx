import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { SearchBar } from './SearchBar'

// Minimal setup for @testing-library/react if not globally configured.
// In CI this runs via vitest with jsdom environment.
describe('SearchBar', () => {
  it('renders the input and button', () => {
    render(<SearchBar onSearch={vi.fn()} isLoading={false} />)
    expect(screen.getByRole('search')).toBeDefined()
    expect(screen.getByRole('button', { name: /lookup tags/i })).toBeDefined()
  })

  it('disables submit when input is empty', () => {
    render(<SearchBar onSearch={vi.fn()} isLoading={false} />)
    const btn = screen.getByRole('button', { name: /lookup tags/i }) as HTMLButtonElement
    expect(btn.disabled).toBe(true)
  })

  it('calls onSearch with trimmed input on submit', () => {
    const onSearch = vi.fn()
    render(<SearchBar onSearch={onSearch} isLoading={false} />)
    const input = screen.getByRole('searchbox') as HTMLInputElement
    fireEvent.change(input, { target: { value: '  torvalds/linux  ' } })
    fireEvent.submit(screen.getByRole('search'))
    expect(onSearch).toHaveBeenCalledWith('torvalds/linux')
  })

  it('disables input when loading', () => {
    render(<SearchBar onSearch={vi.fn()} isLoading={true} />)
    const input = screen.getByRole('searchbox') as HTMLInputElement
    expect(input.disabled).toBe(true)
  })
})
