import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import MockAdapter from 'axios-mock-adapter'
import { api } from '@/lib/api'
import { renderWithProviders } from '@/test-utils'
import { CollectionDetailPanel } from '../CollectionDetailPanel'
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

const doc: Document = {
  ID: 'doc-1',
  DocumentID: 'doc-doc-1',
  ContentTypeID: 'ct-1',
  Status: 'draft',
  Data: { title: 'First Post', body: 'Some content' },
  CreatedAt: '',
  UpdatedAt: '',
}

let mock: MockAdapter

beforeEach(() => {
  mock = new MockAdapter(api)
})

afterEach(() => {
  mock.restore()
})

describe('CollectionDetailPanel', () => {
  it('renders a form field for each key in document.Data', async () => {
    mock.onGet('/api/documents/doc-1').reply(200, doc)

    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)

    await waitFor(() => {
      expect(screen.getByLabelText('title')).toBeInTheDocument()
      expect(screen.getByLabelText('body')).toBeInTheDocument()
    })
  })

  it('shows a back link to the collection list', async () => {
    mock.onGet('/api/documents/doc-1').reply(200, doc)

    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)

    await waitFor(() => {
      const link = screen.getByRole('link', { name: /back/i })
      expect(link).toHaveAttribute('href', '/admin/content-types/blog-posts')
    })
  })

  it('shows Publish button when status is draft', async () => {
    mock.onGet('/api/documents/doc-1').reply(200, doc)

    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)

    await waitFor(() =>
      expect(screen.getByRole('button', { name: /^publish$/i })).toBeInTheDocument(),
    )
    expect(screen.queryByRole('button', { name: /unpublish/i })).not.toBeInTheDocument()
  })

  it('shows Unpublish button when status is published', async () => {
    const published: Document = { ...doc, Status: 'published' }
    mock.onGet('/api/documents/doc-1').reply(200, published)

    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)

    await waitFor(() =>
      expect(screen.getByRole('button', { name: /unpublish/i })).toBeInTheDocument(),
    )
    expect(screen.queryByRole('button', { name: /^publish$/i })).not.toBeInTheDocument()
  })

  it('calls POST /publish when Publish button is clicked', async () => {
    const user = userEvent.setup()
    mock.onGet('/api/documents/doc-1').reply(200, doc)
    mock.onPost('/api/documents/doc-1/publish').reply(200, { status: 'published' })

    renderWithProviders(<CollectionDetailPanel contentType={ct} documentId="doc-1" />)

    await waitFor(() => screen.getByRole('button', { name: /^publish$/i }))
    await user.click(screen.getByRole('button', { name: /^publish$/i }))

    await waitFor(() =>
      expect(mock.history.post.some((r) => r.url?.includes('/publish'))).toBe(true),
    )
  })
})
