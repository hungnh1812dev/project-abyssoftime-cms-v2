import { useFieldArray, useFormContext } from 'react-hook-form';
import { ArrowUp, ArrowDown, Trash2, Plus } from 'lucide-react';
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

interface RepeatableComponentFieldProps {
  name: string;
  label: string;
  depth?: number;
  keyPrefix: string;
  fields: FieldDefinition[];
  renderField: (field: FieldDefinition, prefix: string, depth: number, index: number) => React.ReactNode;
}

export function RepeatableComponentField({ name, label, depth = 0, keyPrefix, fields: childFields, renderField }: RepeatableComponentFieldProps) {
  const { control } = useFormContext();
  const { fields, append, remove, swap } = useFieldArray({ control, name });
  const style = getDepthStyle(depth);

  return (
    <fieldset className={`rounded-md border p-4 ${style.border} ${style.bg}`}>
      <legend className={`px-1 text-sm font-medium ${style.legend}`}>{label}</legend>
      <div className="space-y-4">
        {fields.map((item, index) => (
          <div key={item.id} className="bg-background relative rounded-md border p-4">
            <div className="mb-3 flex items-center justify-between">
              <span className="text-muted-foreground text-xs font-medium">#{index + 1}</span>
              <div className="flex items-center gap-1">
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7"
                  disabled={index === 0}
                  onClick={() => swap(index, index - 1)}
                  aria-label={`Move item ${index + 1} up`}
                >
                  <ArrowUp className="h-3.5 w-3.5" />
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7"
                  disabled={index === fields.length - 1}
                  onClick={() => swap(index, index + 1)}
                  aria-label={`Move item ${index + 1} down`}
                >
                  <ArrowDown className="h-3.5 w-3.5" />
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 text-destructive hover:text-destructive"
                  onClick={() => remove(index)}
                  aria-label={`Remove item ${index + 1}`}
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
            <div className="space-y-4">
              {childFields.map((child, childIndex) => renderField(child, `${name}.${index}.`, depth + 1, childIndex))}
            </div>
          </div>
        ))}
        <Button type="button" variant="outline" size="sm" onClick={() => append({})}>
          <Plus className="mr-1 h-3.5 w-3.5" />
          Add entry
        </Button>
      </div>
    </fieldset>
  );
}
