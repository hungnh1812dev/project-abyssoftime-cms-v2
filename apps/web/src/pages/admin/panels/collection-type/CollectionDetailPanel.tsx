import { useState } from "react";
import { Link } from "react-router-dom";
import {
  useQuery,
  useQueryClient,
  keepPreviousData,
} from "@tanstack/react-query";
import { api } from "@/lib/api";
import {
  useLocales,
  usePublishDocument,
  useUnpublishDocument,
} from "@/hooks/useDocuments";
import { ContentDetailLayout } from "../content-type/ContentDetailLayout";
import { FormProvider } from "@/components/form/FormProvider";
import { Button } from "@/components/ui/button";
import { useCmsFormState } from "@/components/form/FormStateContext";
import type { ContentType, Document } from "@/types/cms";
import { renderSchemaField } from "../content-type/renderSchemaField";

interface Props {
  contentType: ContentType;
  documentId: string;
}

function SaveButton() {
  const { isDirty, submitting } = useCmsFormState();
  return (
    <Button type="submit" disabled={!isDirty || submitting}>
      Save
    </Button>
  );
}

export function CollectionDetailPanel({ contentType, documentId }: Props) {
  const qc = useQueryClient();
  const { data: locales = [] } = useLocales();
  const [locale, setLocale] = useState("");
  const activeLocale = locale || locales[0] || "";

  const { data: doc, isLoading } = useQuery({
    queryKey: [
      "documents",
      "detail",
      contentType.Slug,
      documentId,
      activeLocale,
    ],
    queryFn: () =>
      api
        .get<Document>(
          `/api/document-manager/${contentType.Slug}/${documentId}`,
          {
            params: { locale: activeLocale },
          },
        )
        .then((r) => r.data),
    enabled: Boolean(documentId),
    placeholderData: keepPreviousData,
  });
  const publish = usePublishDocument();
  const unpublish = useUnpublishDocument();

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>;
  }

  if (!doc || !doc.data) {
    return <p className="text-muted-foreground">Document not found.</p>;
  }

  const schema = contentType.Fields ?? [];
  const canPublish = doc.status !== "published";
  const canUnpublish = doc.status !== "draft";

  const mutationFn = async (data: Record<string, unknown>) => {
    const result = await api
      .put<Document>(
        `/api/document-manager/${contentType.Slug}/${doc.documentId}`,
        { data },
        { params: { locale: activeLocale } },
      )
      .then((r) => r.data);
    await qc.invalidateQueries({ queryKey: ["documents", contentType.Slug] });
    return result;
  };

  return (
    <FormProvider
      values={doc.data as Record<string, unknown>}
      mutationFn={mutationFn}
    >
      <ContentDetailLayout
        title={contentType.Name}
        status={doc.status}
        backLink={
          <Link
            to={`/admin/content-type/collection-type/${contentType.Slug}`}
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            ← Back
          </Link>
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
                type="button"
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
                type="button"
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
            <SaveButton />
          </>
        )}
      >
        <div className="space-y-4">
          {schema.map((field) => renderSchemaField(field))}
        </div>
      </ContentDetailLayout>
    </FormProvider>
  );
}
