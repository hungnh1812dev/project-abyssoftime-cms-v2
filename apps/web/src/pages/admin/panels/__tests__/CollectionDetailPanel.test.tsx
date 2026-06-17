import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import MockAdapter from 'axios-mock-adapter'
import { api } from '@/lib/api'
import { renderWithProviders } from '@/test-utils'
import { CollectionDetailPanel } from '../CollectionDetailPanel'
import type { ContentType, Document } from '@/types/cms'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

const ct: ContentType = {
  ID: 'ct-1',
  DocumentID: 'ct-doc-1',
  Name: 'Blog Posts',
  Slug: 'blog-posts',
  Kind: 'collection',
  CreatedAt: '',
  UpdatedAt: '',
}

const doc: Document = {
  EntryID: 'doc-1',
  ContentTypeID: 'ct-1',
  Status: 'draft',
  Data: { title: 'First Post', body: 'Some content' },
  Locale: 'en',
  CreatedAt: '',
  UpdatedAt: '',
  CreatedBy: '',
  UpdatedBy: '',
}

let mock: MockAdapter

beforeEach(() => {
  mock = new MockAdapter(api)
  mock.onGet('/api/locales').reply(200, ['en'])
  mock.onGet('/api/documents/doc-1').reply(200, doc)
})

afterEach(() => {
  mock.restore()
  vi.clearAllMocks()
})

describe('CollectionDetailPanel', () => {
  it('renders a form field for each key in document.Data', async () => {
    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() => {
      expect(screen.getByLabelText('title')).toBeInTheDocument()
      expect(screen.getByLabelText('body')).toBeInTheDocument()
    })
  })

  it('shows a back link to the new collection-type list path', async () => {
    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() => {
      const link = screen.getByRole('link', { name: /back/i })
      expect(link).toHaveAttribute('href', '/admin/content-type/collection-type/blog-posts')
    })
  })

  it('Save button is disabled on initial load', async () => {
    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() => expect(screen.getByRole('button', { name: /save/i })).toBeDisabled())
  })

  it('Save button becomes enabled after editing a field', async () => {
    const user = userEvent.setup()
    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() => expect(screen.getByLabelText('title')).toHaveValue('First Post'))
    await user.clear(screen.getByLabelText('title'))
    await user.type(screen.getByLabelText('title'), 'Updated')
    expect(screen.getByRole('button', { name: /save/i })).not.toBeDisabled()
  })

  it('shows success toast and resets form to clean after successful save', async () => {
    const { toast } = await import('sonner')
    const user = userEvent.setup()
    mock.onPut('/api/documents/doc-1').reply(200, { ...doc, Data: { title: 'Updated', body: 'Some content' } })

    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() => expect(screen.getByLabelText('title')).toHaveValue('First Post'))
    await user.clear(screen.getByLabelText('title'))
    await user.type(screen.getByLabelText('title'), 'Updated')
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

    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() => expect(screen.getByLabelText('title')).toHaveValue('First Post'))
    await user.clear(screen.getByLabelText('title'))
    await user.type(screen.getByLabelText('title'), 'bad')

    await user.click(screen.getByRole('button', { name: /save/i }))

    await waitFor(() => expect(toast.error).toHaveBeenCalledWith('Validation failed'))
    expect(screen.getByLabelText('title')).toHaveValue('bad')
  })

  it('shows Publish button when status is draft', async () => {
    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /^publish$/i })).toBeInTheDocument(),
    )
    expect(screen.queryByRole('button', { name: /unpublish/i })).not.toBeInTheDocument()
  })

  it('shows Unpublish button when status is published', async () => {
    mock.onGet('/api/documents/doc-1').reply(200, { ...doc, Status: 'published' })
    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /unpublish/i })).toBeInTheDocument(),
    )
    expect(screen.queryByRole('button', { name: /^publish$/i })).not.toBeInTheDocument()
  })

  it('calls POST /publish when Publish button is clicked', async () => {
    const user = userEvent.setup()
    mock.onPost('/api/documents/doc-1/publish').reply(200, { status: 'published' })
    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() => screen.getByRole('button', { name: /^publish$/i }))
    await user.click(screen.getByRole('button', { name: /^publish$/i }))
    await waitFor(() =>
      expect(mock.history.post.some((r) => r.url?.includes('/publish'))).toBe(true),
    )
  })

  it('shows both Publish and Unpublish buttons when status is modified', async () => {
    mock.onGet('/api/documents/doc-1').reply(200, { ...doc, Status: 'modified' })
    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^publish$/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /unpublish/i })).toBeInTheDocument()
    })
  })

  it('renders a locale selector when multiple locales are available', async () => {
    mock.onGet('/api/locales').reply(200, ['en', 'vi'])
    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() =>
      expect(screen.getByRole('combobox', { name: /locale/i })).toBeInTheDocument(),
    )
  })

  it('does not show locale selector when only one locale is available', async () => {
    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() => screen.getByLabelText('title'))
    expect(screen.queryByRole('combobox', { name: /locale/i })).not.toBeInTheDocument()
  })

  it('sends active locale as query param on publish', async () => {
    const user = userEvent.setup()
    mock.onGet('/api/locales').reply(200, ['en', 'vi'])
    mock.onPost('/api/documents/doc-1/publish').reply(200, { status: 'published' })
    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)
    await waitFor(() => screen.getByRole('button', { name: /^publish$/i }))
    await user.click(screen.getByRole('button', { name: /^publish$/i }))
    await waitFor(() =>
      expect(mock.history.post.some((r) => r.url?.includes('/publish'))).toBe(true),
    )
    expect(mock.history.post[0].params).toEqual({ locale: 'en' })
  })
})
