import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MockAdapter from 'axios-mock-adapter'
import type { ReactNode } from 'react'
import { createElement } from 'react'
import { api } from '@/lib/api'
import { useMediaList } from '@/hooks/useMedia'

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

describe('useMediaList', () => {
  it('fetches paginated media from GET /api/media', async () => {
    const response = {
      items: [
        { ID: 'a1', url: 'https://cdn/a1.jpg', thumbnailUrl: 'https://cdn/a1.jpg', publicId: 'p1', fileName: 'a1_abc123.jpg', fileExt: 'jpg', hash: 'abc123', documentRef: '', contentTypeId: '', createdAt: '2024-01-01T00:00:00Z' },
      ],
      total: 5,
      page: 1,
      limit: 20,
    }
    mock.onGet('/api/media?page=1&limit=20').reply(200, response)

    const { result } = renderHook(() => useMediaList(1, 20), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.total).toBe(5)
    expect(result.current.data?.items).toHaveLength(1)
    expect(result.current.data?.items[0].fileName).toBe('a1_abc123.jpg')
  })
})
