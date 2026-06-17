import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MockAdapter from 'axios-mock-adapter'
import { api } from '@/lib/api'
import { SingleTypePage } from '../SingleTypePage'
import type { ContentType, Document } from '@/types/cms'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

const ct: ContentType = {
  ID: 'ct-1',
  DocumentID: 'ct-doc-1',
  Name: 'Homepage',
  Slug: 'homepage',
  Kind: 'single',
  CreatedAt: '',
  UpdatedAt: '',
}

const doc: Document = {
  EntryID: 'doc-1',
  ContentTypeID: 'ct-1',
  Status: 'draft',
  Data: { title: 'Hello World', body: 'Some text' },
  Locale: 'en',
  CreatedAt: '',
  UpdatedAt: '',
  CreatedBy: '',
  UpdatedBy: '',
}

let mock: MockAdapter

function createClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
}

function renderPage(initialPath = '/content-type/single-type/homepage') {
  return render(
    <QueryClientProvider client={createClient()}>
      <MemoryRouter initialEntries={[initialPath]}>
        <Routes>
          <Route path="/content-type/single-type/:slug" element={<SingleTypePage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

beforeEach(() => {
  mock = new MockAdapter(api)
  mock.onGet('/api/locales').reply(200, ['en'])
  mock.onGet('/api/content-types').reply(200, [ct])
  mock.onGet('/api/documents').reply(200, [doc])
  mock.onGet('/api/documents/doc-1').reply(200, doc)
})

afterEach(() => {
  mock.restore()
  vi.clearAllMocks()
})

describe('SingleTypePage', () => {
  it('renders the content type name as title', async () => {
    renderPage()
    await waitFor(() => expect(screen.getByText('Homepage')).toBeInTheDocument())
  })

  it('renders form fields pre-filled from document data', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByLabelText('title')).toHaveValue('Hello World')
      expect(screen.getByLabelText('body')).toHaveValue('Some text')
    })
  })

  it('Save button is disabled on initial load (form is clean)', async () => {
    renderPage()
    await waitFor(() => expect(screen.getByRole('button', { name: /save/i })).toBeDisabled())
  })

  it('Save button becomes enabled after editing a field', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => screen.getByLabelText('title'))
    await user.clear(screen.getByLabelText('title'))
    await user.type(screen.getByLabelText('title'), 'New title')
    expect(screen.getByRole('button', { name: /save/i })).not.toBeDisabled()
  })

  it('shows success toast and resets form to clean after successful save', async () => {
    const { toast } = await import('sonner')
    const user = userEvent.setup()
    mock.onPut('/api/documents/doc-1').reply(200, { ...doc, Data: { title: 'New title', body: 'Some text' } })

    renderPage()
    await waitFor(() => screen.getByLabelText('title'))
    await user.clear(screen.getByLabelText('title'))
    await user.type(screen.getByLabelText('title'), 'New title')
    expect(screen.getByRole('button', { name: /save/i })).not.toBeDisabled()

    await user.click(screen.getByRole('button', { name: /save/i }))

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith('Saved')
      expect(screen.getByRole('button', { name: /save/i })).toBeDisabled()
    })
  })

  it('shows error toast and preserves values on failed save', async () => {
    const { toast } = await import('sonner')
    const user = userEvent.setup()
    mock.onPut('/api/documents/doc-1').reply(422, { error: 'Validation failed' })

    renderPage()
    // Wait for the form to be pre-filled before interacting
    await waitFor(() => expect(screen.getByLabelText('title')).toHaveValue('Hello World'))
    const titleInput = screen.getByLabelText('title')
    await user.clear(titleInput)
    await user.type(titleInput, 'bad')

    await user.click(screen.getByRole('button', { name: /save/i }))

    await waitFor(() => expect(toast.error).toHaveBeenCalledWith('Validation failed'))
    expect(screen.getByLabelText('title')).toHaveValue('bad')
  })

  it('does not show locale selector when only one locale is available', async () => {
    renderPage()
    await waitFor(() => screen.getByText('Homepage'))
    expect(screen.queryByRole('combobox', { name: /locale/i })).not.toBeInTheDocument()
  })

  it('shows locale selector when more than one locale is available', async () => {
    mock.onGet('/api/locales').reply(200, ['en', 'vi'])
    renderPage()
    await waitFor(() =>
      expect(screen.getByRole('combobox', { name: /locale/i })).toBeInTheDocument(),
    )
  })

  it('switching locale sends a new document fetch for the new locale', async () => {
    const user = userEvent.setup()
    mock.onGet('/api/locales').reply(200, ['en', 'vi'])
    mock.onGet('/api/documents/doc-1').reply(200, { ...doc, Locale: 'vi', Data: { title: 'Xin chào' } })

    renderPage()
    await waitFor(() => screen.getByRole('combobox', { name: /locale/i }))
    await user.selectOptions(screen.getByRole('combobox', { name: /locale/i }), 'vi')

    await waitFor(() => {
      const detailCalls = mock.history.get.filter((r) => r.url?.includes('doc-1'))
      expect(detailCalls.some((r) => r.params?.locale === 'vi')).toBe(true)
    })
  })

  it('shows Publish button when status is draft', async () => {
    renderPage()
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /^publish$/i })).toBeInTheDocument(),
    )
  })

  it('shows both Publish and Unpublish when status is modified', async () => {
    mock.onGet('/api/documents').reply(200, [{ ...doc, Status: 'modified' }])
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^publish$/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /unpublish/i })).toBeInTheDocument()
    })
  })

  it('shows "content type not found" when slug does not match', async () => {
    mock.onGet('/api/content-types').reply(200, [])
    renderPage('/content-type/single-type/unknown-slug')
    await waitFor(() =>
      expect(screen.getByText(/not found/i)).toBeInTheDocument(),
    )
  })
})
