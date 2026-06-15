import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { toast } from 'sonner'
import { api } from '@/lib/api'
import type { ContentType } from '@/types/cms'

function onMutationError(err: unknown) {
  const msg = (err as AxiosError<{ error: string }>).response?.data?.error ?? 'Something went wrong'
  toast.error(msg)
}

const KEYS = {
  all: ['content-types'] as const,
  detail: (id: string) => ['content-types', id] as const,
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

export function useCreateContentType() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { name: string; slug: string; kind: string }) =>
      api.post<ContentType>('/api/content-types', body).then((r) => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
    onError: onMutationError,
  })
}

export function useUpdateContentType() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...body }: { id: string; name: string; slug: string; kind: string }) =>
      api.put<ContentType>(`/api/content-types/${id}`, body).then((r) => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
    onError: onMutationError,
  })
}

export function useDeleteContentType() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete(`/api/content-types/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
    onError: onMutationError,
  })
}
