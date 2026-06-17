import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useContentTypes } from '@/hooks/useContentTypes'
import { useDocuments, useLocales, usePublishDocument, useUnpublishDocument } from '@/hooks/useDocuments'
import { ContentTypeLayout } from '@/components/content-type/ContentTypeLayout'
import { FormProvider } from '@/components/form/FormProvider'
import { FormField } from '@/components/form/FormField'
import { TextInput } from '@/components/form/inputs/TextInput'
import { Button } from '@/components/ui/button'
import { useCmsFormState } from '@/components/form/FormStateContext'
import { getRegistration } from '@/content-type-registry'
import type { Document } from '@/types/cms'

function SaveButton() {
  const { isDirty, submitting } = useCmsFormState()
  return (
    <Button type="submit" disabled={!isDirty || submitting}>
      Save
    </Button>
  )
}

export function SingleTypePage() {
  const { slug } = useParams<{ slug: string }>()
  const qc = useQueryClient()

  const { data: contentTypes = [], isLoading: ctLoading } = useContentTypes()
  const { data: locales = [] } = useLocales()
  const [locale, setLocale] = useState('')

  const ct = contentTypes.find((c) => c.Slug === slug)
  const activeLocale = locale || locales[0] || ''

  const { data: docs = [], isLoading: docsLoading } = useDocuments(ct?.ID ?? '')
  const publish = usePublishDocument()
  const unpublish = useUnpublishDocument()

  if (ctLoading || docsLoading) {
    return <p className="text-muted-foreground">Loading…</p>
  }

  if (!ct) {
    return <p className="text-muted-foreground">Content type not found.</p>
  }

  const doc = docs.find((d) => d.Locale === activeLocale) ?? docs[0]

  if (!doc) {
    return <p className="text-muted-foreground">No document found for this content type.</p>
  }

  const fieldKeys = Object.keys(doc.Data)
  const canPublish = doc.Status !== 'published'
  const canUnpublish = doc.Status !== 'draft'

  const detailQueryKey = ['documents', 'detail', doc.EntryID, activeLocale, 'data']

  const mutationFn = async (data: Record<string, unknown>) => {
    const result = await api
      .put<Document>(`/api/documents/${doc.EntryID}`, { contentTypeId: ct.ID, data }, { params: { locale: activeLocale } })
      .then((r) => r.data)
    await qc.invalidateQueries({ queryKey: ['documents', ct.ID] })
    return result
  }

  const registration = getRegistration(slug ?? '')
  const Layout = registration?.wrapper ?? ContentTypeLayout

  return (
    <FormProvider
      query={{
        queryKey: detailQueryKey,
        queryFn: () =>
          api
            .get<Document>(`/api/documents/${doc.EntryID}`, { params: { locale: activeLocale } })
            .then((r) => r.data.Data),
      }}
      mutationFn={mutationFn}
    >
      <Layout
        title={ct.Name}
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
                onClick={() => publish.mutate({ id: doc.EntryID, contentTypeId: ct.ID, locale: activeLocale })}
                disabled={publish.isPending}
              >
                Publish
              </Button>
            )}
            {canUnpublish && (
              <Button
                variant="outline"
                onClick={() => unpublish.mutate({ id: doc.EntryID, contentTypeId: ct.ID, locale: activeLocale })}
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
      </Layout>
    </FormProvider>
  )
}
