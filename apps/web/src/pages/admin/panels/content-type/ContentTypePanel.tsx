import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import {
  useCreateDocument,
  useDocuments,
  useLocales,
  usePublishDocument,
  useUnpublishDocument,
  useUpdateDocument,
} from "@/hooks/useDocuments";
import { api } from "@/lib/api";
import type { ContentType, Document as CmsDocument } from "@/types/cms";
import { useState } from "react";
import { ContentDetailLayout } from "./ContentDetailLayout";
import { ContentTypeBuilder } from "./ContentTypeBuilder";

interface Props {
  contentType: ContentType;
  id?: string;
}

export function ContentTypePanel({ contentType, id }: Props) {
  const { data: docs = [], isLoading } = useDocuments(contentType.Slug);
  const { data: locales = [] } = useLocales();
  const [locale, setLocale] = useState("");
  const activeLocale = locale || locales[0] || "";

  const doc = id
    ? (docs.find((d) => d.documentId === id) ?? docs[0])
    : (docs.find((d) => d.locale === activeLocale) ?? docs[0]);

  const { mutateAsync: createDoc } = useCreateDocument();
  const { mutateAsync: updateDoc } = useUpdateDocument();
  const publish = usePublishDocument();
  const unpublish = useUnpublishDocument();

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>;
  }

  if (!doc) {
    const schema = contentType.Fields ?? [];
    const handleFirstSave = async (data: Record<string, unknown>) => {
      await createDoc({ contentTypeSlug: contentType.Slug, data });
    };

    return (
      <ContentDetailLayout title={contentType.Name}>
        <ContentTypeBuilder schema={schema} mutationFn={handleFirstSave} />
      </ContentDetailLayout>
    );
  }

  const mutationFn = (data: Record<string, unknown>) =>
    updateDoc({
      contentTypeSlug: contentType.Slug,
      id: doc.documentId,
      data,
      locale: activeLocale,
    });

  const canPublish = doc.status !== "published";
  const canUnpublish = doc.status !== "draft";
  const schema = contentType.Fields ?? [];

  return (
    <ContentDetailLayout
      title={contentType.Name}
      status={doc.status}
      backLink={
        id ? (
          <Link
            to=".."
            relative="path"
            className="text-sm text-muted-foreground hover:underline"
          >
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
                  contentTypeSlug: contentType.Slug,
                  id: doc.documentId,
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
                  contentTypeSlug: contentType.Slug,
                  id: doc.documentId,
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
          queryKey: [
            "documents",
            "detail",
            contentType.Slug,
            doc.documentId,
            activeLocale,
            "data",
          ],
          queryFn: () =>
            api
              .get<CmsDocument>(
                `/api/document-manager/${contentType.Slug}/${doc.documentId}`,
                {
                  params: { locale: activeLocale },
                },
              )
              .then((r) => (r.data as CmsDocument).data),
        }}
        mutationFn={mutationFn}
      />
    </ContentDetailLayout>
  );
}
