import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import MockAdapter from 'axios-mock-adapter'
import { api } from '@/lib/api'
import { renderWithProviders } from '@/test-utils'
import { CollectionListPage } from '../CollectionListPage'
import type { ContentType, Document } from '@/types/cms'

const ct: ContentType = {
  ID: 'ct-1',
  DocumentID: 'ct-doc-1',
  Name: 'Blog Posts',
  Slug: 'blog-posts',
  Kind: 'collection',
  CreatedAt: '',
  UpdatedAt: '',
}

const doc1: Document = {
  EntryID: 'doc-1',
  ContentTypeID: 'ct-1',
  Status: 'draft',
  Data: { title: 'First Post' },
  Locale: 'en',
  CreatedAt: '',
  UpdatedAt: '',
  CreatedBy: '',
  UpdatedBy: '',
}

const doc2: Document = {
  EntryID: 'doc-2',
  ContentTypeID: 'ct-1',
  Status: 'published',
  Data: { title: 'Second Post' },
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
})

describe('CollectionListPage', () => {
  it('renders a row for each document using the first Data field as display', async () => {
    mock.onGet('/api/documents').reply(200, [doc1, doc2])

    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => {
      expect(screen.getByText('First Post')).toBeInTheDocument()
      expect(screen.getByText('Second Post')).toBeInTheDocument()
    })
  })

  it('shows the status for each document', async () => {
    mock.onGet('/api/documents').reply(200, [doc1, doc2])

    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => {
      expect(screen.getByText('draft')).toBeInTheDocument()
      expect(screen.getByText('published')).toBeInTheDocument()
    })
  })

  it('Edit link points to the detail page', async () => {
    mock.onGet('/api/documents').reply(200, [doc1])

    renderWithProviders(<CollectionListPage contentType={ct} />, {
      initialEntries: ['/admin/content-types/blog-posts'],
    })

    await waitFor(() => {
      const link = screen.getByRole('link', { name: /edit/i })
      expect(link).toHaveAttribute('href', '/admin/content-types/blog-posts/doc-1')
    })
  })

  it('Delete button calls DELETE /api/documents/:id', async () => {
    const user = userEvent.setup()
    mock.onGet('/api/documents').reply(200, [doc1])
    mock.onDelete('/api/documents/doc-1').reply(204)

    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => screen.getByRole('button', { name: /delete/i }))
    await user.click(screen.getByRole('button', { name: /delete/i }))

    await waitFor(() =>
      expect(mock.history.delete.some((r) => r.url === '/api/documents/doc-1')).toBe(true),
    )
  })

  it('shows empty state when no documents exist', async () => {
    mock.onGet('/api/documents').reply(200, [])

    renderWithProviders(<CollectionListPage contentType={ct} />)

    await waitFor(() => expect(screen.getByText(/no entries/i)).toBeInTheDocument())
  })

  it('Add entry button posts to /api/documents', async () => {
    const user = userEvent.setup()
    mock.onGet('/api/documents').reply(200, [])
    mock.onPost('/api/documents').reply(201, { ...doc1, EntryID: 'doc-new' })

    renderWithProviders(<CollectionListPage contentType={ct} />, {
      initialEntries: ['/admin/content-types/blog-posts'],
    })

    await waitFor(() => screen.getByRole('button', { name: /add entry/i }))
    await user.click(screen.getByRole('button', { name: /add entry/i }))

    await waitFor(() =>
      expect(mock.history.post.some((r) => r.url === '/api/documents')).toBe(true),
    )
  })
})
