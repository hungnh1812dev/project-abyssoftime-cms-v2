/* eslint-disable react-refresh/only-export-components */
import { lazy, Suspense } from 'react';
import { FormField, TextInput, BooleanInput, NumberInput, MediaInput } from '@/components/form';
import { RepeatableComponentField } from '@/components/form/inputs/RepeatableComponentField';
import type { FieldDefinition } from '@/types/cms';

const RichTextInput = lazy(() =>
  import('@/components/form/inputs/RichTextInput').then((mod) => ({
    default: mod.RichTextInput,
  })),
);

const JsonInput = lazy(() =>
  import('@/components/form/inputs/JsonInput').then((mod) => ({
    default: mod.JsonInput,
  })),
);

function primitiveInput(field: FieldDefinition): React.ReactElement<Record<string, unknown>> {
  switch (field.type) {
    case 'number':
      return <NumberInput aria-label={field.name} />;
    case 'boolean':
      return <BooleanInput aria-label={field.name} />;
    default:
      return <TextInput aria-label={field.name} placeholder={field.name} />;
  }
}

function widthToColSpan(width: string | undefined): string {
  switch (width) {
    case '50%':
      return 'md:col-span-3';
    case '1/3':
      return 'md:col-span-2';
    default:
      return 'md:col-span-6';
  }
}

const depthStyles = [
  { border: 'border-indigo-300 dark:border-indigo-700', bg: 'bg-indigo-50/50 dark:bg-indigo-950/20', legend: 'text-indigo-700 dark:text-indigo-300' },
  { border: 'border-violet-300 dark:border-violet-700', bg: 'bg-violet-50/50 dark:bg-violet-950/20', legend: 'text-violet-700 dark:text-violet-300' },
  { border: 'border-amber-300 dark:border-amber-700', bg: 'bg-amber-50/50 dark:bg-amber-950/20', legend: 'text-amber-700 dark:text-amber-300' },
] as const;

function renderField(field: FieldDefinition, prefix: string, keyPrefix: string, depth: number, _index: number): React.ReactNode {
  const fieldName = prefix + field.name;
  const fieldKey = `${keyPrefix}${field.name}`;

  if (field.type === 'component') {
    const childKeyPrefix = `${fieldKey}_`;
    if (field.repeatable) {
      return (
        <RepeatableComponentField
          key={fieldKey}
          name={fieldName}
          label={field.name}
          depth={depth}
          keyPrefix={childKeyPrefix}
          fields={field.fields ?? []}
          renderField={(child: FieldDefinition, childPrefix: string, childDepth: number, childIndex: number) => renderField(child, childPrefix, childKeyPrefix, childDepth, childIndex)}
        />
      );
    }
    const childPrefix = fieldName + '.';
    const style = depthStyles[depth % depthStyles.length];
    return (
      <fieldset key={fieldKey} className={`md:col-span-6 rounded-md border p-4 ${style.border} ${style.bg}`}>
        <legend className={`px-1 text-sm font-medium ${style.legend}`}>{field.name}</legend>
        <div className="grid grid-cols-1 md:grid-cols-6 gap-4">
          {(field.fields ?? []).map((child, childIndex) => renderField(child, childPrefix, childKeyPrefix, depth + 1, childIndex))}
        </div>
      </fieldset>
    );
  }

  const colSpan = widthToColSpan(field.width);

  if (field.type === 'json') {
    return (
      <div key={fieldKey} className={colSpan}>
        <label className="mb-1 block text-sm font-medium">{field.name}</label>
        <Suspense fallback={<div className="bg-muted h-48 animate-pulse rounded-md" />}>
          <JsonInput name={fieldName} />
        </Suspense>
      </div>
    );
  }

  if (field.type === 'richtext') {
    return (
      <div key={fieldKey} className={colSpan}>
        <label className="mb-1 block text-sm font-medium">{field.name}</label>
        <Suspense fallback={<div className="bg-muted h-48 animate-pulse rounded-md" />}>
          <RichTextInput name={fieldName} />
        </Suspense>
      </div>
    );
  }

  if (field.type === 'media') {
    return (
      <div key={fieldKey} className={colSpan}>
        <label className="mb-1 block text-sm font-medium">{field.name}</label>
        <MediaInput name={fieldName} ext={field.ext} />
      </div>
    );
  }

  return (
    <div key={fieldKey} className={colSpan}>
      <label className="mb-1 block text-sm font-medium">{field.name}</label>
      <FormField name={fieldName}>{primitiveInput(field)}</FormField>
    </div>
  );
}

export function renderSchemaField(field: FieldDefinition, prefix: string, keyPrefix: string, index: number): React.ReactNode {
  return renderField(field, prefix, keyPrefix, 0, index);
}
