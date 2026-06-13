import { useMutation } from '@tanstack/react-query'
import { api } from '@/lib/api'

interface MediaAsset {
  ID: string
  url: string
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
  })
}
