import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MockAdapter from 'axios-mock-adapter'
import type { ReactNode } from 'react'
import { createElement } from 'react'
import { api } from '@/lib/api'
import { useContentTypes, useContentType, useContentTypeBySlug } from '@/hooks/useContentTypes'
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

describe('useContentTypeBySlug', () => {
  it('returns a content type from GET /api/content-types/by-slug/{slug}', async () => {
    mock.onGet('/api/content-types/by-slug/blog').reply(200, ct)
    const { result } = renderHook(() => useContentTypeBySlug('blog'), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(ct)
  })

  it('is disabled when slug is empty', () => {
    const { result } = renderHook(() => useContentTypeBySlug(''), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe('idle')
  })
})
