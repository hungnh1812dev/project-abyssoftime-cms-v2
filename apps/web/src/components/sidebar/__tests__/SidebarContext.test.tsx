import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SidebarProvider, useSidebar } from '../SidebarContext'

function TestConsumer() {
  const { collapsed, toggle, isMobile, mobileOpen, setMobileOpen } = useSidebar()
  return (
    <div>
      <span data-testid="collapsed">{String(collapsed)}</span>
      <span data-testid="is-mobile">{String(isMobile)}</span>
      <span data-testid="mobile-open">{String(mobileOpen)}</span>
      <button onClick={toggle}>Toggle</button>
      <button onClick={() => setMobileOpen(true)}>Open Mobile</button>
      <button onClick={() => setMobileOpen(false)}>Close Mobile</button>
    </div>
  )
}

beforeEach(() => {
  localStorage.clear()
})

describe('SidebarContext', () => {
  it('provides default collapsed=false', () => {
    render(
      <SidebarProvider>
        <TestConsumer />
      </SidebarProvider>,
    )
    expect(screen.getByTestId('collapsed')).toHaveTextContent('false')
  })

  it('toggles collapsed state', async () => {
    render(
      <SidebarProvider>
        <TestConsumer />
      </SidebarProvider>,
    )
    await userEvent.click(screen.getByRole('button', { name: 'Toggle' }))
    expect(screen.getByTestId('collapsed')).toHaveTextContent('true')
  })

  it('persists collapsed state to localStorage', async () => {
    render(
      <SidebarProvider>
        <TestConsumer />
      </SidebarProvider>,
    )
    await userEvent.click(screen.getByRole('button', { name: 'Toggle' }))
    expect(localStorage.getItem('sidebar-collapsed')).toBe('true')
  })

  it('reads initial collapsed state from localStorage', () => {
    localStorage.setItem('sidebar-collapsed', 'true')
    render(
      <SidebarProvider>
        <TestConsumer />
      </SidebarProvider>,
    )
    expect(screen.getByTestId('collapsed')).toHaveTextContent('true')
  })

  it('provides mobileOpen state and setter', async () => {
    render(
      <SidebarProvider>
        <TestConsumer />
      </SidebarProvider>,
    )
    expect(screen.getByTestId('mobile-open')).toHaveTextContent('false')
    await userEvent.click(screen.getByRole('button', { name: 'Open Mobile' }))
    expect(screen.getByTestId('mobile-open')).toHaveTextContent('true')
    await userEvent.click(screen.getByRole('button', { name: 'Close Mobile' }))
    expect(screen.getByTestId('mobile-open')).toHaveTextContent('false')
  })

  it('detects isMobile from matchMedia', () => {
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockImplementation((query: string) => ({
        matches: query === '(max-width: 1023px)',
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    })

    render(
      <SidebarProvider>
        <TestConsumer />
      </SidebarProvider>,
    )
    expect(screen.getByTestId('is-mobile')).toHaveTextContent('true')
  })
})
