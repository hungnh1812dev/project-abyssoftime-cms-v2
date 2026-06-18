import { useContentTypeBySlug } from "@/hooks/useContentTypes";
import { useParams } from "react-router-dom";
import { CollectionListPage } from "./collection-type/layout/CollectionListPage";
import { ContentTypePanel } from "./ContentTypePanel";

export function ContentTypePage() {
  const { slug } = useParams<{ slug: string }>();
  const { data: contentType, isLoading } = useContentTypeBySlug(slug || "");

  if (!slug) {
    return (
      <p className="text-muted-foreground">No content type slug provided.</p>
    );
  }

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>;
  }

  if (!contentType) {
    return (
      <p className="text-muted-foreground">Content type "{slug}" not found.</p>
    );
  }

  if (contentType.Kind === "single") {
    return <ContentTypePanel contentType={contentType} />;
  }

  return <CollectionListPage contentType={contentType} />;
}
