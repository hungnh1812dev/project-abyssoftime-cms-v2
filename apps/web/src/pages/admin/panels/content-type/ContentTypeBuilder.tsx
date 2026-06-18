import type { UseQueryOptions } from "@tanstack/react-query";
import { FormProvider } from "@/components/form";
import type { FieldDefinition } from "@/types/cms";
import { Button } from "@/components/ui/button";
import { renderSchemaField } from "./renderSchemaField";

interface ContentTypeBuilderProps {
  schema: FieldDefinition[];
  query?: UseQueryOptions;
  mutationFn: (data: Record<string, unknown>) => Promise<unknown>;
  children?: React.ReactNode;
}

export function ContentTypeBuilder({
  schema,
  query,
  mutationFn,
  children,
}: ContentTypeBuilderProps) {
  return (
    <FormProvider query={query} mutationFn={mutationFn}>
      <div className="space-y-4">
        {schema.map((field) => renderSchemaField(field))}
        {children ?? <Button type="submit">Save</Button>}
      </div>
    </FormProvider>
  );
}
