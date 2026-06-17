import { useRef, useState } from 'react'
import { Controller, type Control } from 'react-hook-form'
import { useUploadMedia } from '@/hooks/useMedia'
import { Button } from '@/components/ui/button'

interface MediaInputProps {
  name?: string
  control?: Control
  documentRef?: string
  contentTypeId?: string
}

export function MediaInput({ name, control, documentRef, contentTypeId }: MediaInputProps) {
  const { mutate: upload, isPending } = useUploadMedia()
  const inputRef = useRef<HTMLInputElement>(null)
  const [thumbnailUrl, setThumbnailUrl] = useState('')

  return (
    <Controller
      name={name ?? ''}
      control={control}
      defaultValue=""
      render={({ field }) => (
        <div className="flex flex-col gap-2">
          <input
            ref={inputRef}
            type="file"
            accept="image/*"
            className="hidden"
            aria-label={name}
            onChange={(e) => {
              const file = e.target.files?.[0]
              if (!file) return
              upload(
                { file, documentRef, contentTypeId },
                {
                  onSuccess: (asset) => {
                    field.onChange(asset.url)
                    setThumbnailUrl(asset.thumbnailUrl ?? asset.url)
                  },
                },
              )
            }}
          />
          <Button
            type="button"
            variant="outline"
            disabled={isPending}
            onClick={() => inputRef.current?.click()}
          >
            {isPending ? 'Uploading…' : 'Choose file'}
          </Button>
          {field.value && (
            <img
              src={field.value as string}
              alt="uploaded media"
              className="max-h-40 rounded border object-contain"
            />
          )}
          {thumbnailUrl && thumbnailUrl !== (field.value as string) && (
            <img
              src={thumbnailUrl}
              alt="thumbnail preview"
              className="max-h-20 rounded border object-contain"
            />
          )}
        </div>
      )}
    />
  )
}
