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
  documentRef?: string
  contentTypeId?: string
}

export function useMediaList(page: number, limit: number) {
  return useQuery<MediaListResponse>({
    queryKey: ['media', 'list', page, limit],
    queryFn: () =>
      api.get<MediaListResponse>(`/api/media?page=${page}&limit=${limit}`).then((r) => r.data),
  })
}

export function useUploadMedia() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ file, documentRef = '', contentTypeId = '' }: UploadArgs) => {
      const form = new FormData()
      form.append('file', file, file.name)
      form.append('documentRef', documentRef)
      form.append('contentTypeId', contentTypeId)
      return api.post<MediaAsset>('/api/media/upload', form).then((r) => r.data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['media', 'list'] })
    },
    onError: (err: unknown) => {
      const msg =
        (err as AxiosError<{ error: string }>).response?.data?.error ?? 'Upload failed'
      toast.error(msg)
    },
  })
}
