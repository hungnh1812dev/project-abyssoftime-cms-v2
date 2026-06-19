import { Link, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { useSingleTypeDocument, useSaveSingleType, usePublishSingleType, useUnpublishSingleType } from '@/hooks/useSingleTypeDocuments';
import { useCollectionDocument, useCreateCollectionDocument, useUpdateCollectionDocument, usePublishCollectionDocument, useUnpublishCollectionDocument } from '@/hooks/useCollectionDocuments';
import { useLocales } from '@/hooks/useLocales';
import { api } from '@/lib/api';
import type { Document as CmsDocument, ContentType } from '@/types/cms';
import { useState } from 'react';
import { ContentDetailLayout } from './ContentDetailLayout';
import { ContentTypeBuilder } from './ContentTypeBuilder';

interface Props {
  contentType: ContentType;
  id?: string;
  isNew?: boolean;
}

export function ContentTypePanel({ contentType, id, isNew }: Props) {
  const isSingle = contentType.Kind === 'single';
  const navigate = useNavigate();
  const { data: locales = [] } = useLocales();
  const [locale, setLocale] = useState('');
  const activeLocale = locale || locales[0] || '';

  const singleQuery = useSingleTypeDocument(
    isSingle ? contentType.Slug : '',
    activeLocale,
  );
  const collectionQuery = useCollectionDocument(
    !isSingle && !isNew ? contentType.Slug : '',
    id ?? '',
    activeLocale,
  );

  const saveSingle = useSaveSingleType();
  const createCollection = useCreateCollectionDocument();
  const publishSingle = usePublishSingleType();
  const unpublishSingle = useUnpublishSingleType();

  const updateCollection = useUpdateCollectionDocument();
  const publishCollection = usePublishCollectionDocument();
  const unpublishCollection = useUnpublishCollectionDocument();

  const isLoading = isSingle ? singleQuery.isLoading : (!isNew && collectionQuery.isLoading);
  const doc = isSingle ? singleQuery.data : (isNew ? undefined : collectionQuery.data);

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>;
  }

  if (!doc) {
    const schema = contentType.Fields ?? [];
    const handleFirstSave = isNew
      ? async (data: Record<string, unknown>) => {
          const created = await createCollection.mutateAsync({
            contentTypeSlug: contentType.Slug,
            data,
            locale: activeLocale,
          });
          navigate(
            `/admin/content-type/collection-type/${contentType.Slug}/${created.documentId}?locale=${activeLocale}`,
            { replace: true },
          );
        }
      : async (data: Record<string, unknown>) => {
          await saveSingle.mutateAsync({
            contentTypeSlug: contentType.Slug,
            locale: activeLocale,
            data,
          });
        };

    return (
      <ContentDetailLayout
        title={contentType.Name}
        backLink={
          isNew ? (
            <Link to=".." relative="path" className="text-muted-foreground text-sm hover:underline">
              ← Go back
            </Link>
          ) : undefined
        }
      >
        <ContentTypeBuilder schema={schema} mutationFn={handleFirstSave} />
      </ContentDetailLayout>
    );
  }

  const mutationFn = isSingle
    ? (data: Record<string, unknown>) =>
        saveSingle.mutateAsync({
          contentTypeSlug: contentType.Slug,
          locale: activeLocale,
          data,
        })
    : (data: Record<string, unknown>) =>
        updateCollection.mutateAsync({
          contentTypeSlug: contentType.Slug,
          id: doc.documentId,
          data,
          locale: activeLocale,
        });

  const handlePublish = () => {
    if (isSingle) {
      publishSingle.mutate({ contentTypeSlug: contentType.Slug, locale: activeLocale });
    } else {
      publishCollection.mutate({ contentTypeSlug: contentType.Slug, id: doc.documentId, locale: activeLocale });
    }
  };

  const handleUnpublish = () => {
    if (isSingle) {
      unpublishSingle.mutate({ contentTypeSlug: contentType.Slug, locale: activeLocale });
    } else {
      unpublishCollection.mutate({ contentTypeSlug: contentType.Slug, id: doc.documentId, locale: activeLocale });
    }
  };

  const isPublishing = isSingle ? publishSingle.isPending : publishCollection.isPending;
  const isUnpublishing = isSingle ? unpublishSingle.isPending : unpublishCollection.isPending;

  const canPublish = doc.status !== 'published';
  const canUnpublish = doc.status !== 'draft';
  const schema = contentType.Fields ?? [];

  const apiBase = isSingle
    ? `/api/document-manager/single-type/${contentType.Slug}`
    : `/api/document-manager/collection-type/${contentType.Slug}/${doc.documentId}`;

  return (
    <ContentDetailLayout
      title={contentType.Name}
      status={doc.status}
      backLink={
        id ? (
          <Link to=".." relative="path" className="text-muted-foreground text-sm hover:underline">
            ← Go back
          </Link>
        ) : undefined
      }
      renderActions={() => (
        <>
          {locales.length > 1 && (
            <select aria-label="Locale" value={activeLocale} onChange={(e) => setLocale(e.target.value)}>
              {locales.map((l) => (
                <option key={l} value={l}>
                  {l}
                </option>
              ))}
            </select>
          )}
          {canPublish && (
            <Button onClick={handlePublish} disabled={isPublishing}>
              Publish
            </Button>
          )}
          {canUnpublish && (
            <Button variant="outline" onClick={handleUnpublish} disabled={isUnpublishing}>
              Unpublish
            </Button>
          )}
        </>
      )}>
      <ContentTypeBuilder
        schema={schema}
        query={{
          queryKey: ['documents', isSingle ? 'single-type' : 'collection-type', 'detail', contentType.Slug, doc.documentId, activeLocale, 'data'],
          queryFn: () =>
            api
              .get<CmsDocument>(apiBase, { params: { locale: activeLocale } })
              .then((r) => (r.data as CmsDocument).data),
        }}
        mutationFn={mutationFn}
      />
    </ContentDetailLayout>
  );
}
