import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import {
  useDocument,
  useLocales,
  usePublishDocument,
  useUnpublishDocument,
} from '@/hooks/useDocuments'
import { FormProvider } from '@/components/form/FormProvider'
import { FormField } from '@/components/form/FormField'
import { TextInput } from '@/components/form/inputs/TextInput'
import { RichTextInput } from '@/components/form/inputs/RichTextInput'
import { MediaInput } from '@/components/form/inputs/MediaInput'
import { NumberInput } from '@/components/form/inputs/NumberInput'
import { BooleanInput } from '@/components/form/inputs/BooleanInput'
import { JsonInput } from '@/components/form/inputs/JsonInput'
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

export function BlogPostDetailPanel({ contentType, documentId }: Props) {
  const qc = useQueryClient()
  const { data: locales = [] } = useLocales()
  const [locale, setLocale] = useState('')
  const activeLocale = locale || locales[0] || ''

  const { data: doc, isLoading } = useDocument(documentId, activeLocale)
  const publish = usePublishDocument()
  const unpublish = useUnpublishDocument()

  if (isLoading) return <p className="text-muted-foreground">Loading…</p>
  if (!doc) return <p className="text-muted-foreground">Document not found.</p>

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
      query={{
        queryKey: ['documents', 'detail', doc.EntryID, activeLocale, 'data'],
        queryFn: () =>
          api
            .get<Document>(`/api/documents/${doc.EntryID}`, { params: { locale: activeLocale } })
            .then((r) => r.data.Data),
      }}
      mutationFn={mutationFn}
    >
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Link
              to="/admin/content-type/collection-type/blog-posts"
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              ← Blog Posts
            </Link>
            <h1 className="text-xl font-semibold">Edit Post</h1>
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
                type="button"
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
                type="button"
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
            <SaveButton />
          </div>
        </div>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Title</label>
            <FormField name="title">
              <TextInput placeholder="Post title" aria-label="Title" />
            </FormField>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Slug</label>
            <FormField name="slug">
              <TextInput placeholder="my-post-slug" aria-label="Slug" />
            </FormField>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Cover Image</label>
            <FormField name="coverImage">
              <MediaInput documentRef={doc.EntryID} contentTypeId={contentType.ID} />
            </FormField>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Excerpt</label>
            <FormField name="excerpt">
              <TextInput multiline placeholder="Short summary of the post…" aria-label="Excerpt" />
            </FormField>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Body</label>
            <FormField name="body">
              <RichTextInput />
            </FormField>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Reading Time (min)</label>
            <FormField name="readingTime">
              <NumberInput min={1} aria-label="Reading time in minutes" />
            </FormField>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Featured</label>
            <FormField name="featured">
              <BooleanInput aria-label="Featured post" />
            </FormField>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Metadata</label>
            <FormField name="metadata">
              <JsonInput />
            </FormField>
          </div>
        </div>
      </div>
    </FormProvider>
  )
}
