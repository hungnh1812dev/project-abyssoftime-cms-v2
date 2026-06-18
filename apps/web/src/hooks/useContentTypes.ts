import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { ContentType } from '@/types/cms'

const KEYS = {
  all: ['content-types'] as const,
  detail: (id: string) => ['content-types', id] as const,
  bySlug: (slug: string) => ['content-types', 'by-slug', slug] as const,
}

export function useContentTypes() {
  return useQuery({
    queryKey: KEYS.all,
    queryFn: () => api.get<ContentType[]>('/api/content-types').then((r) => r.data),
  })
}

export function useContentType(id: string) {
  return useQuery({
    queryKey: KEYS.detail(id),
    queryFn: () => api.get<ContentType>(`/api/content-types/${id}`).then((r) => r.data),
    enabled: Boolean(id),
  })
}

export function useContentTypeBySlug(slug: string) {
  return useQuery({
    queryKey: KEYS.bySlug(slug),
    queryFn: () => api.get<ContentType>(`/api/content-types/${slug}`).then((r) => r.data),
    enabled: Boolean(slug),
  })
}
