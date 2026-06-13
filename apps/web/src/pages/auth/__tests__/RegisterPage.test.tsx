import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import MockAdapter from 'axios-mock-adapter'
import { api } from '@/lib/api'
import { renderWithProviders } from '@/test-utils'
import { RegisterPage } from '@/pages/auth/RegisterPage'

let mock: MockAdapter

beforeEach(() => {
  mock = new MockAdapter(api)
})

afterEach(() => {
  mock.restore()
  vi.clearAllMocks()
})

describe('RegisterPage', () => {
  it('renders email and password fields with a submit button', () => {
    renderWithProviders(<RegisterPage />)
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /create account/i })).toBeInTheDocument()
  })

  it('shows validation error for invalid email', async () => {
    const user = userEvent.setup()
    renderWithProviders(<RegisterPage />)

    await user.type(screen.getByLabelText(/email/i), 'bad-email')
    await user.type(screen.getByLabelText(/password/i), 'password123')
    await user.click(screen.getByRole('button', { name: /create account/i }))

    await waitFor(() => {
      expect(screen.getByText(/valid email/i)).toBeInTheDocument()
    })
  })

  it('shows validation error for password shorter than 8 characters', async () => {
    const user = userEvent.setup()
    renderWithProviders(<RegisterPage />)

    await user.type(screen.getByLabelText(/email/i), 'user@example.com')
    await user.type(screen.getByLabelText(/password/i), 'short')
    await user.click(screen.getByRole('button', { name: /create account/i }))

    await waitFor(() => {
      expect(screen.getByText(/at least 8/i)).toBeInTheDocument()
    })
  })

  it('calls POST /auth/register on valid submit', async () => {
    const user = userEvent.setup()
    let capturedBody: unknown
    mock.onPost('/auth/register').reply((config) => {
      capturedBody = JSON.parse(config.data)
      return [201, { id: 'user-1' }]
    })
    renderWithProviders(<RegisterPage />)

    await user.type(screen.getByLabelText(/email/i), 'newuser@example.com')
    await user.type(screen.getByLabelText(/password/i), 'securepass')
    await user.click(screen.getByRole('button', { name: /create account/i }))

    await waitFor(() => {
      expect(capturedBody).toEqual({ email: 'newuser@example.com', password: 'securepass' })
    })
  })

  it('shows error message when registration fails', async () => {
    const user = userEvent.setup()
    mock.onPost('/auth/register').reply(409, { message: 'Email already exists' })
    renderWithProviders(<RegisterPage />)

    await user.type(screen.getByLabelText(/email/i), 'taken@example.com')
    await user.type(screen.getByLabelText(/password/i), 'password123')
    await user.click(screen.getByRole('button', { name: /create account/i }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })
})
