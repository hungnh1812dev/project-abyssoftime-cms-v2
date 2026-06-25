import { useState } from 'react';
import { Controller, type Control } from 'react-hook-form';
import { MediaLibrary } from '@/components/media/MediaLibrary';
import type { MediaAsset } from '@/types/cms';

interface MediaInputProps {
  name?: string;
  control?: Control;
  ext?: string[];
  'aria-label'?: string;
}

function isMediaAssetObject(value: unknown): value is MediaAsset {
  return typeof value === 'object' && value !== null && 'url' in value;
}

export function MediaInput({ name, control, ext, 'aria-label': ariaLabel }: MediaInputProps) {
  return (
    <Controller
      name={name ?? ''}
      control={control}
      defaultValue={null}
      render={({ field }) => <MediaInputInner field={field} ariaLabel={ariaLabel ?? name} ext={ext} />}
    />
  );
}

interface MediaInputInnerProps {
  field: { value: unknown; onChange: (value: unknown) => void };
  ariaLabel?: string;
  ext?: string[];
}

function MediaInputInner({ field, ariaLabel, ext }: MediaInputInnerProps) {
  const [isLibraryOpen, setIsLibraryOpen] = useState(false);

  const asset = isMediaAssetObject(field.value) ? field.value : null;
  const displayUrl = asset ? (asset.thumbnailUrl || asset.url) : null;
  const displayName = asset?.fileName ?? null;

  function handleSelect(selected: MediaAsset) {
    field.onChange(selected);
  }

  function handleRemove(event: React.MouseEvent) {
    event.stopPropagation();
    field.onChange(null);
  }

  return (
    <>
      <div
        data-testid="media-upload-zone"
        aria-label={ariaLabel}
        role="button"
        tabIndex={0}
        className="border-input hover:border-ring relative cursor-pointer overflow-hidden rounded-md border transition-colors"
        onClick={() => setIsLibraryOpen(true)}
        onKeyDown={(event) => {
          if (event.key === 'Enter' || event.key === ' ') setIsLibraryOpen(true);
        }}>
        {displayUrl ? (
          <>
            <img src={displayUrl} alt={displayName ?? 'media preview'} className="h-auto max-h-40 w-full object-contain" />
            {displayName && <span className="text-muted-foreground block truncate border-t px-2 py-1 text-center text-[11px]">{displayName}</span>}
            <button
              type="button"
              aria-label="Remove image"
              className="bg-background/80 text-muted-foreground hover:text-destructive absolute top-1 right-1 rounded-full p-1 transition-colors"
              onClick={handleRemove}>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round">
                <line x1="18" x2="6" y1="6" y2="18" />
                <line x1="6" x2="18" y1="6" y2="18" />
              </svg>
            </button>
          </>
        ) : (
          <div className="text-muted-foreground flex h-28 flex-col items-center justify-center gap-2">
            <span className="text-2xl">↑</span>
            <span className="text-sm">Click to select media</span>
          </div>
        )}
      </div>
      <MediaLibrary isOpen={isLibraryOpen} onClose={() => setIsLibraryOpen(false)} onSelect={handleSelect} ext={ext} />
    </>
  );
}
