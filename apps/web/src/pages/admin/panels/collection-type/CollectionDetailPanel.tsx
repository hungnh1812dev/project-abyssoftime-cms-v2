import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useLocales, usePublishDocument, useUnpublishDocument } from '@/hooks/useDocuments'
import { ContentDetailLayout } from '../layout/ContentDetailLayout'
import { FormProvider } from '@/components/form/FormProvider'
import { FormField } from '@/components/form/FormField'
import { TextInput } from '@/components/form/inputs/TextInput'
import { Button } from '@/components/ui/button'
import { useCmsFormState } from '@/components/form/FormStateContext'
import type { ContentType, Document } from '@/types/cms'

interface Props {
  contentType: ContentType
  documentId: string
}

function SaveButton() {
  const { isDirty, submitting } = useCmsFormState()
  return (
    <Button type="submit" disabled={!isDirty || submitting}>
      Save
    </Button>
  )
}

export function CollectionDetailPanel({ contentType, documentId }: Props) {
  const qc = useQueryClient()
  const { data: locales = [] } = useLocales()
  const [locale, setLocale] = useState('')
  const activeLocale = locale || locales[0] || ''

  const { data: doc, isLoading } = useQuery({
    queryKey: ['documents', 'detail', documentId, activeLocale],
    queryFn: () =>
      api
        .get<Document>(`/api/documents/${documentId}`, { params: { locale: activeLocale } })
        .then((r) => r.data),
    enabled: Boolean(documentId),
    placeholderData: keepPreviousData,
  })
  const publish = usePublishDocument()
  const unpublish = useUnpublishDocument()

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>
  }

  if (!doc || !doc.Data) {
    return <p className="text-muted-foreground">Document not found.</p>
  }

  const fieldKeys = Object.keys(doc.Data)
  const canPublish = doc.Status !== 'published'
  const canUnpublish = doc.Status !== 'draft'

  const mutationFn = async (data: Record<string, unknown>) => {
    const result = await api
      .put<Document>(
        `/api/documents/${doc.EntryID}`,
        { contentTypeId: contentType.ID, data },
        { params: { locale: activeLocale } },
      )
      .then((r) => r.data)
    await qc.invalidateQueries({ queryKey: ['documents', contentType.ID] })
    return result
  }

  return (
    <FormProvider
      values={doc.Data as Record<string, unknown>}
      mutationFn={mutationFn}
    >
      <ContentDetailLayout
        title={contentType.Name}
        status={doc.Status}
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
                  publish.mutate({ id: doc.EntryID, contentTypeId: contentType.ID, locale: activeLocale })
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
                  unpublish.mutate({ id: doc.EntryID, contentTypeId: contentType.ID, locale: activeLocale })
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
          {fieldKeys.map((key) => (
            <div key={key}>
              <label className="block text-sm font-medium mb-1">{key}</label>
              <FormField name={key}>
                <TextInput aria-label={key} placeholder={key} />
              </FormField>
            </div>
          ))}
        </div>
      </ContentDetailLayout>
    </FormProvider>
  )
}
