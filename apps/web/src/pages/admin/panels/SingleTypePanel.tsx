import { api } from '@/lib/api'
import { FormProvider } from '@/components/form/FormProvider'
import { FormField } from '@/components/form/FormField'
import { TextInput } from '@/components/form/inputs/TextInput'
import { Button } from '@/components/ui/button'
import {
  useDocuments,
  useUpdateDocument,
  usePublishDocument,
  useUnpublishDocument,
} from '@/hooks/useDocuments'
import type { ContentType, Document } from '@/types/cms'

interface Props {
  contentType: ContentType
}

export function SingleTypePanel({ contentType }: Props) {
  const { data: docs, isLoading } = useDocuments(contentType.ID)
  const doc = docs?.[0]

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
    updateDoc({ id: doc.EntryID, contentTypeId: contentType.ID, data })

  const fieldKeys = Object.keys(doc.Data)
  const canPublish = doc.Status !== 'published'
  const canUnpublish = doc.Status !== 'draft'

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">{contentType.Name}</h1>
          <span className="text-sm text-muted-foreground capitalize">{doc.Status}</span>
        </div>
        <div className="flex gap-2">
          {canPublish && (
            <Button
              onClick={() => publish.mutate({ id: doc.EntryID, contentTypeId: contentType.ID })}
              disabled={publish.isPending}
            >
              Publish
            </Button>
          )}
          {canUnpublish && (
            <Button
              variant="outline"
              onClick={() => unpublish.mutate({ id: doc.EntryID, contentTypeId: contentType.ID })}
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
            api.get<Document>(`/api/documents/${doc.EntryID}`).then((r) => r.data.Data),
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
