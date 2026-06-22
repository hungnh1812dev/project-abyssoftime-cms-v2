import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table';
import { useCollectionDocuments, useDeleteCollectionDocument, useDuplicateCollectionDocument } from '@/hooks/useCollectionDocuments';
import { useLocales } from '@/hooks/useLocales';
import { getRegistration, type CollectionColumnDef } from '@/content-type-registry';
import type { ContentType, Document, FieldDefinition } from '@/types/cms';
import { Pencil, Trash2, Copy, ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react';

interface Props {
  contentType: ContentType;
}

function deriveColumns(contentType: ContentType): CollectionColumnDef[] {
  const registration = getRegistration(contentType.Slug);
  if (registration?.columns) return registration.columns;

  const listFieldNames = contentType.listFields ?? [];
  const fields = contentType.Fields ?? [];
  const fieldMap = new Map<string, FieldDefinition>();
  for (const field of fields) fieldMap.set(field.name, field);

  const names = listFieldNames.length > 0 ? listFieldNames : fields.slice(0, 3).map((field) => field.name);

  return names.map((name) => {
    const field = fieldMap.get(name);
    const fieldType = field?.type ?? 'text';
    let colType: CollectionColumnDef['type'] = 'text';
    if (fieldType === 'boolean') colType = 'boolean';
    else if (fieldType === 'number') colType = 'number';
    else if (fieldType === 'media') colType = 'image';
    return { key: name, label: name, type: colType };
  });
}

function cellValue(doc: Document, col: CollectionColumnDef): React.ReactNode {
  const raw = doc.data[col.key];
  switch (col.type) {
    case 'boolean':
      return raw ? '✓' : '—';
    case 'number':
      return String(raw ?? '');
    case 'image':
      return <img src={String(raw ?? '')} alt={col.label} className="h-8 w-8 object-cover rounded" />;
    default:
      return String(raw ?? '');
  }
}

function formatDate(value: unknown): string {
  if (!value) return '—';
  const date = new Date(String(value));
  if (isNaN(date.getTime())) return '—';
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  }).format(date);
}

type SortField = 'id' | 'createdAt' | 'updatedAt';
type SortDir = 'asc' | 'desc';

function SortableHeader({
  label,
  field,
  activeField,
  activeDir,
  onSort,
}: {
  label: string;
  field: SortField;
  activeField: string;
  activeDir: SortDir;
  onSort: (field: SortField) => void;
}) {
  const isActive = activeField === field;
  const Icon = isActive ? (activeDir === 'asc' ? ArrowUp : ArrowDown) : ArrowUpDown;
  return (
    <button
      type="button"
      className="flex items-center gap-1 font-medium hover:text-foreground"
      onClick={() => onSort(field)}
    >
      {label}
      <Icon className={`h-3.5 w-3.5 ${isActive ? 'text-foreground' : 'text-muted-foreground'}`} />
    </button>
  );
}

const statusVariant: Record<string, 'draft' | 'published' | 'modified'> = {
  draft: 'draft',
  published: 'published',
  modified: 'modified',
};

const PAGE_SIZE = 20;

export function CollectionListPage({ contentType }: Props) {
  const [start, setStart] = useState(0);
  const [orderBy, setOrderBy] = useState<SortField>('id');
  const [sortDir, setSortDir] = useState<SortDir>('desc');
  const { data: locales = [] } = useLocales();
  const activeLocale = locales[0] || '';

  const { data: page, isLoading } = useCollectionDocuments(contentType.Slug, start, PAGE_SIZE, activeLocale, orderBy, sortDir);
  const { mutate: deleteDoc } = useDeleteCollectionDocument();
  const { mutateAsync: duplicateDoc } = useDuplicateCollectionDocument();
  const navigate = useNavigate();

  const columns = deriveColumns(contentType);
  const docs = page?.items ?? [];
  const total = page?.total ?? 0;

  function handleCreate() {
    navigate(`/admin/content-type/collection-type/${contentType.Slug}/new`);
  }

  function handleDelete(event: React.MouseEvent, doc: Document) {
    event.stopPropagation();
    if (!window.confirm('Delete this entry?')) return;
    deleteDoc({ contentTypeSlug: contentType.Slug, id: doc.data.documentId as string });
  }

  function handleDuplicate(event: React.MouseEvent, doc: Document) {
    event.stopPropagation();
    duplicateDoc({
      contentTypeSlug: contentType.Slug,
      id: doc.data.documentId as string,
      locale: activeLocale,
    }).then((newDoc) => {
      navigate(`/admin/content-type/collection-type/${contentType.Slug}/${newDoc.data.documentId}`);
    });
  }

  function handleEdit(event: React.MouseEvent, doc: Document) {
    event.stopPropagation();
    navigate(`/admin/content-type/collection-type/${contentType.Slug}/${doc.data.documentId}`);
  }

  function handleSort(sortField: SortField) {
    if (orderBy === sortField) {
      setSortDir((currentDir) => (currentDir === 'desc' ? 'asc' : 'desc'));
    } else {
      setOrderBy(sortField);
      setSortDir('desc');
    }
    setStart(0);
  }

  function handleRowClick(doc: Document) {
    navigate(`/admin/content-type/collection-type/${contentType.Slug}/${doc.data.documentId}`);
  }

  if (isLoading) {
    return <p className="text-muted-foreground p-6">Loading…</p>;
  }

  const showingFrom = total === 0 ? 0 : start + 1;
  const showingTo = Math.min(start + PAGE_SIZE, total);

  return (
    <div className="space-y-4 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{contentType.Name}</h1>
        <Button onClick={handleCreate}>Add new item</Button>
      </div>

      {docs.length === 0 ? (
        <p className="text-muted-foreground">No entries yet.</p>
      ) : (
        <>
          <div className="rounded-md border overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-16">
                    <SortableHeader label="Id" field="id" activeField={orderBy} activeDir={sortDir} onSort={handleSort} />
                  </TableHead>
                  {columns.map((column) => (
                    <TableHead key={column.key}>{column.label}</TableHead>
                  ))}
                  <TableHead>Status</TableHead>
                  <TableHead>
                    <SortableHeader label="Created At" field="createdAt" activeField={orderBy} activeDir={sortDir} onSort={handleSort} />
                  </TableHead>
                  <TableHead>
                    <SortableHeader label="Updated At" field="updatedAt" activeField={orderBy} activeDir={sortDir} onSort={handleSort} />
                  </TableHead>
                  <TableHead>Updated By</TableHead>
                  <TableHead className="w-24 text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {docs.map((doc) => (
                  <TableRow
                    key={doc.data.documentId as string}
                    className="cursor-pointer"
                    onClick={() => handleRowClick(doc)}
                  >
                    <TableCell className="font-mono text-sm">{String(doc.data.id ?? '')}</TableCell>
                    {columns.map((column) => (
                      <TableCell key={column.key}>{cellValue(doc, column)}</TableCell>
                    ))}
                    <TableCell>
                      <Badge variant={statusVariant[doc.status] ?? 'draft'}>{doc.status}</Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">{formatDate(doc.data.createdAt)}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{formatDate(doc.data.updatedAt)}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{String(doc.data.updatedByName ?? '')}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button variant="outline" size="icon-xs" className="hover:bg-accent-foreground/10" aria-label="Edit" onClick={(event) => handleEdit(event, doc)}>
                          <Pencil className="h-3 w-3" />
                        </Button>
                        <Button variant="outline" size="icon-xs" className="hover:bg-accent-foreground/10" aria-label="Duplicate" onClick={(event) => handleDuplicate(event, doc)}>
                          <Copy className="h-3 w-3" />
                        </Button>
                        <Button variant="destructive" size="icon-xs" aria-label="Delete" onClick={(event) => handleDelete(event, doc)}>
                          <Trash2 className="h-3 w-3" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          <div className="text-muted-foreground flex items-center justify-between text-sm">
            <span>
              Showing {showingFrom}–{showingTo} of {total}
            </span>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" disabled={start === 0} onClick={() => setStart(Math.max(0, start - PAGE_SIZE))}>
                Previous
              </Button>
              <Button variant="outline" size="sm" disabled={start + PAGE_SIZE >= total} onClick={() => setStart(start + PAGE_SIZE)}>
                Next
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
