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
  list: (contentTypeId: string) => ['documents', contentTypeId] as const,
  detail: (id: string, locale: string) => ['documents', 'detail', id, locale] as const,
  locales: ['locales'] as const,
}

export function useDocuments(contentTypeId: string) {
  return useQuery({
    queryKey: KEYS.list(contentTypeId),
    queryFn: () =>
      api
        .get<Document[]>('/api/documents', { params: { contentType: contentTypeId } })
        .then((r) => r.data),
    enabled: Boolean(contentTypeId),
  })
}

export function useDocument(id: string, locale: string) {
  return useQuery({
    queryKey: KEYS.detail(id, locale),
    queryFn: () => api.get<Document>(`/api/documents/${id}`, { params: { locale } }).then((r) => r.data),
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
    mutationFn: (body: { contentTypeId: string; data: Record<string, unknown> }) =>
      api.post<Document>('/api/documents', body).then((r) => r.data),
    onSuccess: (data) => qc.invalidateQueries({ queryKey: KEYS.list(data.ContentTypeID) }),
    onError: onMutationError,
  })
}

export function useUpdateDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      id,
      locale,
      ...body
    }: {
      id: string
      contentTypeId: string
      data: Record<string, unknown>
      locale?: string
    }) => api.put<Document>(`/api/documents/${id}`, body, { params: { locale } }).then((r) => r.data),
    onSuccess: (data) => {
      qc.invalidateQueries({ queryKey: KEYS.list(data.ContentTypeID) })
      qc.invalidateQueries({ queryKey: KEYS.detail(data.EntryID, data.Locale) })
    },
    onError: onMutationError,
  })
}

export function useDeleteDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id }: { id: string; contentTypeId: string }) =>
      api.delete(`/api/documents/${id}`),
    onSuccess: (_, { contentTypeId }) =>
      qc.invalidateQueries({ queryKey: KEYS.list(contentTypeId) }),
    onError: onMutationError,
  })
}

export function usePublishDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, locale }: { id: string; contentTypeId: string; locale?: string }) =>
      api
        .post<{ status: string }>(`/api/documents/${id}/publish`, undefined, { params: { locale } })
        .then((r) => r.data),
    onSuccess: (_, { id, contentTypeId, locale }) => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id, locale ?? '') })
      qc.invalidateQueries({ queryKey: KEYS.list(contentTypeId) })
    },
    onError: onMutationError,
  })
}

export function useUnpublishDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, locale }: { id: string; contentTypeId: string; locale?: string }) =>
      api
        .post<{ status: string }>(`/api/documents/${id}/unpublish`, undefined, { params: { locale } })
        .then((r) => r.data),
    onSuccess: (_, { id, contentTypeId, locale }) => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id, locale ?? '') })
      qc.invalidateQueries({ queryKey: KEYS.list(contentTypeId) })
    },
    onError: onMutationError,
  })
}
