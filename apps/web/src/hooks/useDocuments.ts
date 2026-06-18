import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { toast } from 'sonner'
import { api } from '@/lib/api'
import type { Document } from '@/types/cms'

function onMutationError(err: unknown) {
  const msg = (err as AxiosError<{ error: string }>).response?.data?.error ?? 'Something went wrong'
  toast.error(msg)
}

const KEYS = {
  list: (contentTypeSlug: string) => ['documents', contentTypeSlug] as const,
  detail: (contentTypeSlug: string, id: string, locale: string) =>
    ['documents', 'detail', contentTypeSlug, id, locale] as const,
  locales: ['locales'] as const,
}

export function useDocuments(contentTypeSlug: string) {
  return useQuery({
    queryKey: KEYS.list(contentTypeSlug),
    queryFn: () =>
      api
        .get<Document[]>(`/api/content-types/${contentTypeSlug}/documents`)
        .then((r) => r.data),
    enabled: Boolean(contentTypeSlug),
  })
}

export function useDocument(contentTypeSlug: string, id: string, locale: string) {
  return useQuery({
    queryKey: KEYS.detail(contentTypeSlug, id, locale),
    queryFn: () =>
      api
        .get<Document>(`/api/content-types/${contentTypeSlug}/documents/${id}`, {
          params: { locale },
        })
        .then((r) => r.data),
    enabled: Boolean(id),
  })
}

export function useLocales() {
  return useQuery({
    queryKey: KEYS.locales,
    queryFn: () => api.get<string[]>('/api/locales').then((r) => r.data),
  })
}

export function useCreateDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      contentTypeSlug,
      data,
    }: {
      contentTypeSlug: string
      data: Record<string, unknown>
    }) =>
      api
        .post<Document>(`/api/content-types/${contentTypeSlug}/documents`, { data })
        .then((r) => r.data),
    onSuccess: (_, { contentTypeSlug }) =>
      qc.invalidateQueries({ queryKey: KEYS.list(contentTypeSlug) }),
    onError: onMutationError,
  })
}

export function useUpdateDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      contentTypeSlug,
      id,
      locale,
      data,
    }: {
      contentTypeSlug: string
      id: string
      data: Record<string, unknown>
      locale?: string
    }) =>
      api
        .put<Document>(`/api/content-types/${contentTypeSlug}/documents/${id}`, { data }, { params: { locale } })
        .then((r) => r.data),
    onSuccess: (data, { contentTypeSlug }) => {
      qc.invalidateQueries({ queryKey: KEYS.list(contentTypeSlug) })
      qc.invalidateQueries({
        queryKey: KEYS.detail(contentTypeSlug, data.DocumentID, data.Locale),
      })
    },
    onError: onMutationError,
  })
}

export function useDeleteDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      contentTypeSlug,
      id,
    }: {
      contentTypeSlug: string
      id: string
    }) => api.delete(`/api/content-types/${contentTypeSlug}/documents/${id}`),
    onSuccess: (_, { contentTypeSlug }) =>
      qc.invalidateQueries({ queryKey: KEYS.list(contentTypeSlug) }),
    onError: onMutationError,
  })
}

export function usePublishDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      contentTypeSlug,
      id,
      locale,
    }: {
      contentTypeSlug: string
      id: string
      locale?: string
    }) =>
      api
        .post<{ status: string }>(
          `/api/content-types/${contentTypeSlug}/documents/${id}/publish`,
          undefined,
          { params: { locale } },
        )
        .then((r) => r.data),
    onSuccess: (_, { contentTypeSlug, id, locale }) => {
      qc.invalidateQueries({
        queryKey: KEYS.detail(contentTypeSlug, id, locale ?? ''),
      })
      qc.invalidateQueries({ queryKey: KEYS.list(contentTypeSlug) })
    },
    onError: onMutationError,
  })
}

export function useUnpublishDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      contentTypeSlug,
      id,
      locale,
    }: {
      contentTypeSlug: string
      id: string
      locale?: string
    }) =>
      api
        .post<{ status: string }>(
          `/api/content-types/${contentTypeSlug}/documents/${id}/unpublish`,
          undefined,
          { params: { locale } },
        )
        .then((r) => r.data),
    onSuccess: (_, { contentTypeSlug, id, locale }) => {
      qc.invalidateQueries({
        queryKey: KEYS.detail(contentTypeSlug, id, locale ?? ''),
      })
      qc.invalidateQueries({ queryKey: KEYS.list(contentTypeSlug) })
    },
    onError: onMutationError,
  })
}
