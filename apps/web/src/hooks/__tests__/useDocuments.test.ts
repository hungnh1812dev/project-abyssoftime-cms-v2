import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MockAdapter from 'axios-mock-adapter'
import type { ReactNode } from 'react'
import { createElement } from 'react'
import { api } from '@/lib/api'
import {
  useDocuments,
  useDocument,
  useCreateDocument,
  useUpdateDocument,
  useDeleteDocument,
  usePublishDocument,
  useUnpublishDocument,
  useLocales,
} from '@/hooks/useDocuments'
import type { Document } from '@/types/cms'

let mock: MockAdapter

beforeEach(() => {
  mock = new MockAdapter(api)
})

afterEach(() => {
  mock.restore()
})

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children)
  }
}

const contentTypeSlug = 'articles'

const doc: Document = {
  documentId: '1',
  contentTypeId: 'ct-1',
  status: 'draft',
  data: { title: 'Hello' },
  locale: 'en',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
  createdBy: 'user-1',
  updatedBy: 'user-1',
}

describe('useDocuments', () => {
  it('returns documents for a content type from GET /api/document-manager/{slug}', async () => {
    mock.onGet(`/api/document-manager/${contentTypeSlug}`).reply(200, [doc])
    const { result } = renderHook(() => useDocuments(contentTypeSlug), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual([doc])
  })

  it('is disabled when contentTypeSlug is empty', () => {
    const { result } = renderHook(() => useDocuments(''), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe('idle')
  })
})

describe('useDocument', () => {
  it('returns a single document from GET /api/document-manager/{slug}/{id}', async () => {
    mock.onGet(`/api/document-manager/${contentTypeSlug}/1`).reply(200, doc)
    const { result } = renderHook(() => useDocument(contentTypeSlug, '1', 'en'), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(doc)
  })

  it('is disabled when id is empty', () => {
    const { result } = renderHook(() => useDocument(contentTypeSlug, '', 'en'), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe('idle')
  })

  it('sends locale as a query param and includes it in the query key', async () => {
    mock.onGet(`/api/document-manager/${contentTypeSlug}/1`).reply(200, doc)
    const { result } = renderHook(() => useDocument(contentTypeSlug, '1', 'vi'), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mock.history.get[0].params).toEqual({ locale: 'vi' })
  })

  it('refetches when locale changes (different query key)', async () => {
    mock.onGet(`/api/document-manager/${contentTypeSlug}/1`).reply(200, doc)
    const { result, rerender } = renderHook(({ locale }: { locale: string }) => useDocument(contentTypeSlug, '1', locale), {
      wrapper: createWrapper(),
      initialProps: { locale: 'en' },
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    rerender({ locale: 'vi' })
    await waitFor(() => expect(mock.history.get).toHaveLength(2))
    expect(mock.history.get[1].params).toEqual({ locale: 'vi' })
  })
})

describe('useLocales', () => {
  it('returns the configured locale list from GET /api/locales', async () => {
    mock.onGet('/api/locales').reply(200, ['en', 'vi'])
    const { result } = renderHook(() => useLocales(), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(['en', 'vi'])
  })
})

describe('useCreateDocument', () => {
  it('posts to /api/document-manager/{slug} and succeeds', async () => {
    mock.onPost(`/api/document-manager/${contentTypeSlug}`).reply(201, doc)
    const { result } = renderHook(() => useCreateDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ contentTypeSlug, data: { title: 'Hello' } })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(doc)
  })
})

describe('useUpdateDocument', () => {
  it('puts to /api/document-manager/{slug}/{id} and succeeds', async () => {
    mock.onPut(`/api/document-manager/${contentTypeSlug}/1`).reply(200, doc)
    const { result } = renderHook(() => useUpdateDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ contentTypeSlug, id: '1', data: { title: 'Updated' } })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(doc)
  })

  it('sends locale as a query param', async () => {
    mock.onPut(`/api/document-manager/${contentTypeSlug}/1`).reply(200, doc)
    const { result } = renderHook(() => useUpdateDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ contentTypeSlug, id: '1', data: { title: 'Updated' }, locale: 'vi' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mock.history.put[0].params).toEqual({ locale: 'vi' })
  })
})

describe('useDeleteDocument', () => {
  it('deletes /api/document-manager/{slug}/{id} and succeeds', async () => {
    mock.onDelete(`/api/document-manager/${contentTypeSlug}/1`).reply(204)
    const { result } = renderHook(() => useDeleteDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ contentTypeSlug, id: '1' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
  })
})

describe('usePublishDocument', () => {
  it('posts to /api/document-manager/{slug}/{id}/publish and succeeds', async () => {
    mock.onPost(`/api/document-manager/${contentTypeSlug}/1/publish`).reply(200, { status: 'published' })
    const { result } = renderHook(() => usePublishDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ contentTypeSlug, id: '1' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual({ status: 'published' })
  })

  it('sends locale as a query param', async () => {
    mock.onPost(`/api/document-manager/${contentTypeSlug}/1/publish`).reply(200, { status: 'published' })
    const { result } = renderHook(() => usePublishDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ contentTypeSlug, id: '1', locale: 'vi' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mock.history.post[0].params).toEqual({ locale: 'vi' })
  })
})

describe('useUnpublishDocument', () => {
  it('posts to /api/document-manager/{slug}/{id}/unpublish and succeeds', async () => {
    mock.onPost(`/api/document-manager/${contentTypeSlug}/1/unpublish`).reply(200, { status: 'draft' })
    const { result } = renderHook(() => useUnpublishDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ contentTypeSlug, id: '1' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual({ status: 'draft' })
  })

  it('sends locale as a query param', async () => {
    mock.onPost(`/api/document-manager/${contentTypeSlug}/1/unpublish`).reply(200, { status: 'draft' })
    const { result } = renderHook(() => useUnpublishDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ contentTypeSlug, id: '1', locale: 'vi' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mock.history.post[0].params).toEqual({ locale: 'vi' })
  })
})
