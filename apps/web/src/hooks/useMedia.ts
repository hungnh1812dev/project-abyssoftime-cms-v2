import { useMutation } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { toast } from 'sonner'
import { api } from '@/lib/api'

interface MediaAsset {
  ID: string
  url: string
  thumbnailUrl: string
  publicId: string
  documentRef: string
  contentTypeId: string
}

interface UploadArgs {
  file: File
  documentRef?: string
  contentTypeId?: string
}

export function useUploadMedia() {
  return useMutation({
    mutationFn: ({ file, documentRef = '', contentTypeId = '' }: UploadArgs) => {
      const form = new FormData()
      form.append('file', file, file.name)
      form.append('documentRef', documentRef)
      form.append('contentTypeId', contentTypeId)
      return api.post<MediaAsset>('/api/media/upload', form).then((r) => r.data)
    },
    onError: (err: unknown) => {
      const msg =
        (err as AxiosError<{ error: string }>).response?.data?.error ?? 'Upload failed'
      toast.error(msg)
    },
  })
}
