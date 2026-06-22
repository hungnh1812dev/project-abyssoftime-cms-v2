import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { toast } from 'sonner'
import { api } from '@/lib/api'
import type { MediaAsset } from '@/types/cms'

interface MediaListResponse {
  items: MediaAsset[]
  total: number
  page: number
  limit: number
}

interface UploadArgs {
  file: File
}

export function useMediaList(page: number, limit: number) {
  return useQuery<MediaListResponse>({
    queryKey: ['media', 'list', page, limit],
    queryFn: () =>
      api.get<MediaListResponse>(`/api/media?page=${page}&limit=${limit}`).then((response) => response.data),
  })
}

export function useUploadMedia() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ file }: UploadArgs) => {
      const form = new FormData()
      form.append('file', file, file.name)
      return api.post<MediaAsset>('/api/media/upload', form).then((response) => response.data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['media', 'list'] })
    },
    onError: (error: unknown) => {
      const msg =
        (error as AxiosError<{ error: string }>).response?.data?.error ?? 'Upload failed'
      toast.error(msg)
    },
  })
}

export function useDeleteMedia() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete(`/api/media/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['media', 'list'] })
    },
    onError: (error: unknown) => {
      const msg =
        (error as AxiosError<{ error: string }>).response?.data?.error ?? 'Delete failed'
      toast.error(msg)
    },
  })
}
