import { useState } from 'react';
import { Trash2, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog';
import { useMediaList, useUploadMedia, useDeleteMedia } from '@/hooks/useMedia';
import type { MediaAsset } from '@/types/cms';

export function MediaLibraryPage() {
  const [page, setPage] = useState(1);
  const [stagedFiles, setStagedFiles] = useState<File[]>([]);
  const [deleteTarget, setDeleteTarget] = useState<MediaAsset | null>(null);

  const { data, isLoading } = useMediaList(page, 20);
  const upload = useUploadMedia();
  const deleteMedia = useDeleteMedia();

  const items = data?.items ?? [];
  const total = data?.total ?? 0;
  const hasNext = page * 20 < total;
  const hasPrev = page > 1;

  async function handleUpload() {
    for (const file of stagedFiles) {
      await upload.mutateAsync({ file });
    }
    setStagedFiles([]);
  }

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Media Library</h1>
        <div className="flex items-center gap-3">
          <label className="border-muted-foreground/30 bg-muted/30 text-foreground/70 hover:border-primary/50 hover:bg-muted/50 inline-flex cursor-pointer items-center gap-2 rounded-lg border-2 border-dashed px-4 py-2 text-sm font-medium transition-colors">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="17 8 12 3 7 8" />
              <line x1="12" x2="12" y1="3" y2="15" />
            </svg>
            {stagedFiles.length > 0 ? `${stagedFiles.length} file${stagedFiles.length !== 1 ? 's' : ''} selected` : 'Choose files'}
            <input type="file" multiple accept="image/*" onChange={(event) => setStagedFiles(Array.from(event.target.files ?? []))} className="sr-only" />
          </label>
          {stagedFiles.length > 0 && (
            <Button onClick={handleUpload} disabled={upload.isPending} size="sm">
              {upload.isPending ? 'Uploading…' : `Upload ${stagedFiles.length} file${stagedFiles.length !== 1 ? 's' : ''}`}
            </Button>
          )}
        </div>
      </div>

      {stagedFiles.length > 0 && (
        <div className="space-y-3 rounded-lg border p-4">
          <p className="text-muted-foreground text-sm font-medium">Ready to upload</p>
          <div className="grid grid-cols-5 gap-4">
            {stagedFiles.map((file, fileIndex) => (
              <div key={fileIndex} className="group relative">
                <div className="bg-muted relative aspect-square overflow-hidden rounded border">
                  <img
                    src={URL.createObjectURL(file)}
                    alt={file.name}
                    className="h-full w-full object-contain"
                    onLoad={(event) => URL.revokeObjectURL((event.target as HTMLImageElement).src)}
                  />
                  <span className="absolute right-0 bottom-0 left-0 truncate bg-black/60 px-1.5 py-0.5 text-[10px] text-white">{file.name}</span>
                </div>
                <button
                  type="button"
                  aria-label="Remove file"
                  className="bg-background/80 text-muted-foreground hover:text-destructive absolute top-1 right-1 rounded p-1 opacity-0 transition-colors group-hover:opacity-100"
                  onClick={() => setStagedFiles((prev) => prev.filter((_, idx) => idx !== fileIndex))}>
                  <X size={14} />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {isLoading ? (
        <p className="text-muted-foreground">Loading…</p>
      ) : (
        <div className="grid grid-cols-5 gap-4">
          {items.map((asset) => (
            <div key={asset.ID} className="group relative">
              <div className="relative aspect-square overflow-hidden rounded border">
                <img src={asset.thumbnailUrl || asset.url} alt={asset.fileName} className="h-full w-full object-contain" />
                <span className="absolute right-0 bottom-0 left-0 truncate bg-black/60 px-1.5 py-0.5 text-[10px] text-white">{asset.fileName}</span>
              </div>
              <button
                type="button"
                aria-label="Delete asset"
                className="bg-background/80 text-muted-foreground absolute top-1 right-1 rounded p-1 opacity-0 transition-colors group-hover:opacity-100"
                onClick={() => setDeleteTarget(asset)}>
                <Trash2 size={14} />
              </button>
            </div>
          ))}
        </div>
      )}

      <div className="flex items-center justify-between">
        <span className="text-muted-foreground text-sm">
          {total} asset{total !== 1 ? 's' : ''}
        </span>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => setPage((currentPage) => currentPage - 1)} disabled={!hasPrev}>
            Prev
          </Button>
          <Button variant="outline" size="sm" onClick={() => setPage((currentPage) => currentPage + 1)} disabled={!hasNext}>
            Next
          </Button>
        </div>
      </div>

      <Dialog open={deleteTarget !== null} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete media</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete <strong>{deleteTarget?.fileName}</strong>? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              loading={deleteMedia.isPending}
              onClick={() => {
                if (!deleteTarget) return;
                deleteMedia.mutate(deleteTarget.documentId, {
                  onSuccess: () => setDeleteTarget(null),
                });
              }}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
