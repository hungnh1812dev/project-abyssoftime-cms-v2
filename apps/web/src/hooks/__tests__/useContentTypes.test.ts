import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MockAdapter from 'axios-mock-adapter'
import type { ReactNode } from 'react'
import { createElement } from 'react'
import { api } from '@/lib/api'
import {
  useContentTypes,
  useContentType,
  useCreateContentType,
  useUpdateContentType,
  useDeleteContentType,
} from '@/hooks/useContentTypes'
import type { ContentType } from '@/types/cms'

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

const ct: ContentType = {
  ID: '1',
  DocumentID: 'doc-1',
  Name: 'Blog',
  Slug: 'blog',
  Kind: 'collection',
  CreatedAt: '2024-01-01T00:00:00Z',
  UpdatedAt: '2024-01-01T00:00:00Z',
}

describe('useContentTypes', () => {
  it('returns list of content types from GET /api/content-types', async () => {
    mock.onGet('/api/content-types').reply(200, [ct])
    const { result } = renderHook(() => useContentTypes(), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual([ct])
  })
})

describe('useContentType', () => {
  it('returns a single content type from GET /api/content-types/{id}', async () => {
    mock.onGet('/api/content-types/1').reply(200, ct)
    const { result } = renderHook(() => useContentType('1'), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(ct)
  })

  it('is disabled when id is empty', () => {
    const { result } = renderHook(() => useContentType(''), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe('idle')
  })
})

describe('useCreateContentType', () => {
  it('posts to /api/content-types and succeeds', async () => {
    mock.onPost('/api/content-types').reply(201, ct)
    const { result } = renderHook(() => useCreateContentType(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ name: 'Blog', slug: 'blog', kind: 'collection' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(ct)
  })
})

describe('useUpdateContentType', () => {
  it('puts to /api/content-types/{id} and succeeds', async () => {
    mock.onPut('/api/content-types/1').reply(200, ct)
    const { result } = renderHook(() => useUpdateContentType(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate({ id: '1', name: 'Blog Updated', slug: 'blog', kind: 'collection' })
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(ct)
  })
})

describe('useDeleteContentType', () => {
  it('deletes /api/content-types/{id} and succeeds', async () => {
    mock.onDelete('/api/content-types/1').reply(204)
    const { result } = renderHook(() => useDeleteContentType(), { wrapper: createWrapper() })
    await act(async () => {
      result.current.mutate('1')
    })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
  })
})
