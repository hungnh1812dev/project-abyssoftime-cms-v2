import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { toast } from 'sonner'
import { api } from '@/lib/api'
import type { Document, PaginatedResponse } from '@/types/cms'

function onMutationError(err: unknown) {
  const msg = (err as AxiosError<{ error: string }>).response?.data?.error ?? 'Something went wrong'
  toast.error(msg)
}

const KEYS = {
  list: (slug: string) => ['documents', 'collection-type', slug] as const,
  detail: (slug: string, id: string, locale: string) =>
    ['documents', 'collection-type', 'detail', slug, id, locale] as const,
}

export function useCollectionDocuments(slug: string, start: number, size: number, locale: string, orderBy: string = 'id', sortDir: 'asc' | 'desc' = 'desc') {
  return useQuery({
    queryKey: [...KEYS.list(slug), start, size, locale, orderBy, sortDir] as const,
    queryFn: () =>
      api
        .get<PaginatedResponse<Document>>(`/api/document-manager/collection-type/${slug}`, {
          params: { start, size, locale, orderBy, sortDir },
        })
        .then((r) => r.data),
    enabled: Boolean(slug),
  })
}

export function useCollectionDocument(slug: string, documentId: string, locale: string) {
  return useQuery({
    queryKey: KEYS.detail(slug, documentId, locale),
    queryFn: () =>
      api
        .get<Document>(`/api/document-manager/collection-type/${slug}/${documentId}`, {
          params: { locale },
        })
        .then((r) => r.data),
    enabled: Boolean(documentId),
  })
}

export function useCreateCollectionDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      contentTypeSlug,
      locale,
      data,
    }: {
      contentTypeSlug: string
      locale?: string
      data: Record<string, unknown>
    }) =>
      api
        .post<Document>(`/api/document-manager/collection-type/${contentTypeSlug}`, { data }, { params: { locale } })
        .then((r) => r.data),
    onSuccess: (_, { contentTypeSlug }) =>
      qc.invalidateQueries({ queryKey: KEYS.list(contentTypeSlug) }),
    onError: onMutationError,
  })
}

export function useUpdateCollectionDocument() {
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
        .put<Document>(
          `/api/document-manager/collection-type/${contentTypeSlug}/${id}`,
          { data },
          { params: { locale } },
        )
        .then((r) => r.data),
    onSuccess: (result, { contentTypeSlug }) => {
      qc.invalidateQueries({ queryKey: KEYS.list(contentTypeSlug) })
      qc.invalidateQueries({
        queryKey: KEYS.detail(contentTypeSlug, result.data.documentId as string, result.data.locale as string),
      })
    },
    onError: onMutationError,
  })
}

export function useDeleteCollectionDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      contentTypeSlug,
      id,
    }: {
      contentTypeSlug: string
      id: string
    }) => api.delete(`/api/document-manager/collection-type/${contentTypeSlug}/${id}`),
    onSuccess: (_, { contentTypeSlug }) =>
      qc.invalidateQueries({ queryKey: KEYS.list(contentTypeSlug) }),
    onError: onMutationError,
  })
}

export function usePublishCollectionDocument() {
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
          `/api/document-manager/collection-type/${contentTypeSlug}/${id}/publish`,
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

export function useUnpublishCollectionDocument() {
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
          `/api/document-manager/collection-type/${contentTypeSlug}/${id}/unpublish`,
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
