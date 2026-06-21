import { useState } from 'react'
import { Trash2 } from 'lucide-react'
import { useMediaList, useUploadMedia, useDeleteMedia } from '@/hooks/useMedia'
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
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null)

  const { data, isLoading } = useMediaList(page, 20)
  const upload = useUploadMedia()
  const deleteMedia = useDeleteMedia()

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
          <div className="p-4 border-b space-y-3">
            <label className="flex cursor-pointer items-center justify-center gap-2 rounded-lg border-2 border-dashed border-muted-foreground/30 bg-muted/30 px-4 py-3 text-sm font-medium text-foreground/70 transition-colors hover:border-primary/50 hover:bg-muted/50">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" x2="12" y1="3" y2="15"/></svg>
              {stagedFiles.length > 0
                ? `${stagedFiles.length} file${stagedFiles.length !== 1 ? 's' : ''} selected`
                : 'Choose files to upload'}
              <input
                type="file"
                multiple
                accept={ext ? ext.map((e) => `.${e}`).join(',') : 'image/*'}
                onChange={(e) => setStagedFiles(Array.from(e.target.files ?? []))}
                className="sr-only"
              />
            </label>
            {stagedFiles.length > 0 && (
              <>
                <div className="grid grid-cols-4 gap-2">
                  {stagedFiles.map((file, i) => (
                    <div key={i} className="relative rounded-lg border overflow-hidden aspect-square bg-muted">
                      <img
                        src={URL.createObjectURL(file)}
                        alt={file.name}
                        className="w-full h-full object-cover"
                        onLoad={(e) => URL.revokeObjectURL((e.target as HTMLImageElement).src)}
                      />
                      <button
                        type="button"
                        className="absolute top-1 right-1 rounded-full bg-background/80 p-0.5 text-muted-foreground hover:text-destructive transition-colors"
                        onClick={() => setStagedFiles((prev) => prev.filter((_, j) => j !== i))}
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/></svg>
                      </button>
                      <span className="absolute bottom-0 left-0 right-0 bg-black/50 text-white text-[10px] px-1 py-0.5 truncate">{file.name}</span>
                    </div>
                  ))}
                </div>
                <Button size="sm" className="w-full" onClick={handleUpload} disabled={upload.isPending}>
                  {upload.isPending ? 'Uploading…' : `Upload ${stagedFiles.length} file${stagedFiles.length !== 1 ? 's' : ''}`}
                </Button>
              </>
            )}
          </div>
        )}

        <div className="flex-1 overflow-y-auto p-4">
          {isLoading ? (
            <p className="text-muted-foreground text-sm">Loading…</p>
          ) : (
            <div className="grid grid-cols-4 gap-3">
              {filteredItems.map((asset) => (
                <div key={asset.ID} className="relative group">
                  <button
                    type="button"
                    className={`w-full rounded-lg border-2 overflow-hidden aspect-square transition-all ${
                      ext && !ext.includes(asset.fileExt)
                        ? 'opacity-40 border-muted'
                        : 'border-border hover:border-primary hover:ring-2 hover:ring-primary/20 hover:shadow-md'
                    }`}
                    disabled={deleteMedia.isPending}
                    onClick={() => { onSelect(asset); onClose() }}
                  >
                    <img
                      src={asset.thumbnailUrl || asset.url}
                      alt={asset.fileName}
                      className="w-full h-full object-cover"
                    />
                    <span className="absolute bottom-0 left-0 right-0 bg-black/60 text-white text-[10px] px-1.5 py-0.5 truncate">
                      {asset.fileName}
                    </span>
                  </button>
                  <button
                    type="button"
                    aria-label={pendingDeleteId === asset.ID ? 'Confirm delete' : 'Delete asset'}
                    className={`absolute top-1 right-1 rounded p-0.5 bg-background/80 transition-colors ${
                      pendingDeleteId === asset.ID
                        ? 'text-red-500'
                        : 'text-muted-foreground opacity-0 group-hover:opacity-100'
                    }`}
                    onClick={(e) => {
                      e.stopPropagation()
                      if (pendingDeleteId === asset.ID) {
                        deleteMedia.mutate(asset.ID)
                        setPendingDeleteId(null)
                      } else {
                        setPendingDeleteId(asset.ID)
                      }
                    }}
                    onMouseLeave={() => setPendingDeleteId(null)}
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
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
