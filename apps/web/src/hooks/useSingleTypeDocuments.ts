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
  document: (slug: string, locale: string) =>
    ['documents', 'single-type', slug, locale] as const,
}

export function useSingleTypeDocument(slug: string, locale: string) {
  return useQuery({
    queryKey: KEYS.document(slug, locale),
    queryFn: async () => {
      try {
        const res = await api.get<Document>(
          `/api/document-manager/single-type/${slug}`,
          { params: { locale } },
        )
        return res.data
      } catch (err) {
        const status = (err as AxiosError).response?.status
        if (status === 404) return undefined
        throw err
      }
    },
    enabled: Boolean(slug),
    retry: (failureCount, error) => {
      if ((error as AxiosError).response?.status === 404) return false
      return failureCount < 3
    },
  })
}

export function useSaveSingleType() {
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
        .put<Document>(`/api/document-manager/single-type/${contentTypeSlug}`, { data }, { params: { locale } })
        .then((r) => r.data),
    onSuccess: (result) => {
      qc.invalidateQueries({ queryKey: ['documents', 'single-type'] })
      return result
    },
    onError: onMutationError,
  })
}

export function usePublishSingleType() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      contentTypeSlug,
      locale,
    }: {
      contentTypeSlug: string
      locale?: string
    }) =>
      api
        .post<{ status: string }>(
          `/api/document-manager/single-type/${contentTypeSlug}/publish`,
          undefined,
          { params: { locale } },
        )
        .then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['documents', 'single-type'] })
    },
    onError: onMutationError,
  })
}

export function useUnpublishSingleType() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      contentTypeSlug,
      locale,
    }: {
      contentTypeSlug: string
      locale?: string
    }) =>
      api
        .post<{ status: string }>(
          `/api/document-manager/single-type/${contentTypeSlug}/unpublish`,
          undefined,
          { params: { locale } },
        )
        .then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['documents', 'single-type'] })
    },
    onError: onMutationError,
  })
}
