import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import MockAdapter from 'axios-mock-adapter'
import { api } from '@/lib/api'
import { renderWithProviders } from '@/test-utils'
import { SingleTypePanel } from '../SingleTypePanel'
import type { ContentType, Document } from '@/types/cms'

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

beforeEach(() => {
  mock = new MockAdapter(api)
})

afterEach(() => {
  mock.restore()
})

describe('SingleTypePanel', () => {
  it('renders a form field for each key in document.Data', async () => {
    mock.onGet('/api/documents').reply(200, [doc])
    mock.onGet('/api/documents/doc-1').reply(200, doc)

    renderWithProviders(<SingleTypePanel contentType={ct} />)

    await waitFor(() => {
      expect(screen.getByLabelText('title')).toBeInTheDocument()
      expect(screen.getByLabelText('body')).toBeInTheDocument()
    })
  })

  it('shows Publish button when document status is draft', async () => {
    mock.onGet('/api/documents').reply(200, [doc])
    mock.onGet('/api/documents/doc-1').reply(200, doc)

    renderWithProviders(<SingleTypePanel contentType={ct} />)

    await waitFor(() =>
      expect(screen.getByRole('button', { name: /^publish$/i })).toBeInTheDocument(),
    )
    expect(screen.queryByRole('button', { name: /unpublish/i })).not.toBeInTheDocument()
  })

  it('shows Unpublish button when document status is published', async () => {
    const published: Document = { ...doc, Status: 'published' }
    mock.onGet('/api/documents').reply(200, [published])
    mock.onGet('/api/documents/doc-1').reply(200, published)

    renderWithProviders(<SingleTypePanel contentType={ct} />)

    await waitFor(() =>
      expect(screen.getByRole('button', { name: /unpublish/i })).toBeInTheDocument(),
    )
    expect(screen.queryByRole('button', { name: /^publish$/i })).not.toBeInTheDocument()
  })

  it('calls POST /publish when Publish button is clicked', async () => {
    const user = userEvent.setup()
    mock.onGet('/api/documents').reply(200, [doc])
    mock.onGet('/api/documents/doc-1').reply(200, doc)
    mock.onPost('/api/documents/doc-1/publish').reply(200, { status: 'published' })

    renderWithProviders(<SingleTypePanel contentType={ct} />)

    await waitFor(() => screen.getByRole('button', { name: /^publish$/i }))
    await user.click(screen.getByRole('button', { name: /^publish$/i }))

    await waitFor(() =>
      expect(mock.history.post.some((r) => r.url?.includes('/publish'))).toBe(true),
    )
  })

  it('shows empty state when no document exists', async () => {
    mock.onGet('/api/documents').reply(200, [])

    renderWithProviders(<SingleTypePanel contentType={ct} />)

    await waitFor(() => expect(screen.getByText(/no document/i)).toBeInTheDocument())
  })

  it('shows both Publish and Unpublish buttons when status is modified', async () => {
    const modified: Document = { ...doc, Status: 'modified' }
    mock.onGet('/api/documents').reply(200, [modified])
    mock.onGet('/api/documents/doc-1').reply(200, modified)

    renderWithProviders(<SingleTypePanel contentType={ct} />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^publish$/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /unpublish/i })).toBeInTheDocument()
    })
  })

  it('renders a locale selector when multiple locales are available', async () => {
    mock.onGet('/api/locales').reply(200, ['en', 'vi'])
    mock.onGet('/api/documents').reply(200, [doc])
    mock.onGet('/api/documents/doc-1').reply(200, doc)

    renderWithProviders(<SingleTypePanel contentType={ct} />)

    await waitFor(() =>
      expect(screen.getByRole('combobox', { name: /locale/i })).toBeInTheDocument(),
    )
  })

  it('sends active locale as query param on publish', async () => {
    const user = userEvent.setup()
    mock.onGet('/api/locales').reply(200, ['en', 'vi'])
    mock.onGet('/api/documents').reply(200, [doc])
    mock.onGet('/api/documents/doc-1').reply(200, doc)
    mock.onPost('/api/documents/doc-1/publish').reply(200, { status: 'published' })

    renderWithProviders(<SingleTypePanel contentType={ct} />)

    await waitFor(() => screen.getByRole('button', { name: /^publish$/i }))
    await user.click(screen.getByRole('button', { name: /^publish$/i }))

    await waitFor(() =>
      expect(mock.history.post.some((r) => r.url?.includes('/publish'))).toBe(true),
    )
    expect(mock.history.post[0].params).toEqual({ locale: 'en' })
  })
})
