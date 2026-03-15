import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import { SearchBar } from './SearchBar'

const noop = vi.fn()

afterEach(() => { cleanup() })

describe('SearchBar', () => {
  it('renders the input and button', () => {
    render(<SearchBar onSearch={vi.fn()} isLoading={false} addToast={noop} />)
    expect(screen.getByRole('search')).toBeDefined()
    expect(screen.getByRole('button', { name: /search/i })).toBeDefined()
  })

  it('disables submit when input is empty', () => {
    render(<SearchBar onSearch={vi.fn()} isLoading={false} addToast={noop} />)
    const btn = screen.getByRole('button', { name: /search/i }) as HTMLButtonElement
    expect(btn.disabled).toBe(true)
  })

  it('calls onSearch with trimmed input on submit', () => {
    const onSearch = vi.fn()
    render(<SearchBar onSearch={onSearch} isLoading={false} addToast={noop} />)
    const input = screen.getByRole('textbox', { name: /image reference/i }) as HTMLInputElement
    fireEvent.change(input, { target: { value: '  torvalds/linux  ' } })
    fireEvent.submit(screen.getByRole('search'))
    expect(onSearch).toHaveBeenCalledWith('torvalds/linux')
  })

  it('disables input when loading', () => {
    render(<SearchBar onSearch={vi.fn()} isLoading={true} addToast={noop} />)
    const input = screen.getByRole('textbox', { name: /image reference/i }) as HTMLInputElement
    expect(input.disabled).toBe(true)
  })
})
