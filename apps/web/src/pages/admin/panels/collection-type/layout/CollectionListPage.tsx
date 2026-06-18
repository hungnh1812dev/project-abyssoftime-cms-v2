import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { useCollectionDocuments, useDeleteCollectionDocument, useCreateCollectionDocument } from '@/hooks/useCollectionDocuments'
import { useLocales } from '@/hooks/useLocales'
import { getRegistration, type CollectionColumnDef } from '@/content-type-registry'
import type { ContentType, Document, FieldDefinition } from '@/types/cms'

interface Props {
  contentType: ContentType
}

function deriveColumns(contentType: ContentType): CollectionColumnDef[] {
  const registration = getRegistration(contentType.Slug)
  if (registration?.columns) return registration.columns

  const listFieldNames = contentType.listFields ?? []
  const fields = contentType.Fields ?? []
  const fieldMap = new Map<string, FieldDefinition>()
  for (const f of fields) fieldMap.set(f.name, f)

  const names = listFieldNames.length > 0
    ? listFieldNames
    : fields.slice(0, 3).map((f) => f.name)

  return names.map((name) => {
    const f = fieldMap.get(name)
    const fieldType = f?.type ?? 'text'
    let colType: CollectionColumnDef['type'] = 'text'
    if (fieldType === 'boolean') colType = 'boolean'
    else if (fieldType === 'number') colType = 'number'
    else if (fieldType === 'media') colType = 'image'
    return { key: name, label: name, type: colType }
  })
}

function cellValue(doc: Document, col: CollectionColumnDef): React.ReactNode {
  const raw = doc.data[col.key]
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

const PAGE_SIZE = 20

export function CollectionListPage({ contentType }: Props) {
  const [start, setStart] = useState(0)
  const { data: locales = [] } = useLocales()
  const activeLocale = locales[0] || ''

  const { data: page, isLoading } = useCollectionDocuments(contentType.Slug, start, PAGE_SIZE, activeLocale)
  const { mutate: deleteDoc } = useDeleteCollectionDocument()
  const { mutateAsync: createDoc } = useCreateCollectionDocument()
  const navigate = useNavigate()

  const columns = deriveColumns(contentType)
  const docs = page?.items ?? []
  const total = page?.total ?? 0

  async function handleCreate() {
    const newDoc = await createDoc({ contentTypeSlug: contentType.Slug, data: {} })
    navigate(`/admin/content-type/collection-type/${contentType.Slug}/${newDoc.documentId}`)
  }

  function handleDelete(doc: Document) {
    if (!window.confirm('Delete this entry?')) return
    deleteDoc({ contentTypeSlug: contentType.Slug, id: doc.documentId })
  }

  if (isLoading) {
    return <p className="text-muted-foreground">Loading…</p>
  }

  const showingFrom = total === 0 ? 0 : start + 1
  const showingTo = Math.min(start + PAGE_SIZE, total)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{contentType.Name}</h1>
        <Button onClick={handleCreate}>Add new item</Button>
      </div>

      {docs.length === 0 ? (
        <p className="text-muted-foreground">No entries yet.</p>
      ) : (
        <>
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left">
                {columns.map((col) => (
                  <th key={col.key} className="py-2 pr-4 font-medium">
                    {col.label}
                  </th>
                ))}
                <th className="py-2 pr-4 font-medium">Status</th>
                <th className="py-2 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {docs.map((doc) => (
                <tr key={doc.documentId} className="border-b">
                  {columns.map((col) => (
                    <td key={col.key} className="py-2 pr-4">
                      {cellValue(doc, col)}
                    </td>
                  ))}
                  <td className="py-2 pr-4 capitalize">{doc.status}</td>
                  <td className="py-2 flex gap-2">
                    <Link
                      to={`/admin/content-type/collection-type/${contentType.Slug}/${doc.documentId}`}
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

          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <span>Showing {showingFrom}–{showingTo} of {total}</span>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={start === 0}
                onClick={() => setStart(Math.max(0, start - PAGE_SIZE))}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={start + PAGE_SIZE >= total}
                onClick={() => setStart(start + PAGE_SIZE)}
              >
                Next
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
