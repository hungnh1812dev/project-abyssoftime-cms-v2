import {
  FormField,
  TextInput,
  BooleanInput,
  NumberInput,
  JsonInput,
  MediaInput,
  RichTextInput,
} from "@/components/form";
import type { FieldDefinition } from "@/types/cms";

function primitiveInput(field: FieldDefinition): React.ReactElement<Record<string, unknown>> {
  switch (field.type) {
    case "number":
      return <NumberInput aria-label={field.name} />;
    case "boolean":
      return <BooleanInput aria-label={field.name} />;
    case "richtext":
      return <RichTextInput aria-label={field.name} />;
    case "json":
      return <JsonInput aria-label={field.name} />;
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
