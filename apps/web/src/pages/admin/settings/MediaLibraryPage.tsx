import { useState } from 'react'
import { Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useMediaList, useUploadMedia, useDeleteMedia } from '@/hooks/useMedia'

export function MediaLibraryPage() {
  const [page, setPage] = useState(1)
  const [stagedFiles, setStagedFiles] = useState<File[]>([])

  const { data, isLoading } = useMediaList(page, 20)
  const upload = useUploadMedia()
  const deleteMedia = useDeleteMedia()
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null)

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

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Media Library</h1>
        <div className="flex items-center gap-3">
          <label className="inline-flex cursor-pointer items-center gap-2 rounded-lg border-2 border-dashed border-muted-foreground/30 bg-muted/30 px-4 py-2 text-sm font-medium text-foreground/70 transition-colors hover:border-primary/50 hover:bg-muted/50">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" x2="12" y1="3" y2="15"/></svg>
            {stagedFiles.length > 0
              ? `${stagedFiles.length} file${stagedFiles.length !== 1 ? 's' : ''} selected`
              : 'Choose files'}
            <input
              type="file"
              multiple
              accept="image/*"
              onChange={(e) => setStagedFiles(Array.from(e.target.files ?? []))}
              className="sr-only"
            />
          </label>
          {stagedFiles.length > 0 && (
            <Button onClick={handleUpload} disabled={upload.isPending} size="sm">
              {upload.isPending ? 'Uploading…' : `Upload ${stagedFiles.length} file${stagedFiles.length !== 1 ? 's' : ''}`}
            </Button>
          )}
        </div>
      </div>

      {isLoading ? (
        <p className="text-muted-foreground">Loading…</p>
      ) : (
        <div className="grid grid-cols-5 gap-4">
          {items.map((asset) => (
            <div key={asset.ID} className="relative group">
              <div className="border rounded overflow-hidden aspect-square">
                <img
                  src={asset.thumbnailUrl || asset.url}
                  alt={asset.fileName}
                  className="w-full h-full object-cover"
                />
              </div>
              <button
                type="button"
                aria-label={pendingDeleteId === asset.ID ? 'Confirm delete' : 'Delete asset'}
                className={`absolute top-1 right-1 rounded p-1 bg-background/80 transition-colors ${
                  pendingDeleteId === asset.ID
                    ? 'text-destructive'
                    : 'text-muted-foreground opacity-0 group-hover:opacity-100'
                }`}
                onClick={() => {
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
