import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import MockAdapter from 'axios-mock-adapter'
import { api } from '@/lib/api'
import { renderWithProviders } from '@/test-utils'
import { ContentTypePanel } from '../ContentTypePanel'
import type { ContentType, Document } from '@/types/cms'

const ct: ContentType = {
  ID: 'ct-1',
  Name: 'Homepage',
  Slug: 'homepage',
  Kind: 'single',
  Fields: [
    { name: 'title', type: 'text' },
    { name: 'heroImage', type: 'media' },
  ],
  CreatedAt: '',
  UpdatedAt: '',
}

const doc: Document = {
  documentId: 'ct-1',
  contentTypeId: 'ct-1',
  status: 'draft',
  data: { title: 'Hello' },
  locale: 'en',
  createdAt: '',
  updatedAt: '',
  createdBy: '',
  updatedBy: '',
}

let mock: MockAdapter

beforeEach(() => {
  mock = new MockAdapter(api)
})

afterEach(() => {
  mock.restore()
})

describe('ContentTypePanel', () => {
  it('renders schema-driven fields from contentType.Fields', async () => {
    mock.onGet('/api/document-manager/single-type/homepage').reply(200, doc)
    mock.onGet('/api/locales').reply(200, ['en'])
    mock.onGet('/api/document-manager/collection-type/homepage/ct-1').reply(200, doc)

    renderWithProviders(<ContentTypePanel contentType={ct} />)

    await waitFor(() => {
      expect(screen.getByLabelText('title')).toBeInTheDocument()
    })
  })

  it('does not show a Go Back link when no id prop is given', async () => {
    mock.onGet('/api/document-manager/single-type/homepage').reply(200, doc)
    mock.onGet('/api/locales').reply(200, ['en'])

    renderWithProviders(<ContentTypePanel contentType={ct} />)

    await waitFor(() => expect(screen.queryByText(/go back/i)).not.toBeInTheDocument())
  })

  it('shows a Go Back link when id prop is given', async () => {
    const collectionDoc: Document = { ...doc, documentId: 'entry-99' }
    mock.onGet('/api/document-manager/single-type/homepage').reply(200, collectionDoc)
    mock.onGet('/api/locales').reply(200, ['en'])
    mock.onGet('/api/document-manager/collection-type/homepage/entry-99').reply(200, collectionDoc)

    renderWithProviders(<ContentTypePanel contentType={{ ...ct, Kind: 'collection' }} id="entry-99" />)

    await waitFor(() => expect(screen.getByText(/go back/i)).toBeInTheDocument())
  })
})
