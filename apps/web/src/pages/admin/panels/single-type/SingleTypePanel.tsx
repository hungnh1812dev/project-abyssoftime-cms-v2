import { useState } from 'react'
import { api } from '@/lib/api'
import { ContentDetailLayout } from '../layout/ContentDetailLayout'
import { FormProvider } from '@/components/form/FormProvider'
import { FormField } from '@/components/form/FormField'
import { TextInput } from '@/components/form/inputs/TextInput'
import { Button } from '@/components/ui/button'
import {
  useDocuments,
  useLocales,
  useUpdateDocument,
  usePublishDocument,
  useUnpublishDocument,
} from '@/hooks/useDocuments'
import type { ContentType, Document } from '@/types/cms'

interface Props {
  contentType: ContentType
}

export function SingleTypePanel({ contentType }: Props) {
  const { data: docs = [], isLoading } = useDocuments(contentType.ID)
  const { data: locales = [] } = useLocales()
  const [locale, setLocale] = useState('')
  const activeLocale = locale || locales[0] || ''

  const doc = docs.find((d) => d.Locale === activeLocale) ?? docs[0]

  const { mutateAsync: updateDoc } = useUpdateDocument()
  const publish = usePublishDocument()
  const unpublish = useUnpublishDocument()

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>
  }

  if (!doc) {
    return <p className="text-muted-foreground">No document found for this content type.</p>
  }

  const mutationFn = (data: Record<string, unknown>) =>
    updateDoc({ id: doc.EntryID, contentTypeId: contentType.ID, data, locale: activeLocale })

  const fieldKeys = Object.keys(doc.Data)
  const canPublish = doc.Status !== 'published'
  const canUnpublish = doc.Status !== 'draft'

  return (
    <ContentDetailLayout
      title={contentType.Name}
      status={doc.Status}
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
      <FormProvider
        query={{
          queryKey: ['documents', 'detail', doc.EntryID, activeLocale, 'data'],
          queryFn: () =>
            api
              .get<Document>(`/api/documents/${doc.EntryID}`, { params: { locale: activeLocale } })
              .then((r) => r.data.Data),
        }}
        mutationFn={mutationFn}
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
          <Button type="submit">Save</Button>
        </div>
      </FormProvider>
    </ContentDetailLayout>
  )
}
