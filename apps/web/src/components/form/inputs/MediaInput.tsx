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
              className="cursor-pointer border border-input rounded-md transition-colors hover:border-ring relative min-h-28 max-h-28 overflow-hidden"
              onClick={() => setIsLibraryOpen(true)}
              onKeyDown={(e) => {
                if (e.key === "Enter" || e.key === " ") setIsLibraryOpen(true);
              }}
            >
              {displayUrl ? (
                <>
                  <img
                    src={displayUrl}
                    alt="media preview"
                    className="w-full max-h-28 object-contain inline-block"
                  />
                  <button
                    type="button"
                    aria-label="Remove image"
                    className="absolute top-1 right-1 rounded-full bg-background/80 p-1 text-muted-foreground hover:text-destructive transition-colors"
                    onClick={(e) => {
                      e.stopPropagation();
                      field.onChange(null);
                    }}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/></svg>
                  </button>
                </>
              ) : (
                <div className="flex flex-col items-center justify-center min-h-28 gap-2 text-muted-foreground">
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
