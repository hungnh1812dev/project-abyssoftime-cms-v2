import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { useMediaList, useUploadMedia } from '@/hooks/useMedia'
import type { MediaAsset } from '@/types/cms'

export function MediaLibraryPage() {
  const [page, setPage] = useState(1)
  const [stagedFiles, setStagedFiles] = useState<File[]>([])

  const { data, isLoading } = useMediaList(page, 20)
  const upload = useUploadMedia()

  const items = data?.items ?? []
  const total = data?.total ?? 0
  const hasNext = page * 20 < total
  const hasPrev = page > 1

  async function handleUpload() {
    for (const file of stagedFiles) {
      await upload.mutateAsync({ file })
    }
    setStagedFiles([])
  }

  function handleSelect(_asset: MediaAsset) {
    // standalone page — no selection callback
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Media Library</h1>
        <div className="flex items-center gap-3">
          <input
            type="file"
            multiple
            accept="image/*"
            onChange={(e) => setStagedFiles(Array.from(e.target.files ?? []))}
          />
          {stagedFiles.length > 0 && (
            <Button onClick={handleUpload} disabled={upload.isPending} size="sm">
              Upload {stagedFiles.length} file{stagedFiles.length !== 1 ? 's' : ''}
            </Button>
          )}
        </div>
      </div>

      {isLoading ? (
        <p className="text-muted-foreground">Loading…</p>
      ) : (
        <div className="grid grid-cols-5 gap-4">
          {items.map((asset) => (
            <div
              key={asset.ID}
              className="border rounded overflow-hidden aspect-square cursor-pointer hover:border-ring transition-colors"
              onClick={() => handleSelect(asset)}
            >
              <img
                src={asset.thumbnailUrl || asset.url}
                alt={asset.fileName}
                className="w-full h-full object-cover"
              />
            </div>
          ))}
        </div>
      )}

      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">{total} asset{total !== 1 ? 's' : ''}</span>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => setPage((p) => p - 1)} disabled={!hasPrev}>
            Prev
          </Button>
          <Button variant="outline" size="sm" onClick={() => setPage((p) => p + 1)} disabled={!hasNext}>
            Next
          </Button>
        </div>
      </div>
    </div>
  )
}
