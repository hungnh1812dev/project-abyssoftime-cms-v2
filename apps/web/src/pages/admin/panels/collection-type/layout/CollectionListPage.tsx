import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter, DialogClose } from '@/components/ui/dialog';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { useCollectionDocuments, useDeleteCollectionDocument, useDuplicateCollectionDocument } from '@/hooks/useCollectionDocuments';
import { useUpdateListFields } from '@/hooks/useContentTypes';
import { useLocales } from '@/hooks/useLocales';
import { getRegistration, type CollectionColumnDef } from '@/content-type-registry';
import { ColumnChooserDialog } from '@/components/collection/ColumnChooserDialog';
import { LocaleSelector } from '@/components/locale/LocaleSelector';
import { flattenFields, type ContentType, type Document, type FieldDefinition } from '@/types/cms';
import { Pencil, Trash2, Copy, ArrowUpDown, ArrowUp, ArrowDown, Settings2 } from 'lucide-react';
import { Breadcrumb } from '@/components/ui/breadcrumb';

interface Props {
  contentType: ContentType;
}

const SYSTEM_FIELD_KEYS = new Set(['createdAt', 'updatedAt', 'updatedByName']);

function deriveColumns(contentType: ContentType): CollectionColumnDef[] {
  const registration = getRegistration(contentType.Slug);
  if (registration?.columns) return registration.columns;

  const listFieldNames = (contentType.listFields ?? []).filter((name) => !SYSTEM_FIELD_KEYS.has(name));
  const fields = flattenFields(contentType.Fields ?? []);
  const fieldMap = new Map<string, FieldDefinition>();
  for (const field of fields) fieldMap.set(field.name, field);

  const names =
    listFieldNames.length > 0
      ? listFieldNames
      : fields
          .filter((field) => field.type !== 'component')
          .slice(0, 3)
          .map((field) => field.name);

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

function deriveSystemVisibility(listFields: string[]): { showCreatedAt: boolean; showUpdatedAt: boolean; showUpdatedBy: boolean } {
  if (listFields.length === 0) {
    return { showCreatedAt: true, showUpdatedAt: true, showUpdatedBy: true };
  }
  return {
    showCreatedAt: listFields.includes('createdAt'),
    showUpdatedAt: listFields.includes('updatedAt'),
    showUpdatedBy: listFields.includes('updatedByName'),
  };
}

function cellValue(doc: Document, col: CollectionColumnDef): React.ReactNode {
  const raw = doc.data[col.key];
  switch (col.type) {
    case 'boolean':
      return raw ? '✓' : '—';
    case 'number':
      return String(raw ?? '');
    case 'image':
      return <img src={String(raw ?? '')} alt={col.label} className="h-8 w-8 rounded object-cover" />;
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
    <button type="button" className="hover:text-foreground flex items-center gap-1 font-medium" onClick={() => onSort(field)}>
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
  const [columnChooserOpen, setColumnChooserOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const { data: locales = [] } = useLocales();
  const [selectedLocale, setSelectedLocale] = useState('');
  const activeLocale = selectedLocale || locales.find((loc) => loc.isDefault)?.code || locales[0]?.code || '';

  const queryClient = useQueryClient();
  const { data: page, isLoading } = useCollectionDocuments(contentType.Slug, start, PAGE_SIZE, activeLocale, orderBy, sortDir);
  const { mutate: deleteDoc } = useDeleteCollectionDocument();
  const { mutateAsync: duplicateDoc } = useDuplicateCollectionDocument();
  const updateListFields = useUpdateListFields();
  const navigate = useNavigate();

  const hasRegistryOverride = Boolean(getRegistration(contentType.Slug)?.columns);
  const columns = deriveColumns(contentType);
  const systemVis = deriveSystemVisibility(contentType.listFields ?? []);
  const docs = page?.items ?? [];
  const total = page?.total ?? 0;

  function handleCreate() {
    navigate(`/admin/content-type/collection-type/${contentType.Slug}/new`);
  }

  function handleDelete(event: React.MouseEvent, doc: Document) {
    event.stopPropagation();
    setDeleteTarget(doc.data.documentId as string);
  }

  function confirmDelete() {
    if (!deleteTarget) return;
    deleteDoc({ contentTypeSlug: contentType.Slug, id: deleteTarget });
    setDeleteTarget(null);
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

  function handleSaveListFields(selectedFields: string[]) {
    updateListFields.mutate(
      { slug: contentType.Slug, listFields: selectedFields },
      {
        onSuccess: () => {
          setColumnChooserOpen(false);
          queryClient.invalidateQueries({ queryKey: ['documents', 'collection-type', contentType.Slug] });
        },
      },
    );
  }

  if (isLoading) {
    return <p className="text-muted-foreground p-6">Loading…</p>;
  }

  const showingFrom = total === 0 ? 0 : start + 1;
  const showingTo = Math.min(start + PAGE_SIZE, total);

  return (
    <div className="space-y-4 p-6">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-0.5">
          <Breadcrumb
            items={[
              { label: 'Home', to: '/admin' },
              { label: 'Content Manager' },
            ]}
          />
          <h1 className="text-xl font-semibold">{contentType.Name}</h1>
        </div>
        <div className="flex items-center gap-2">
          <LocaleSelector
            value={activeLocale}
            onChange={(code) => {
              setSelectedLocale(code);
              setStart(0);
            }}
          />
          {!hasRegistryOverride && (
            <Button variant="outline" size="icon" aria-label="Configure columns" onClick={() => setColumnChooserOpen(true)}>
              <Settings2 className="h-4 w-4" />
            </Button>
          )}
          <Button onClick={handleCreate}>Add new item</Button>
        </div>
      </div>

      {!hasRegistryOverride && (
        <ColumnChooserDialog
          open={columnChooserOpen}
          onOpenChange={setColumnChooserOpen}
          contentType={contentType}
          currentListFields={contentType.listFields ?? []}
          onSave={handleSaveListFields}
          isSaving={updateListFields.isPending}
        />
      )}

      <Dialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}>
        <DialogContent showCloseButton={false}>
          <DialogHeader>
            <DialogTitle>Delete entry</DialogTitle>
            <DialogDescription>Are you sure you want to delete this entry? This action cannot be undone.</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose render={<Button variant="outline" />}>Cancel</DialogClose>
            <Button variant="destructive" onClick={confirmDelete}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {docs.length === 0 ? (
        <p className="text-muted-foreground">No entries yet.</p>
      ) : (
        <>
          <div className="overflow-x-auto rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-16">
                    <SortableHeader label="Id" field="id" activeField={orderBy} activeDir={sortDir} onSort={handleSort} />
                  </TableHead>
                  {columns.map((column) => (
                    <TableHead key={column.key}>{column.label}</TableHead>
                  ))}
                  {systemVis.showCreatedAt && (
                    <TableHead>
                      <SortableHeader label="Created At" field="createdAt" activeField={orderBy} activeDir={sortDir} onSort={handleSort} />
                    </TableHead>
                  )}
                  {systemVis.showUpdatedAt && (
                    <TableHead>
                      <SortableHeader label="Updated At" field="updatedAt" activeField={orderBy} activeDir={sortDir} onSort={handleSort} />
                    </TableHead>
                  )}
                  {systemVis.showUpdatedBy && <TableHead>Updated By</TableHead>}
                  <TableHead>Status</TableHead>
                  <TableHead className="w-24 text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {docs.map((doc) => (
                  <TableRow key={doc.data.documentId as string} className="cursor-pointer" onClick={() => handleRowClick(doc)}>
                    <TableCell className="font-mono text-sm">{String(doc.data.id ?? '')}</TableCell>
                    {columns.map((column) => (
                      <TableCell key={column.key}>{cellValue(doc, column)}</TableCell>
                    ))}
                    {systemVis.showCreatedAt && <TableCell className="text-muted-foreground text-sm">{formatDate(doc.data.createdAt)}</TableCell>}
                    {systemVis.showUpdatedAt && <TableCell className="text-muted-foreground text-sm">{formatDate(doc.data.updatedAt)}</TableCell>}
                    {systemVis.showUpdatedBy && <TableCell className="text-muted-foreground text-sm">{String(doc.data.updatedByName ?? '')}</TableCell>}
                    <TableCell>
                      <Badge variant={statusVariant[doc.status] ?? 'draft'}>{doc.status}</Badge>
                    </TableCell>
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
