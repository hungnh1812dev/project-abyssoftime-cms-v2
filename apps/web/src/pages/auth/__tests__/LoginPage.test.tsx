import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import MockAdapter from 'axios-mock-adapter'
import { api, getAccessToken, setAccessToken } from '@/lib/api'
import { renderWithProviders } from '@/test-utils'
import { LoginPage } from '@/pages/auth/LoginPage'

let mock: MockAdapter

beforeEach(() => {
  mock = new MockAdapter(api)
  setAccessToken(null)
})

afterEach(() => {
  mock.restore()
  vi.clearAllMocks()
})

describe('LoginPage', () => {
  it('renders email and password fields with a submit button', () => {
    renderWithProviders(<LoginPage />)
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('shows validation error for invalid email', async () => {
    const user = userEvent.setup()
    renderWithProviders(<LoginPage />)

    await user.type(screen.getByLabelText(/email/i), 'not-an-email')
    await user.type(screen.getByLabelText(/password/i), 'password123')
    await user.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => {
      expect(screen.getByText(/valid email/i)).toBeInTheDocument()
    })
  })

  it('shows validation error for password shorter than 8 characters', async () => {
    const user = userEvent.setup()
    renderWithProviders(<LoginPage />)

    await user.type(screen.getByLabelText(/email/i), 'user@example.com')
    await user.type(screen.getByLabelText(/password/i), 'short')
    await user.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => {
      expect(screen.getByText(/at least 8/i)).toBeInTheDocument()
    })
  })

  it('calls POST /auth/login and stores token on success', async () => {
    const user = userEvent.setup()
    mock.onPost('/auth/login').reply(200, { accessToken: 'tok-123' })
    renderWithProviders(<LoginPage />)

    await user.type(screen.getByLabelText(/email/i), 'user@example.com')
    await user.type(screen.getByLabelText(/password/i), 'password123')
    await user.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => {
      expect(getAccessToken()).toBe('tok-123')
    })
  })

  it('shows error message when login fails', async () => {
    const user = userEvent.setup()
    mock.onPost('/auth/login').reply(401, { message: 'Invalid credentials' })
    renderWithProviders(<LoginPage />)

    await user.type(screen.getByLabelText(/email/i), 'user@example.com')
    await user.type(screen.getByLabelText(/password/i), 'password123')
    await user.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })
})
