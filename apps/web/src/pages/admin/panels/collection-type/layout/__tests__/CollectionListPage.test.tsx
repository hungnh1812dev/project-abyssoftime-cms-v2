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
  Fields: [
    { name: 'title', type: 'text' },
    { name: 'active', type: 'boolean' },
    { name: 'views', type: 'number' },
  ],
  CreatedAt: '',
  UpdatedAt: '',
}

const doc1: Document = {
  status: 'draft',
  data: { documentId: 'doc-1', locale: 'en', createdAt: '', updatedAt: '', title: 'First Post', active: true, views: 42 },
}

const doc2: Document = {
  status: 'published',
  data: { documentId: 'doc-2', locale: 'en', createdAt: '', updatedAt: '', title: 'Second Post', active: false, views: 7 },
}

let mock: MockAdapter

beforeEach(() => {
  mock = new MockAdapter(api)
  mock.onGet('/api/locales').reply(200, ['en'])
})

afterEach(() => {
  mock.restore()
  vi.clearAllMocks()
})

describe('CollectionListPage — fallback (no registry columns)', () => {
  it('renders a row for each document using the first Data field as display', async () => {
    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1, doc2], total: 2, start: 0, size: 20 })
    renderWithProviders(<CollectionListPage contentType={ct} />)
    await waitFor(() => {
      expect(screen.getByText('First Post')).toBeInTheDocument()
      expect(screen.getByText('Second Post')).toBeInTheDocument()
    })
  })

  it('shows the status for each document', async () => {
    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1, doc2], total: 2, start: 0, size: 20 })
    renderWithProviders(<CollectionListPage contentType={ct} />)
    await waitFor(() => {
      expect(screen.getByText('draft')).toBeInTheDocument()
      expect(screen.getByText('published')).toBeInTheDocument()
    })
  })

  it('shows empty state when no documents exist', async () => {
    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [], total: 0, start: 0, size: 20 })
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

    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1], total: 1, start: 0, size: 20 })
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

    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1, doc2], total: 2, start: 0, size: 20 })
    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => {
      expect(screen.getByText('✓')).toBeInTheDocument()
      expect(screen.getAllByText('—').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('renders number column as a string value', async () => {
    const { getRegistration } = await import('@/content-type-registry')
    vi.mocked(getRegistration).mockReturnValue({
      slug: 'blog-posts',
      kind: 'collection',
      columns: [{ key: 'views', label: 'Views', type: 'number' }],
    })

    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1], total: 1, start: 0, size: 20 })
    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => expect(screen.getByText('42')).toBeInTheDocument())
  })

  it('renders image column as an img element', async () => {
    const { getRegistration } = await import('@/content-type-registry')
    const imgDoc: Document = { ...doc1, data: { ...doc1.data, cover: 'https://example.com/img.jpg' } }
    vi.mocked(getRegistration).mockReturnValue({
      slug: 'blog-posts',
      kind: 'collection',
      columns: [{ key: 'cover', label: 'Cover', type: 'image' }],
    })

    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [imgDoc], total: 1, start: 0, size: 20 })
    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => {
      const img = screen.getByRole('img')
      expect(img).toHaveAttribute('src', 'https://example.com/img.jpg')
    })
  })
})

describe('CollectionListPage — navigation', () => {
  it('Edit icon button is rendered for each document', async () => {
    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1], total: 1, start: 0, size: 20 })
    renderWithProviders(<CollectionListPage contentType={ct} />, {
      initialEntries: ['/admin/content-type/collection-type/blog-posts'],
    })
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /edit/i })).toBeInTheDocument()
    })
  })

  it('Add new item navigates to /new without creating a document', async () => {
    const user = userEvent.setup()
    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [], total: 0, start: 0, size: 20 })

    renderWithProviders(<CollectionListPage contentType={ct} />, {
      initialEntries: ['/admin/content-type/collection-type/blog-posts'],
    })

    await waitFor(() => screen.getByRole('button', { name: /add/i }))
    await user.click(screen.getByRole('button', { name: /add/i }))

    expect(mock.history.post).toHaveLength(0)
  })

  it('Delete button shows confirm dialog and calls DELETE', async () => {
    const user = userEvent.setup()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1], total: 1, start: 0, size: 20 })
    mock.onDelete('/api/document-manager/collection-type/blog-posts/doc-1').reply(204)

    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => screen.getByRole('button', { name: /delete/i }))
    await user.click(screen.getByRole('button', { name: /delete/i }))

    expect(window.confirm).toHaveBeenCalled()
    await waitFor(() =>
      expect(mock.history.delete.some((r) => r.url === '/api/document-manager/collection-type/blog-posts/doc-1')).toBe(true),
    )
  })

  it('Delete button does not call DELETE when user cancels confirm', async () => {
    const user = userEvent.setup()
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1], total: 1, start: 0, size: 20 })

    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => screen.getByRole('button', { name: /delete/i }))
    await user.click(screen.getByRole('button', { name: /delete/i }))

    expect(mock.history.delete).toHaveLength(0)
  })

  it('Duplicate button is rendered for each document', async () => {
    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1, doc2], total: 2, start: 0, size: 20 })
    renderWithProviders(<CollectionListPage contentType={ct} />)
    await waitFor(() => {
      expect(screen.getAllByRole('button', { name: /duplicate/i })).toHaveLength(2)
    })
  })

  it('Duplicate button calls POST duplicate endpoint', async () => {
    const user = userEvent.setup()
    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1], total: 1, start: 0, size: 20 })
    mock.onPost(/\/blog-posts\/doc-1\/duplicate/).reply(201, {
      data: { documentId: 'new-dup', locale: 'en', title: 'First Post' },
      status: 'draft',
    })

    renderWithProviders(<CollectionListPage contentType={ct} />, {
      initialEntries: ['/admin/content-type/collection-type/blog-posts'],
    })

    await waitFor(() => screen.getByRole('button', { name: /duplicate/i }))
    await user.click(screen.getByRole('button', { name: /duplicate/i }))

    await waitFor(() =>
      expect(mock.history.post.some((r) => r.url?.includes('/duplicate'))).toBe(true),
    )
  })
})

describe('CollectionListPage — column chooser', () => {
  it('shows configure columns button when no registry override', async () => {
    const { getRegistration } = await import('@/content-type-registry')
    vi.mocked(getRegistration).mockReturnValue(undefined)

    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1], total: 1, start: 0, size: 20 })
    renderWithProviders(<CollectionListPage contentType={ct} />)
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /configure columns/i })).toBeInTheDocument()
    })
  })

  it('hides configure columns button when registry override exists', async () => {
    const { getRegistration } = await import('@/content-type-registry')
    vi.mocked(getRegistration).mockReturnValue({
      slug: 'blog-posts',
      kind: 'collection',
      columns: [{ key: 'title', label: 'Title', type: 'text' }],
    })

    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1], total: 1, start: 0, size: 20 })
    renderWithProviders(<CollectionListPage contentType={ct} />)
    await waitFor(() => {
      expect(screen.queryByRole('button', { name: /configure columns/i })).not.toBeInTheDocument()
    })
  })

  it('hides system columns when not in listFields', async () => {
    const { getRegistration } = await import('@/content-type-registry')
    vi.mocked(getRegistration).mockReturnValue(undefined)

    const ctWithListFields: ContentType = {
      ...ct,
      listFields: ['title'],
    }
    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1], total: 1, start: 0, size: 20 })
    renderWithProviders(<CollectionListPage contentType={ctWithListFields} />)
    await waitFor(() => {
      expect(screen.getByText('First Post')).toBeInTheDocument()
    })
    expect(screen.queryByText('Created At')).not.toBeInTheDocument()
    expect(screen.queryByText('Updated At')).not.toBeInTheDocument()
    expect(screen.queryByText('Updated By')).not.toBeInTheDocument()
  })

  it('shows system columns when included in listFields', async () => {
    const { getRegistration } = await import('@/content-type-registry')
    vi.mocked(getRegistration).mockReturnValue(undefined)

    const ctWithListFields: ContentType = {
      ...ct,
      listFields: ['title', 'createdAt', 'updatedByName'],
    }
    mock.onGet('/api/document-manager/collection-type/blog-posts').reply(200, { items: [doc1], total: 1, start: 0, size: 20 })
    renderWithProviders(<CollectionListPage contentType={ctWithListFields} />)
    await waitFor(() => {
      expect(screen.getByText('First Post')).toBeInTheDocument()
    })
    expect(screen.getByText('Created At')).toBeInTheDocument()
    expect(screen.queryByText('Updated At')).not.toBeInTheDocument()
    expect(screen.getByText('Updated By')).toBeInTheDocument()
  })
})
