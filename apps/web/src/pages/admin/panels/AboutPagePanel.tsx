import { useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import {
  useDocuments,
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
}

function SaveButton() {
  const { isDirty, submitting } = useCmsFormState()
  return (
    <Button type="submit" disabled={!isDirty || submitting}>
      Save
    </Button>
  )
}

export function AboutPagePanel({ contentType }: Props) {
  const qc = useQueryClient()
  const { data: docs = [], isLoading } = useDocuments(contentType.ID)
  const publish = usePublishDocument()
  const unpublish = useUnpublishDocument()

  if (isLoading) return <p className="text-muted-foreground">Loading…</p>

  const doc = docs[0]
  if (!doc) return <p className="text-muted-foreground">No document found.</p>

  const canPublish = doc.Status !== 'published'
  const canUnpublish = doc.Status !== 'draft'

  const mutationFn = async (data: Record<string, unknown>) => {
    const result = await api
      .put<Document>(`/api/documents/${doc.EntryID}`, { contentTypeId: contentType.ID, data })
      .then((r) => r.data)
    await qc.invalidateQueries({ queryKey: ['documents', contentType.ID] })
    return result
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">About Page</h1>
          <span className="text-sm text-muted-foreground capitalize">{doc.Status}</span>
        </div>
        <div className="flex gap-2">
          {canPublish && (
            <Button
              onClick={() =>
                publish.mutate({ id: doc.EntryID, contentTypeId: contentType.ID })
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
                unpublish.mutate({ id: doc.EntryID, contentTypeId: contentType.ID })
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
          queryKey: ['documents', 'detail', doc.EntryID, 'data'],
          queryFn: () =>
            api
              .get<Document>(`/api/documents/${doc.EntryID}`)
              .then((r) => r.data.Data),
        }}
        mutationFn={mutationFn}
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Headline</label>
            <FormField name="headline">
              <TextInput placeholder="Your headline" aria-label="Headline" />
            </FormField>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Introduction</label>
            <FormField name="introduction">
              <RichTextInput />
            </FormField>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Profile Photo</label>
            <FormField name="profilePhoto">
              <MediaInput documentRef={doc.EntryID} contentTypeId={contentType.ID} />
            </FormField>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Years of Experience</label>
            <FormField name="yearsOfExp">
              <NumberInput min={0} aria-label="Years of experience" />
            </FormField>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Open to Work</label>
            <FormField name="openToWork">
              <BooleanInput aria-label="Open to work" />
            </FormField>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Social Links</label>
            <FormField name="socialLinks">
              <JsonInput />
            </FormField>
          </div>

          <SaveButton />
        </div>
      </FormProvider>
    </div>
  )
}
