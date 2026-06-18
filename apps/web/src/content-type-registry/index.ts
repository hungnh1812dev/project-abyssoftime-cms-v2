import { type ComponentType } from "react";
import type { ContentTypeLayoutProps } from "@/components/content-type/ContentTypeLayout";

export interface CollectionColumnDef {
  key: string;
  label: string;
  type: "text" | "boolean" | "number" | "image";
}

export interface ContentTypeRegistration {
  slug: string;
  kind: "single" | "collection";
  columns?: CollectionColumnDef[];
  wrapper?: ComponentType<ContentTypeLayoutProps>;
}

export const contentTypeRegistry: ContentTypeRegistration[] = [
  {
    slug: "blog-posts",
    kind: "collection",
    columns: [
      { key: "title", label: "Title", type: "text" },
      { key: "slug", label: "Slug", type: "text" },
      { key: "coverImage", label: "Cover", type: "image" },
      { key: "featured", label: "Featured", type: "boolean" },
    ],
  },
];

export function getRegistration(
  slug: string,
): ContentTypeRegistration | undefined {
  return contentTypeRegistry.find((r) => r.slug === slug);
}
