import type { UseQueryOptions } from '@tanstack/react-query';
import { FormProvider, useCmsFormState } from '@/components/form';
import type { FieldDefinition } from '@/types/cms';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { renderSchemaField } from './renderSchemaField';

interface ContentTypeBuilderProps {
  schema: FieldDefinition[];
  query?: UseQueryOptions;
  mutationFn: (data: Record<string, unknown>) => Promise<unknown>;
  renderActions?: (formState: { isDirty: boolean; submitting: boolean }) => React.ReactNode;
}

function FormActions({ renderActions }: { renderActions?: ContentTypeBuilderProps['renderActions'] }) {
  const { isDirty, submitting } = useCmsFormState();

  return (
    <div className="flex items-center gap-2">
      <Button type="submit" variant={isDirty ? 'default' : 'secondary'} disabled={!isDirty || submitting} loading={submitting} loadingText="Saving...">
        Save
      </Button>
      {renderActions?.({ isDirty, submitting })}
    </div>
  );
}

export function ContentTypeBuilder({ schema, query, mutationFn, renderActions }: ContentTypeBuilderProps) {
  return (
    <FormProvider query={query} mutationFn={mutationFn}>
      <div className="space-y-6">
        <Card>
          <CardContent>
            <div className="space-y-4">{schema.map((field) => renderSchemaField(field))}</div>
          </CardContent>
        </Card>
        <FormActions renderActions={renderActions} />
      </div>
    </FormProvider>
  );
}
