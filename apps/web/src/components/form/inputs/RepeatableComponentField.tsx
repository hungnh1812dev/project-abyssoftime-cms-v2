import { useState } from 'react';
import { useFieldArray, useFormContext } from 'react-hook-form';
import { ArrowUp, ArrowDown, Trash2, Plus, ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import type { FieldDefinition } from '@/types/cms';

const depthStyles = [
  { border: 'border-indigo-300 dark:border-indigo-700', bg: 'bg-indigo-50/50 dark:bg-indigo-950/20', legend: 'text-indigo-700 dark:text-indigo-300' },
  { border: 'border-violet-300 dark:border-violet-700', bg: 'bg-violet-50/50 dark:bg-violet-950/20', legend: 'text-violet-700 dark:text-violet-300' },
  { border: 'border-amber-300 dark:border-amber-700', bg: 'bg-amber-50/50 dark:bg-amber-950/20', legend: 'text-amber-700 dark:text-amber-300' },
] as const;

function getDepthStyle(depth: number) {
  return depthStyles[depth % depthStyles.length];
}

function findFirstTextFieldName(fields: FieldDefinition[]): string | undefined {
  return fields.find((fld) => fld.type === 'text')?.name;
}

function formatHintText(value: unknown): string {
  if (typeof value !== 'string') return '';
  const trimmed = value.trim();
  if (!trimmed) return '';
  return trimmed.length > 60 ? trimmed.slice(0, 60) + '...' : trimmed;
}

interface RepeatableEntryProps {
  entryId: string;
  entryIndex: number;
  totalEntries: number;
  parentName: string;
  childFields: FieldDefinition[];
  onSwap: (indexA: number, indexB: number) => void;
  onRemove: (index: number) => void;
  renderField: (field: FieldDefinition, prefix: string, depth: number, index: number) => React.ReactNode;
  depth: number;
}

function RepeatableEntry({ entryId: _entryId, entryIndex, totalEntries, parentName, childFields, onSwap, onRemove, renderField, depth }: RepeatableEntryProps) {
  const [expanded, setExpanded] = useState(false);
  const { watch } = useFormContext();

  const firstTextName = findFirstTextFieldName(childFields);
  const rawValue = firstTextName ? watch(`${parentName}.${entryIndex}.${firstTextName}`) : undefined;
  const hintText = formatHintText(rawValue);

  return (
    <div className="bg-background relative rounded-md border p-4">
      <div className="mb-3 flex items-center justify-between">
        <button
          type="button"
          className="flex items-center gap-1"
          onClick={() => setExpanded((prev) => !prev)}
          aria-expanded={expanded}
        >
          <ChevronRight className={cn('size-3.5 shrink-0 transition-transform duration-200', expanded && 'rotate-90')} />
          <span className="text-muted-foreground text-xs font-medium">#{entryIndex + 1}</span>
          {hintText && <span className="text-muted-foreground ml-1 truncate text-xs font-normal">{'— '}{hintText}</span>}
        </button>
        <div className="flex items-center gap-1">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            disabled={entryIndex === 0}
            onClick={() => onSwap(entryIndex, entryIndex - 1)}
            aria-label={`Move item ${entryIndex + 1} up`}
          >
            <ArrowUp className="h-3.5 w-3.5" />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            disabled={entryIndex === totalEntries - 1}
            onClick={() => onSwap(entryIndex, entryIndex + 1)}
            aria-label={`Move item ${entryIndex + 1} down`}
          >
            <ArrowDown className="h-3.5 w-3.5" />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-destructive hover:text-destructive"
            onClick={() => onRemove(entryIndex)}
            aria-label={`Remove item ${entryIndex + 1}`}
          >
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>
      {expanded && (
        <div className="grid grid-cols-1 md:grid-cols-6 gap-4">
          {childFields.map((child, childIndex) => renderField(child, `${parentName}.${entryIndex}.`, depth + 1, childIndex))}
        </div>
      )}
    </div>
  );
}

interface RepeatableComponentFieldProps {
  name: string;
  label: string;
  depth?: number;
  keyPrefix: string;
  fields: FieldDefinition[];
  renderField: (field: FieldDefinition, prefix: string, depth: number, index: number) => React.ReactNode;
}

export function RepeatableComponentField({ name, label, depth = 0, keyPrefix: _keyPrefix, fields: childFields, renderField }: RepeatableComponentFieldProps) {
  const { control } = useFormContext();
  const { fields, append, remove, swap } = useFieldArray({ control, name });
  const style = getDepthStyle(depth);

  return (
    <fieldset className={`md:col-span-6 rounded-md border p-4 ${style.border} ${style.bg}`}>
      <legend className={`px-1 text-sm font-medium ${style.legend}`}>{label}</legend>
      <div className="space-y-4">
        {fields.map((item, index) => (
          <RepeatableEntry
            key={item.id}
            entryId={item.id}
            entryIndex={index}
            totalEntries={fields.length}
            parentName={name}
            childFields={childFields}
            onSwap={swap}
            onRemove={remove}
            renderField={renderField}
            depth={depth}
          />
        ))}
        <Button type="button" variant="outline" size="sm" onClick={() => append({})}>
          <Plus className="mr-1 h-3.5 w-3.5" />
          Add entry
        </Button>
      </div>
    </fieldset>
  );
}
