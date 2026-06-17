import { useState } from 'react'
import { useMediaList, useUploadMedia } from '@/hooks/useMedia'
import { Button } from '@/components/ui/button'
import type { MediaAsset } from '@/types/cms'

interface MediaLibraryProps {
  isOpen: boolean
  onClose: () => void
  onSelect: (asset: MediaAsset) => void
  ext?: string[]
}

export function MediaLibrary({ isOpen, onClose, onSelect, ext }: MediaLibraryProps) {
  const [page, setPage] = useState(1)
  const [showUpload, setShowUpload] = useState(false)
  const [stagedFiles, setStagedFiles] = useState<File[]>([])

  const { data, isLoading } = useMediaList(page, 20)
  const upload = useUploadMedia()

  if (!isOpen) return null

  const items = data?.items ?? []
  const total = data?.total ?? 0
  const hasNext = page * 20 < total
  const hasPrev = page > 1

  const filteredItems = ext
    ? items.filter((a) => ext.includes(a.fileExt))
    : items

  async function handleUpload() {
    for (const file of stagedFiles) {
      await upload.mutateAsync({ file })
    }
    setStagedFiles([])
    setShowUpload(false)
  }

  return (
    <div
      role="dialog"
      aria-modal="true"
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={(e) => { if (e.target === e.currentTarget) onClose() }}
    >
      <div className="bg-background rounded-lg shadow-lg w-[720px] max-h-[80vh] flex flex-col overflow-hidden">
        <div className="flex items-center justify-between p-4 border-b">
          <h2 className="font-semibold">Media Library</h2>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => setShowUpload(!showUpload)}>
              Upload More
            </Button>
            <Button variant="outline" size="sm" onClick={onClose}>
              Close
            </Button>
          </div>
        </div>

        {showUpload && (
          <div className="p-4 border-b space-y-2">
            <input
              type="file"
              multiple
              accept={ext ? ext.map((e) => `.${e}`).join(',') : 'image/*'}
              onChange={(e) => setStagedFiles(Array.from(e.target.files ?? []))}
            />
            {stagedFiles.length > 0 && (
              <Button size="sm" onClick={handleUpload} disabled={upload.isPending}>
                Start Upload
              </Button>
            )}
          </div>
        )}

        <div className="flex-1 overflow-y-auto p-4">
          {isLoading ? (
            <p className="text-muted-foreground text-sm">Loading…</p>
          ) : (
            <div className="grid grid-cols-4 gap-3">
              {filteredItems.map((asset) => (
                <button
                  key={asset.ID}
                  type="button"
                  className={`border rounded overflow-hidden aspect-square hover:border-ring transition-colors ${
                    ext && !ext.includes(asset.fileExt) ? 'opacity-40' : ''
                  }`}
                  onClick={() => { onSelect(asset); onClose() }}
                >
                  <img
                    src={asset.thumbnailUrl || asset.url}
                    alt={asset.fileName}
                    className="w-full h-full object-cover"
                  />
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="flex items-center justify-between p-4 border-t">
          <span className="text-sm text-muted-foreground">
            {total} asset{total !== 1 ? 's' : ''}
          </span>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => p - 1)}
              disabled={!hasPrev}
              aria-label="Previous page"
            >
              Prev
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => p + 1)}
              disabled={!hasNext}
              aria-label="Next page"
            >
              Next
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
