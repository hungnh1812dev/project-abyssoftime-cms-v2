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

const doc: Document = {
  EntryID: '1',
  ContentTypeID: 'ct-1',
  Status: 'draft',
  Data: { title: 'Hello' },
  Locale: 'en',
  CreatedAt: '2024-01-01T00:00:00Z',
  UpdatedAt: '2024-01-01T00:00:00Z',
  CreatedBy: 'user-1',
  UpdatedBy: 'user-1',
}

describe('useDocuments', () => {
  it('returns documents for a content type from GET /api/documents?contentType=ct-1', async () => {
    mock.onGet('/api/documents').reply(200, [doc])
    const { result } = renderHook(() => useDocuments('ct-1'), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual([doc])
  })

  it('is disabled when contentTypeId is empty', () => {
    const { result } = renderHook(() => useDocuments(''), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe('idle')
  })
})

describe('useDocument', () => {
  it('returns a single document from GET /api/documents/{id}', async () => {
    mock.onGet('/api/documents/1').reply(200, doc)
    const { result } = renderHook(() => useDocument('1', 'en'), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(doc)
  })

  it('is disabled when id is empty', () => {
    const { result } = renderHook(() => useDocument('', 'en'), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe('idle')
  })

  it('sends locale as a query param and includes it in the query key', async () => {
    mock.onGet('/api/documents/1').reply(200, doc)
    const { result } = renderHook(() => useDocument('1', 'vi'), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mock.history.get[0].params).toEqual({ locale: 'vi' })
  })

  it('refetches when locale changes (different query key)', async () => {
    mock.onGet('/api/documents/1').reply(200, doc)
    const { result, rerender } = renderHook(({ locale }: { locale: string }) => useDocument('1', locale), {
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
  it('posts to /api/documents and succeeds', async () => {
    mock.onPost('/api/documents').reply(201, doc)
    const { result } = renderHook(() => useCreateDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ contentTypeId: 'ct-1', data: { title: 'Hello' } })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(doc)
  })
})

describe('useUpdateDocument', () => {
  it('puts to /api/documents/{id} and succeeds', async () => {
    mock.onPut('/api/documents/1').reply(200, doc)
    const { result } = renderHook(() => useUpdateDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ id: '1', contentTypeId: 'ct-1', data: { title: 'Updated' } })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(doc)
  })

  it('sends locale as a query param', async () => {
    mock.onPut('/api/documents/1').reply(200, doc)
    const { result } = renderHook(() => useUpdateDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ id: '1', contentTypeId: 'ct-1', data: { title: 'Updated' }, locale: 'vi' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mock.history.put[0].params).toEqual({ locale: 'vi' })
  })
})

describe('useDeleteDocument', () => {
  it('deletes /api/documents/{id} and succeeds', async () => {
    mock.onDelete('/api/documents/1').reply(204)
    const { result } = renderHook(() => useDeleteDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ id: '1', contentTypeId: 'ct-1' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
  })
})

describe('usePublishDocument', () => {
  it('posts to /api/documents/{id}/publish and succeeds', async () => {
    mock.onPost('/api/documents/1/publish').reply(200, { status: 'published' })
    const { result } = renderHook(() => usePublishDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ id: '1', contentTypeId: 'ct-1' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual({ status: 'published' })
  })

  it('sends locale as a query param', async () => {
    mock.onPost('/api/documents/1/publish').reply(200, { status: 'published' })
    const { result } = renderHook(() => usePublishDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ id: '1', contentTypeId: 'ct-1', locale: 'vi' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mock.history.post[0].params).toEqual({ locale: 'vi' })
  })
})

describe('useUnpublishDocument', () => {
  it('posts to /api/documents/{id}/unpublish and succeeds', async () => {
    mock.onPost('/api/documents/1/unpublish').reply(200, { status: 'draft' })
    const { result } = renderHook(() => useUnpublishDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ id: '1', contentTypeId: 'ct-1' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual({ status: 'draft' })
  })

  it('sends locale as a query param', async () => {
    mock.onPost('/api/documents/1/unpublish').reply(200, { status: 'draft' })
    const { result } = renderHook(() => useUnpublishDocument(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ id: '1', contentTypeId: 'ct-1', locale: 'vi' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mock.history.post[0].params).toEqual({ locale: 'vi' })
  })
})
