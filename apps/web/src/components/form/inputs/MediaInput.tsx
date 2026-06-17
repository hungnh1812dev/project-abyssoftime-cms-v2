import { useState } from "react";
import { Controller, type Control } from "react-hook-form";
import { MediaLibrary } from "@/components/media/MediaLibrary";
import type { MediaAsset } from "@/types/cms";

interface MediaInputProps {
  name?: string;
  control?: Control;
  documentRef?: string;
  contentTypeId?: string;
  ext?: string[];
  "aria-label"?: string;
}

export function MediaInput({
  name,
  control,
  ext,
  "aria-label": ariaLabel,
}: MediaInputProps) {
  const [isLibraryOpen, setIsLibraryOpen] = useState(false);

  return (
    <Controller
      name={name ?? ""}
      control={control}
      defaultValue={null}
      render={({ field }) => {
        const displayUrl = (field.value as string | null) || null;

        function handleSelect(asset: MediaAsset) {
          field.onChange(asset.thumbnailUrl || asset.url);
        }

        return (
          <>
            <div
              data-testid="media-upload-zone"
              aria-label={ariaLabel ?? name}
              role="button"
              tabIndex={0}
              className="cursor-pointer border border-input rounded-md transition-colors hover:border-ring relative"
              onClick={() => setIsLibraryOpen(true)}
              onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') setIsLibraryOpen(true) }}
            >
              {displayUrl ? (
                <img
                  src={displayUrl}
                  alt="media preview"
                  className="w-full h-auto object-contain"
                />
              ) : (
                <div className="flex flex-col items-center justify-center min-h-[7.5rem] gap-2 text-muted-foreground">
                  <span className="text-2xl">↑</span>
                  <span className="text-sm">Click to select media</span>
                </div>
              )}
            </div>
            <MediaLibrary
              isOpen={isLibraryOpen}
              onClose={() => setIsLibraryOpen(false)}
              onSelect={handleSelect}
              ext={ext}
            />
          </>
        );
      }}
    />
  );
}
