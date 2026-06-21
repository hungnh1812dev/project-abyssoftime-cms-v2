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
  mock.onGet('/auth/setup').reply(200, { adminExists: false })
})

afterEach(() => {
  mock.restore()
  vi.clearAllMocks()
})

describe('RegisterPage', () => {
  it('renders email and password fields with a submit button', async () => {
    renderWithProviders(<RegisterPage />)
    expect(await screen.findByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /create admin account/i })).toBeInTheDocument()
  })

  it('shows validation error for invalid email', async () => {
    const user = userEvent.setup()
    renderWithProviders(<RegisterPage />)

    const emailInput = await screen.findByLabelText(/email/i)
    await user.type(emailInput, 'bad-email')
    await user.type(screen.getByLabelText(/password/i), 'password123')
    await user.click(screen.getByRole('button', { name: /create admin account/i }))

    await waitFor(() => {
      expect(screen.getByText(/valid email/i)).toBeInTheDocument()
    })
  })

  it('shows validation error for password shorter than 8 characters', async () => {
    const user = userEvent.setup()
    renderWithProviders(<RegisterPage />)

    const emailInput = await screen.findByLabelText(/email/i)
    await user.type(emailInput, 'user@example.com')
    await user.type(screen.getByLabelText(/password/i), 'short')
    await user.click(screen.getByRole('button', { name: /create admin account/i }))

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

    const emailInput = await screen.findByLabelText(/email/i)
    await user.type(emailInput, 'newuser@example.com')
    await user.type(screen.getByLabelText(/password/i), 'securepass')
    await user.click(screen.getByRole('button', { name: /create admin account/i }))

    await waitFor(() => {
      expect(capturedBody).toEqual({ email: 'newuser@example.com', password: 'securepass' })
    })
  })

  it('redirects to /login when admin already exists', async () => {
    mock.onGet('/auth/setup').reply(200, { adminExists: true })
    renderWithProviders(<RegisterPage />)

    await waitFor(() => {
      expect(screen.queryByLabelText(/email/i)).not.toBeInTheDocument()
      expect(screen.queryByLabelText(/password/i)).not.toBeInTheDocument()
    })
  })

  it('shows error message when registration fails', async () => {
    const user = userEvent.setup()
    mock.onPost('/auth/register').reply(409, { message: 'Email already exists' })
    renderWithProviders(<RegisterPage />)

    const emailInput = await screen.findByLabelText(/email/i)
    await user.type(emailInput, 'taken@example.com')
    await user.type(screen.getByLabelText(/password/i), 'password123')
    await user.click(screen.getByRole('button', { name: /create admin account/i }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })
})
