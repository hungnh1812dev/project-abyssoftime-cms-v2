import { Link, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { useDocuments, useDeleteDocument, useCreateDocument } from '@/hooks/useDocuments'
import { getRegistration, type CollectionColumnDef } from '@/content-type-registry'
import type { ContentType, Document } from '@/types/cms'

interface Props {
  contentType: ContentType
}

function cellValue(doc: Document, col: CollectionColumnDef): React.ReactNode {
  const raw = doc.Data[col.key]
  switch (col.type) {
    case 'boolean':
      return raw ? '✓' : '—'
    case 'number':
      return String(raw ?? '')
    case 'image':
      return <img src={String(raw ?? '')} alt={col.label} className="h-8 w-8 object-cover" />
    default:
      return String(raw ?? '')
  }
}

export function CollectionListPage({ contentType }: Props) {
  const { data: docs = [], isLoading } = useDocuments(contentType.ID)
  const { mutate: deleteDoc } = useDeleteDocument()
  const { mutateAsync: createDoc } = useCreateDocument()
  const navigate = useNavigate()

  const registration = getRegistration(contentType.Slug)
  const columns = registration?.columns

  async function handleCreate() {
    const newDoc = await createDoc({ contentTypeId: contentType.ID, data: {} })
    navigate(`/admin/content-type/collection-type/${contentType.Slug}/${newDoc.EntryID}`)
  }

  function handleDelete(doc: Document) {
    if (!window.confirm('Delete this entry?')) return
    deleteDoc({ id: doc.EntryID, contentTypeId: contentType.ID })
  }

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{contentType.Name}</h1>
        <Button onClick={handleCreate}>Add new item</Button>
      </div>

      {docs.length === 0 ? (
        <p className="text-muted-foreground">No entries yet.</p>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b text-left">
              {columns ? (
                columns.map((col) => (
                  <th key={col.key} className="py-2 pr-4 font-medium">
                    {col.label}
                  </th>
                ))
              ) : (
                <>
                  <th className="py-2 pr-4 font-medium">Entry</th>
                  <th className="py-2 pr-4 font-medium">Status</th>
                </>
              )}
              <th className="py-2 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {docs.map((doc) => (
              <tr key={doc.EntryID} className="border-b">
                {columns ? (
                  columns.map((col) => (
                    <td key={col.key} className="py-2 pr-4">
                      {cellValue(doc, col)}
                    </td>
                  ))
                ) : (
                  <>
                    <td className="py-2 pr-4">
                      {String(Object.values(doc.Data)[0] ?? doc.EntryID)}
                    </td>
                    <td className="py-2 pr-4 capitalize">{doc.Status}</td>
                  </>
                )}
                <td className="py-2 flex gap-2">
                  <Link
                    to={`/admin/content-type/collection-type/${contentType.Slug}/${doc.EntryID}`}
                    className="text-primary underline-offset-4 hover:underline"
                  >
                    Edit
                  </Link>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() => handleDelete(doc)}
                  >
                    Delete
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
