import { useState } from 'react'
import { api } from '@/lib/api'
import { ContentDetailLayout } from '../layout/ContentDetailLayout'
import { FormProvider } from '@/components/form/FormProvider'
import { FormField } from '@/components/form/FormField'
import { TextInput } from '@/components/form/inputs/TextInput'
import { Button } from '@/components/ui/button'
import {
  useDocuments,
  useCreateDocument,
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
  const { data: docs = [], isLoading } = useDocuments(contentType.Slug)
  const { data: locales = [] } = useLocales()
  const [locale, setLocale] = useState('')
  const activeLocale = locale || locales[0] || ''

  const doc = docs.find((d) => d.Locale === activeLocale) ?? docs[0]

  const { mutateAsync: createDoc } = useCreateDocument()
  const { mutateAsync: updateDoc } = useUpdateDocument()
  const publish = usePublishDocument()
  const unpublish = useUnpublishDocument()

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>
  }

  if (!doc) {
    const handleFirstSave = async (data: Record<string, unknown>) => {
      await createDoc({ contentTypeSlug: contentType.Slug, data })
    }

    return (
      <ContentDetailLayout title={contentType.Name}>
        <FormProvider mutationFn={handleFirstSave}>
          <div className="space-y-4">
            {(contentType.Fields ?? []).map((field) => (
              <div key={field.name}>
                <label className="block text-sm font-medium mb-1">{field.name}</label>
                <FormField name={field.name}>
                  <TextInput aria-label={field.name} placeholder={field.name} />
                </FormField>
              </div>
            ))}
            <Button type="submit">Save</Button>
          </div>
        </FormProvider>
      </ContentDetailLayout>
    )
  }

  const mutationFn = (data: Record<string, unknown>) =>
    updateDoc({ contentTypeSlug: contentType.Slug, id: doc.DocumentID, data, locale: activeLocale })

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
                  contentTypeSlug: contentType.Slug,
                  id: doc.DocumentID,
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
                  id: doc.DocumentID,
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
          queryKey: ['documents', 'detail', contentType.Slug, doc.DocumentID, activeLocale, 'data'],
          queryFn: () =>
            api
              .get<Document>(`/api/content-types/${contentType.Slug}/documents/${doc.DocumentID}`, {
                params: { locale: activeLocale },
              })
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
