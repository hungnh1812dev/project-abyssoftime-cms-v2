import { Link, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { useDocuments, useDeleteDocument, useCreateDocument } from '@/hooks/useDocuments'
import type { ContentType } from '@/types/cms'

interface Props {
  contentType: ContentType
}

export function CollectionListPage({ contentType }: Props) {
  const { data: docs = [], isLoading } = useDocuments(contentType.ID)
  const { mutate: deleteDoc } = useDeleteDocument()
  const { mutateAsync: createDoc } = useCreateDocument()
  const navigate = useNavigate()

  async function handleCreate() {
    const newDoc = await createDoc({ contentTypeId: contentType.ID, data: {} })
    navigate(`/admin/content-types/${contentType.Slug}/${newDoc.EntryID}`)
  }

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{contentType.Name}</h1>
        <Button onClick={handleCreate}>Add entry</Button>
      </div>

      {docs.length === 0 ? (
        <p className="text-muted-foreground">No entries yet.</p>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b text-left">
              <th className="py-2 pr-4 font-medium">Entry</th>
              <th className="py-2 pr-4 font-medium">Status</th>
              <th className="py-2 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {docs.map((doc) => {
              const preview = String(Object.values(doc.Data)[0] ?? doc.EntryID)
              return (
                <tr key={doc.EntryID} className="border-b">
                  <td className="py-2 pr-4">{preview}</td>
                  <td className="py-2 pr-4 capitalize">{doc.Status}</td>
                  <td className="py-2 flex gap-2">
                    <Link
                      to={`/admin/content-types/${contentType.Slug}/${doc.EntryID}`}
                      className="text-primary underline-offset-4 hover:underline"
                    >
                      Edit
                    </Link>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() =>
                        deleteDoc({ id: doc.EntryID, contentTypeId: contentType.ID })
                      }
                    >
                      Delete
                    </Button>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      )}
    </div>
  )
}
