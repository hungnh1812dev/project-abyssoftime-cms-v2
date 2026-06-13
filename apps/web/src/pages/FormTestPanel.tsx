import { FormProvider } from '@/components/form/FormProvider'
import { FormField } from '@/components/form/FormField'
import { TextInput } from '@/components/form/inputs/TextInput'
import { NumberInput } from '@/components/form/inputs/NumberInput'
import { BooleanInput } from '@/components/form/inputs/BooleanInput'
import { RichTextInput } from '@/components/form/inputs/RichTextInput'
import { JsonInput } from '@/components/form/inputs/JsonInput'

export function FormTestPanel() {
  return (
    <div className="mx-auto max-w-2xl space-y-6 p-8">
      <h1 className="text-2xl font-semibold">Form Input Test Panel</h1>
      <FormProvider
        mutationFn={(data) => {
          console.log('Form submitted:', JSON.stringify(data, null, 2))
          return Promise.resolve()
        }}
      >
        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium">Title (text)</label>
            <FormField name="article.title">
              <TextInput aria-label="article.title" placeholder="Enter title…" />
            </FormField>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium">Body (rich text)</label>
            <FormField name="article.body">
              <RichTextInput />
            </FormField>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium">View count (number)</label>
            <FormField name="article.stats.views">
              <NumberInput aria-label="article.stats.views" min={0} />
            </FormField>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium">Published (boolean)</label>
            <FormField name="article.published">
              <BooleanInput aria-label="article.published" />
            </FormField>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium">Metadata (JSON)</label>
            <FormField name="article.metadata">
              <JsonInput />
            </FormField>
          </div>
        </div>

        <button
          type="submit"
          className="mt-6 rounded-md bg-primary px-4 py-2 text-primary-foreground"
        >
          Submit (check console)
        </button>
      </FormProvider>
    </div>
  )
}
