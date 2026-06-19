import { useParams } from "react-router-dom";
import { useContentTypeBySlug } from "@/hooks/useContentTypes";
import { ContentTypePanel } from "../content-type/ContentTypePanel";

export function CollectionDetailPage() {
  const { slug, id } = useParams<{ slug: string; id: string }>();
  const { data: contentType, isLoading } = useContentTypeBySlug(slug || '');

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>;
  }

  if (!contentType) {
    return (
      <p className="text-muted-foreground">Content type "{slug}" not found.</p>
    );
  }

  return <ContentTypePanel contentType={contentType} id={id} isNew={!id} />;
}
