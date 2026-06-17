import { useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '@/lib/api'
import { FormProvider } from '@/components/form/FormProvider'
import { FormField } from '@/components/form/FormField'
import { TextInput } from '@/components/form/inputs/TextInput'
import { Button } from '@/components/ui/button'
import {
  useDocument,
  useLocales,
  useUpdateDocument,
  usePublishDocument,
  useUnpublishDocument,
} from '@/hooks/useDocuments'
import type { ContentType, Document } from '@/types/cms'

interface Props {
  contentType: ContentType
  documentId: string
}

export function CollectionDetailPanel({ contentType, documentId }: Props) {
  const { data: locales = [] } = useLocales()
  const [locale, setLocale] = useState('')
  const activeLocale = locale || locales[0] || ''

  const { data: doc, isLoading } = useDocument(documentId, activeLocale)

  const { mutateAsync: updateDoc } = useUpdateDocument()
  const publish = usePublishDocument()
  const unpublish = useUnpublishDocument()

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>
  }

  if (!doc) {
    return <p className="text-muted-foreground">Document not found.</p>
  }

  const mutationFn = (data: Record<string, unknown>) =>
    updateDoc({ id: doc.EntryID, contentTypeId: contentType.ID, data, locale: activeLocale })

  const fieldKeys = Object.keys(doc.Data)
  const canPublish = doc.Status !== 'published'
  const canUnpublish = doc.Status !== 'draft'

  return (
    <div className="space-y-6">
      <Link
        to={`/admin/content-types/${contentType.Slug}`}
        className="text-sm text-muted-foreground hover:text-foreground"
      >
        ← Back
      </Link>

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">{contentType.Name}</h1>
          <span className="text-sm text-muted-foreground capitalize">{doc.Status}</span>
        </div>
        <div className="flex gap-2 items-center">
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
        </div>
      </div>

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
    </div>
  )
}
