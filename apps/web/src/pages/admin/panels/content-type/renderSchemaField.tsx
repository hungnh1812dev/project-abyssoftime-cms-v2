import { lazy, Suspense } from "react";
import {
  FormField,
  TextInput,
  BooleanInput,
  NumberInput,
  MediaInput,
} from "@/components/form";
import type { FieldDefinition } from "@/types/cms";

const RichTextInput = lazy(() =>
  import("@/components/form/inputs/RichTextInput").then((m) => ({
    default: m.RichTextInput,
  })),
);

const JsonInput = lazy(() =>
  import("@/components/form/inputs/JsonInput").then((m) => ({
    default: m.JsonInput,
  })),
);

function primitiveInput(field: FieldDefinition): React.ReactElement<Record<string, unknown>> {
  switch (field.type) {
    case "number":
      return <NumberInput aria-label={field.name} />;
    case "boolean":
      return <BooleanInput aria-label={field.name} />;
    case "richtext":
      return <Suspense fallback={<div className="h-48 animate-pulse rounded-md bg-muted" />}><RichTextInput aria-label={field.name} /></Suspense>;
    case "json":
      return <Suspense fallback={<div className="h-48 animate-pulse rounded-md bg-muted" />}><JsonInput aria-label={field.name} /></Suspense>;
    case "media":
      return <MediaInput aria-label={field.name} />;
    default:
      return <TextInput aria-label={field.name} placeholder={field.name} />;
  }
}

function renderField(field: FieldDefinition, prefix = ""): React.ReactNode {
  const fieldName = prefix + field.name;

  if (field.type === "layout") {
    return (
      <div key={fieldName} className="grid md:grid-cols-2 gap-4">
        {(field.fields ?? []).map((child) => renderField(child, prefix))}
      </div>
    );
  }

  if (field.type === "component") {
    const childPrefix = fieldName + ".";
    return (
      <fieldset key={fieldName} className="border rounded-md p-4">
        <legend className="px-1 text-sm font-medium">{field.name}</legend>
        {(field.fields ?? []).map((child) => renderField(child, childPrefix))}
      </fieldset>
    );
  }

  return (
    <div key={fieldName}>
      <label className="block text-sm font-medium mb-1">{field.name}</label>
      <FormField name={fieldName}>{primitiveInput(field)}</FormField>
    </div>
  );
}

export function renderSchemaField(
  field: FieldDefinition,
  prefix = "",
): React.ReactNode {
  return renderField(field, prefix);
}
