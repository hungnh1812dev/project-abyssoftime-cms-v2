import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import MockAdapter from 'axios-mock-adapter'
import { api } from '@/lib/api'
import { renderWithProviders } from '@/test-utils'
import { CollectionListPage } from '../CollectionListPage'
import type { ContentType, Document } from '@/types/cms'

vi.mock('@/content-type-registry', () => ({
  getRegistration: vi.fn().mockReturnValue(undefined),
}))

const ct: ContentType = {
  ID: 'ct-1',
  Name: 'Blog Posts',
  Slug: 'blog-posts',
  Kind: 'collection',
  CreatedAt: '',
  UpdatedAt: '',
}

const doc1: Document = {
  DocumentID: 'doc-1',
  ContentTypeID: 'ct-1',
  Status: 'draft',
  Data: { title: 'First Post', active: true, views: 42 },
  Locale: 'en',
  CreatedAt: '',
  UpdatedAt: '',
  CreatedBy: '',
  UpdatedBy: '',
}

const doc2: Document = {
  DocumentID: 'doc-2',
  ContentTypeID: 'ct-1',
  Status: 'published',
  Data: { title: 'Second Post', active: false, views: 7 },
  Locale: 'en',
  CreatedAt: '',
  UpdatedAt: '',
  CreatedBy: '',
  UpdatedBy: '',
}

let mock: MockAdapter

beforeEach(() => {
  mock = new MockAdapter(api)
})

afterEach(() => {
  mock.restore()
  vi.clearAllMocks()
})

describe('CollectionListPage — fallback (no registry columns)', () => {
  it('renders a row for each document using the first Data field as display', async () => {
    mock.onGet('/api/content-types/blog-posts/documents').reply(200, [doc1, doc2])
    renderWithProviders(<CollectionListPage contentType={ct} />)
    await waitFor(() => {
      expect(screen.getByText('First Post')).toBeInTheDocument()
      expect(screen.getByText('Second Post')).toBeInTheDocument()
    })
  })

  it('shows the status for each document', async () => {
    mock.onGet('/api/content-types/blog-posts/documents').reply(200, [doc1, doc2])
    renderWithProviders(<CollectionListPage contentType={ct} />)
    await waitFor(() => {
      expect(screen.getByText('draft')).toBeInTheDocument()
      expect(screen.getByText('published')).toBeInTheDocument()
    })
  })

  it('shows empty state when no documents exist', async () => {
    mock.onGet('/api/content-types/blog-posts/documents').reply(200, [])
    renderWithProviders(<CollectionListPage contentType={ct} />)
    await waitFor(() => expect(screen.getByText(/no entries/i)).toBeInTheDocument())
  })
})

describe('CollectionListPage — registry columns', () => {
  it('renders columns defined in the registry', async () => {
    const { getRegistration } = await import('@/content-type-registry')
    vi.mocked(getRegistration).mockReturnValue({
      slug: 'blog-posts',
      kind: 'collection',
      columns: [
        { key: 'title', label: 'Title', type: 'text' },
        { key: 'active', label: 'Active', type: 'boolean' },
        { key: 'views', label: 'Views', type: 'number' },
      ],
    })

    mock.onGet('/api/content-types/blog-posts/documents').reply(200, [doc1])
    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => {
      expect(screen.getByRole('columnheader', { name: 'Title' })).toBeInTheDocument()
      expect(screen.getByRole('columnheader', { name: 'Active' })).toBeInTheDocument()
      expect(screen.getByRole('columnheader', { name: 'Views' })).toBeInTheDocument()
    })
  })

  it('renders boolean column as ✓ when true and — when false', async () => {
    const { getRegistration } = await import('@/content-type-registry')
    vi.mocked(getRegistration).mockReturnValue({
      slug: 'blog-posts',
      kind: 'collection',
      columns: [{ key: 'active', label: 'Active', type: 'boolean' }],
    })

    mock.onGet('/api/content-types/blog-posts/documents').reply(200, [doc1, doc2])
    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => {
      expect(screen.getByText('✓')).toBeInTheDocument()
      expect(screen.getByText('—')).toBeInTheDocument()
    })
  })

  it('renders number column as a string value', async () => {
    const { getRegistration } = await import('@/content-type-registry')
    vi.mocked(getRegistration).mockReturnValue({
      slug: 'blog-posts',
      kind: 'collection',
      columns: [{ key: 'views', label: 'Views', type: 'number' }],
    })

    mock.onGet('/api/content-types/blog-posts/documents').reply(200, [doc1])
    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => expect(screen.getByText('42')).toBeInTheDocument())
  })

  it('renders image column as an img element', async () => {
    const { getRegistration } = await import('@/content-type-registry')
    const imgDoc: Document = { ...doc1, Data: { cover: 'https://example.com/img.jpg' } }
    vi.mocked(getRegistration).mockReturnValue({
      slug: 'blog-posts',
      kind: 'collection',
      columns: [{ key: 'cover', label: 'Cover', type: 'image' }],
    })

    mock.onGet('/api/content-types/blog-posts/documents').reply(200, [imgDoc])
    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => {
      const img = screen.getByRole('img')
      expect(img).toHaveAttribute('src', 'https://example.com/img.jpg')
    })
  })
})

describe('CollectionListPage — navigation', () => {
  it('Edit link points to the new collection-type detail path', async () => {
    mock.onGet('/api/content-types/blog-posts/documents').reply(200, [doc1])
    renderWithProviders(<CollectionListPage contentType={ct} />, {
      initialEntries: ['/admin/content-type/collection-type/blog-posts'],
    })
    await waitFor(() => {
      const link = screen.getByRole('link', { name: /edit/i })
      expect(link).toHaveAttribute('href', '/admin/content-type/collection-type/blog-posts/doc-1')
    })
  })

  it('Add entry button creates a document and navigates to detail page', async () => {
    const user = userEvent.setup()
    mock.onGet('/api/content-types/blog-posts/documents').reply(200, [])
    mock.onPost('/api/content-types/blog-posts/documents').reply(201, { ...doc1, DocumentID: 'doc-new' })

    renderWithProviders(<CollectionListPage contentType={ct} />, {
      initialEntries: ['/admin/content-type/collection-type/blog-posts'],
    })

    await waitFor(() => screen.getByRole('button', { name: /add/i }))
    await user.click(screen.getByRole('button', { name: /add/i }))

    await waitFor(() =>
      expect(mock.history.post.some((r) => r.url === '/api/content-types/blog-posts/documents')).toBe(true),
    )
  })

  it('Delete button shows confirm dialog and calls DELETE', async () => {
    const user = userEvent.setup()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    mock.onGet('/api/content-types/blog-posts/documents').reply(200, [doc1])
    mock.onDelete('/api/content-types/blog-posts/documents/doc-1').reply(204)

    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => screen.getByRole('button', { name: /delete/i }))
    await user.click(screen.getByRole('button', { name: /delete/i }))

    expect(window.confirm).toHaveBeenCalled()
    await waitFor(() =>
      expect(mock.history.delete.some((r) => r.url === '/api/content-types/blog-posts/documents/doc-1')).toBe(true),
    )
  })

  it('Delete button does not call DELETE when user cancels confirm', async () => {
    const user = userEvent.setup()
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    mock.onGet('/api/content-types/blog-posts/documents').reply(200, [doc1])

    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => screen.getByRole('button', { name: /delete/i }))
    await user.click(screen.getByRole('button', { name: /delete/i }))

    expect(mock.history.delete).toHaveLength(0)
  })
})
