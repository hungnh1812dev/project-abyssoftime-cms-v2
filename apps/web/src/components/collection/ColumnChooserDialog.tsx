import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { flattenFields, type ContentType } from '@/types/cms';

const SYSTEM_DISPLAY_FIELDS = [
  { key: 'createdAt', label: 'Created At' },
  { key: 'updatedAt', label: 'Updated At' },
  { key: 'updatedByName', label: 'Updated By' },
] as const;

interface ColumnChooserDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  contentType: ContentType;
  currentListFields: string[];
  onSave: (selectedFields: string[]) => void;
  isSaving: boolean;
}

function defaultSelection(contentType: ContentType): Set<string> {
  const fields = flattenFields(contentType.Fields ?? []).filter((field) => field.type !== 'component');
  const contentDefaults = fields.slice(0, 3).map((field) => field.name);
  const systemDefaults = SYSTEM_DISPLAY_FIELDS.map((field) => field.key);
  return new Set([...contentDefaults, ...systemDefaults]);
}

function initialSelection(contentType: ContentType, currentListFields: string[]): Set<string> {
  return currentListFields.length > 0
    ? new Set(currentListFields)
    : defaultSelection(contentType);
}

export function ColumnChooserDialog({
  open,
  onOpenChange,
  contentType,
  currentListFields,
  onSave,
  isSaving,
}: ColumnChooserDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {open && (
        <ColumnChooserContent
          contentType={contentType}
          currentListFields={currentListFields}
          onOpenChange={onOpenChange}
          onSave={onSave}
          isSaving={isSaving}
        />
      )}
    </Dialog>
  );
}

function ColumnChooserContent({
  contentType,
  currentListFields,
  onOpenChange,
  onSave,
  isSaving,
}: Omit<ColumnChooserDialogProps, 'open'>) {
  const [selected, setSelected] = useState<Set<string>>(
    () => initialSelection(contentType, currentListFields),
  );

  const contentFields = flattenFields(contentType.Fields ?? []).filter((field) => field.type !== 'component');

  function handleToggle(key: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  }

  function handleSave() {
    const contentKeys = contentFields.filter((field) => selected.has(field.name)).map((field) => field.name);
    const systemKeys = SYSTEM_DISPLAY_FIELDS.filter((field) => selected.has(field.key)).map((field) => field.key);
    onSave([...contentKeys, ...systemKeys]);
  }

  return (
    <DialogContent className="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>Configure columns</DialogTitle>
        <DialogDescription>Choose which columns to display in the list view.</DialogDescription>
      </DialogHeader>

      <div className="space-y-4 max-h-80 overflow-y-auto">
        <div>
          <h4 className="text-sm font-medium mb-2">Content fields</h4>
          <div className="space-y-2">
            {contentFields.map((field) => (
              <label key={field.name} className="flex items-center gap-2 text-sm cursor-pointer">
                <input
                  type="checkbox"
                  checked={selected.has(field.name)}
                  onChange={() => handleToggle(field.name)}
                  className="rounded border-input"
                />
                {field.name}
              </label>
            ))}
          </div>
        </div>

        <div>
          <h4 className="text-sm font-medium mb-2">System fields</h4>
          <div className="space-y-2">
            {SYSTEM_DISPLAY_FIELDS.map((field) => (
              <label key={field.key} className="flex items-center gap-2 text-sm cursor-pointer">
                <input
                  type="checkbox"
                  checked={selected.has(field.key)}
                  onChange={() => handleToggle(field.key)}
                  className="rounded border-input"
                />
                {field.label}
              </label>
            ))}
          </div>
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isSaving}>
          Cancel
        </Button>
        <Button onClick={handleSave} disabled={isSaving}>
          {isSaving ? 'Saving...' : 'Save'}
        </Button>
      </DialogFooter>
    </DialogContent>
  );
}
