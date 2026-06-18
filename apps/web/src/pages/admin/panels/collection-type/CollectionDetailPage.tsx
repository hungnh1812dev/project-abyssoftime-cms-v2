import { useParams } from "react-router-dom";
import { useContentTypes } from "@/hooks/useContentTypes";
import { ContentTypePanel } from "../content-type/ContentTypePanel";

export function CollectionDetailPage() {
  const { slug, id } = useParams<{ slug: string; id: string }>();
  const { data: contentTypes = [], isLoading } = useContentTypes();

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>;
  }

  const ct = contentTypes.find((c) => c.Slug === slug);

  if (!ct) {
    return (
      <p className="text-muted-foreground">Content type "{slug}" not found.</p>
    );
  }

  if (!id) {
    return <p className="text-muted-foreground">No document ID provided.</p>;
  }

  return <ContentTypePanel contentType={ct} documentId={id} />;
}
