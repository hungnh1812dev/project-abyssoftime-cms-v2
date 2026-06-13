import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import MockAdapter from 'axios-mock-adapter'
import { api, setAccessToken, getAccessToken } from '@/lib/api'
import { renderWithProviders } from '@/test-utils'
import { AuthProvider } from '@/context/AuthContext'
import { Sidebar } from '@/pages/admin/layout/Sidebar'
import { TopBar } from '@/pages/admin/layout/TopBar'
import type { ContentType } from '@/types/cms'

function makeToken(payload: Record<string, unknown>) {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
  const body = btoa(JSON.stringify(payload))
  return `${header}.${body}.fakesig`
}

const ADMIN_TOKEN = makeToken({ userId: 'u1', role: 'admin', exp: 9999999999 })

const contentTypes: ContentType[] = [
  { ID: '1', DocumentID: 'd1', Name: 'Blog', Slug: 'blog', Kind: 'collection', CreatedAt: '', UpdatedAt: '' },
  { ID: '2', DocumentID: 'd2', Name: 'About', Slug: 'about', Kind: 'single', CreatedAt: '', UpdatedAt: '' },
]

let mock: MockAdapter

beforeEach(() => {
  mock = new MockAdapter(api)
  setAccessToken(null)
})

afterEach(() => {
  mock.restore()
  vi.clearAllMocks()
})

describe('Sidebar', () => {
  it('renders content type names fetched from the API', async () => {
    mock.onGet('/api/content-types').reply(200, contentTypes)
    renderWithProviders(<Sidebar />, { initialEntries: ['/admin'] })
    await waitFor(() => expect(screen.getByText('Blog')).toBeInTheDocument())
    expect(screen.getByText('About')).toBeInTheDocument()
  })

  it('renders nav links pointing to /admin/content-types/:slug', async () => {
    mock.onGet('/api/content-types').reply(200, contentTypes)
    renderWithProviders(<Sidebar />, { initialEntries: ['/admin'] })
    await waitFor(() => expect(screen.getByRole('link', { name: 'Blog' })).toBeInTheDocument())
    expect(screen.getByRole('link', { name: 'Blog' })).toHaveAttribute(
      'href',
      '/admin/content-types/blog',
    )
    expect(screen.getByRole('link', { name: 'About' })).toHaveAttribute(
      'href',
      '/admin/content-types/about',
    )
  })

  it('renders empty state when no content types exist', async () => {
    mock.onGet('/api/content-types').reply(200, [])
    renderWithProviders(<Sidebar />, { initialEntries: ['/admin'] })
    await waitFor(() => expect(screen.queryByRole('link')).toBeNull())
  })
})

describe('TopBar', () => {
  it('renders a Logout button', async () => {
    mock.onPost('/auth/refresh').reply(200, { accessToken: ADMIN_TOKEN })
    renderWithProviders(<AuthProvider><TopBar /></AuthProvider>)
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /logout/i })).toBeInTheDocument(),
    )
  })

  it('clears the access token when Logout is clicked', async () => {
    mock.onPost('/auth/refresh').reply(200, { accessToken: ADMIN_TOKEN })
    mock.onPost('/auth/logout').reply(200)
    const user = userEvent.setup()
    renderWithProviders(<AuthProvider><TopBar /></AuthProvider>)
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /logout/i })).toBeInTheDocument(),
    )
    await user.click(screen.getByRole('button', { name: /logout/i }))
    expect(getAccessToken()).toBeNull()
  })
})
