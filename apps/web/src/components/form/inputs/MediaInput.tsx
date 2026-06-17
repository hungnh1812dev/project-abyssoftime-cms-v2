import { useRef, useState } from "react";
import { Controller, type Control } from "react-hook-form";
import { useUploadMedia } from "@/hooks/useMedia";

interface MediaInputProps {
  name?: string;
  control?: Control;
  documentRef?: string;
  contentTypeId?: string;
}

export function MediaInput({
  name,
  control,
  documentRef,
  contentTypeId,
}: MediaInputProps) {
  const { mutate: upload, isPending } = useUploadMedia();
  const inputRef = useRef<HTMLInputElement>(null);
  const [thumbnailUrl, setThumbnailUrl] = useState<string | null>(null);

  return (
    <Controller
      name={name ?? ""}
      control={control}
      defaultValue={null}
      render={({ field }) => {
        const displayUrl =
          thumbnailUrl || (field.value as string | null) || null;

        return (
          <div className="">
            <input
              ref={inputRef}
              type="file"
              accept="image/*"
              className="hidden"
              aria-label={name}
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (!file) return;
                upload(
                  { file, documentRef, contentTypeId },
                  {
                    onSuccess: (asset) => {
                      field.onChange(asset.url);
                      setThumbnailUrl(asset.thumbnailUrl ?? asset.url);
                    },
                  },
                );
              }}
            />
            <div
              data-testid="media-upload-zone"
              className="min-h-30 max-h-30 cursor-pointer border rounded relative"
              onClick={() => inputRef.current?.click()}
            >
              {isPending && (
                <div
                  data-testid="upload-spinner"
                  className="absolute inset-0 flex items-center justify-center bg-background/60"
                >
                  <span className="text-sm text-muted-foreground">
                    Uploading…
                  </span>
                </div>
              )}
              {displayUrl ? (
                <img
                  src={displayUrl}
                  alt="media preview"
                  className="absolute top-0 left-0  w-full h-full object-contain"
                />
              ) : (
                <div className="flex flex-col items-center justify-center h-full min-h-[7.5em] gap-2 text-muted-foreground">
                  <span className="text-2xl">↑</span>
                  <span className="text-sm">Click to upload</span>
                </div>
              )}
            </div>
          </div>
        );
      }}
    />
  );
}
