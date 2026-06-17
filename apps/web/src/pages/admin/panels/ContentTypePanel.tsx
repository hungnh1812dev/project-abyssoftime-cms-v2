import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import {
  useDocuments,
  useLocales,
  usePublishDocument,
  useUnpublishDocument,
  useUpdateDocument,
} from "@/hooks/useDocuments";
import { api } from "@/lib/api";
import type { ContentType, Document as CmsDocument } from "@/types/cms";
import { useState } from "react";
import { ContentDetailLayout } from "./layout/ContentDetailLayout";
import { ContentTypeBuilder } from "./ContentTypeBuilder";

interface Props {
  contentType: ContentType;
  id?: string;
}

export function ContentTypePanel({ contentType, id }: Props) {
  const { data: docs = [], isLoading } = useDocuments(contentType.ID);
  const { data: locales = [] } = useLocales();
  const [locale, setLocale] = useState("");
  const activeLocale = locale || locales[0] || "";

  const doc = id
    ? docs.find((d) => d.EntryID === id) ?? docs[0]
    : docs.find((d) => d.Locale === activeLocale) ?? docs[0];

  const { mutateAsync: updateDoc } = useUpdateDocument();
  const publish = usePublishDocument();
  const unpublish = useUnpublishDocument();

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>;
  }

  if (!doc) {
    return (
      <p className="text-muted-foreground">
        No document found for this content type.
      </p>
    );
  }

  const mutationFn = (data: Record<string, unknown>) =>
    updateDoc({
      id: doc.EntryID,
      contentTypeId: contentType.ID,
      data,
      locale: activeLocale,
    });

  const canPublish = doc.Status !== "published";
  const canUnpublish = doc.Status !== "draft";
  const schema = contentType.Fields ?? [];

  return (
    <ContentDetailLayout
      title={contentType.Name}
      status={doc.Status}
      backLink={
        id ? (
          <Link to=".." relative="path" className="text-sm text-muted-foreground hover:underline">
            ← Go back
          </Link>
        ) : undefined
      }
      renderActions={() => (
        <>
          {locales.length > 1 && (
            <select
              aria-label="Locale"
              value={activeLocale}
              onChange={(e) => setLocale(e.target.value)}
            >
              {locales.map((l) => (
                <option key={l} value={l}>
                  {l}
                </option>
              ))}
            </select>
          )}
          {canPublish && (
            <Button
              onClick={() =>
                publish.mutate({
                  id: doc.EntryID,
                  contentTypeId: contentType.ID,
                  locale: activeLocale,
                })
              }
              disabled={publish.isPending}
            >
              Publish
            </Button>
          )}
          {canUnpublish && (
            <Button
              variant="outline"
              onClick={() =>
                unpublish.mutate({
                  id: doc.EntryID,
                  contentTypeId: contentType.ID,
                  locale: activeLocale,
                })
              }
              disabled={unpublish.isPending}
            >
              Unpublish
            </Button>
          )}
        </>
      )}
    >
      <ContentTypeBuilder
        schema={schema}
        query={{
          queryKey: ["documents", "detail", doc.EntryID, activeLocale, "data"],
          queryFn: () =>
            api
              .get<CmsDocument>(`/api/documents/${doc.EntryID}`, {
                params: { locale: activeLocale },
              })
              .then((r) => (r.data as CmsDocument).Data),
        }}
        mutationFn={mutationFn}
      />
    </ContentDetailLayout>
  );
}
