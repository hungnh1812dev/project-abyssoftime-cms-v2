# Panel Creation Guide

A panel is a hard-coded React page that edits a specific content type. The generic panel (`ContentTypePanelPage`) handles any content type automatically, but you can override it with a custom panel that controls exactly which fields appear and how.

This guide walks through creating a custom **"Site Settings"** single-type panel as a concrete example.

---

## Overview of the 6 steps

1. Create the content type via the CMS UI
2. Create the panel component file
3. Fetch the document with a query hook
4. Wire mutations for save / publish
5. Build the form with `FormProvider` + `FormField` + inputs
6. Register the panel as a route in `router.tsx`

---

## Step 1 — Create the content type

Navigate to **Admin → Content Types** in the running CMS and create:

| Field | Value |
|-------|-------|
| Name | `Site Settings` |
| Slug | `site-settings` |
| Kind | `single` |

Alternatively, POST directly to the API:

```sh
curl -X POST http://localhost:8080/api/content-types \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Site Settings","slug":"site-settings","kind":"single"}'
```

The sidebar will show **Site Settings** immediately (it reads from `useContentTypes()`). Clicking it loads the generic panel. Steps 2–6 replace that with a custom one.

---

## Step 2 — Create the panel file

```
apps/web/src/pages/admin/panels/SiteSettingsPanel.tsx
```

Start with the skeleton:

```tsx
import type { ContentType } from '@/types/cms'

interface Props {
  contentType: ContentType
}

export function SiteSettingsPanel({ contentType }: Props) {
  return <div>Loading…</div>
}
```

---

## Step 3 — Fetch the document

Import `useDocuments` and pull the single document for this content type:

```tsx
import { useDocuments } from '@/hooks/useDocuments'

export function SiteSettingsPanel({ contentType }: Props) {
  const { data: docs, isLoading } = useDocuments(contentType.ID)
  const doc = docs?.[0]

  if (isLoading) return <p className="text-muted-foreground">Loading…</p>
  if (!doc) return <p className="text-muted-foreground">No document found.</p>

  // continue in Step 4
}
```

---

## Step 4 — Wire save and publish mutations

```tsx
import {
  useUpdateDocument,
  usePublishDocument,
  useUnpublishDocument,
} from '@/hooks/useDocuments'
import { Button } from '@/components/ui/button'

export function SiteSettingsPanel({ contentType }: Props) {
  const { data: docs, isLoading } = useDocuments(contentType.ID)
  const doc = docs?.[0]

  const { mutateAsync: updateDoc } = useUpdateDocument()
  const publish = usePublishDocument()
  const unpublish = useUnpublishDocument()

  if (isLoading) return <p className="text-muted-foreground">Loading…</p>
  if (!doc) return <p className="text-muted-foreground">No document found.</p>

  const mutationFn = (data: Record<string, unknown>) =>
    updateDoc({ id: doc.ID, contentTypeId: contentType.ID, data })

  // continue in Step 5
}
```

---

## Step 5 — Build the form

Wrap fields in `FormProvider`. Each field uses `FormField` with a dot-notation `name` — nested names (e.g. `seo.title`) produce nested JSON on submit.

```tsx
import { FormProvider } from '@/components/form/FormProvider'
import { FormField } from '@/components/form/FormField'
import { TextInput } from '@/components/form/inputs/TextInput'
import { BooleanInput } from '@/components/form/inputs/BooleanInput'
import { api } from '@/lib/api'
import type { Document } from '@/types/cms'

// Inside the component, after resolving `doc` and `mutationFn`:
return (
  <div className="space-y-6">
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-xl font-semibold">{contentType.Name}</h1>
        <span className="text-sm text-muted-foreground capitalize">{doc.Status}</span>
      </div>
      <div className="flex gap-2">
        {doc.Status === 'draft' ? (
          <Button
            onClick={() => publish.mutate({ id: doc.ID, contentTypeId: contentType.ID })}
            disabled={publish.isPending}
          >
            Publish
          </Button>
        ) : (
          <Button
            variant="outline"
            onClick={() => unpublish.mutate({ id: doc.ID, contentTypeId: contentType.ID })}
            disabled={unpublish.isPending}
          >
            Unpublish
          </Button>
        )}
      </div>
    </div>

    <FormProvider
      query={{
        queryKey: ['documents', 'detail', doc.ID, 'data'],
        queryFn: () =>
          api.get<Document>(`/api/documents/${doc.ID}`).then((r) => r.data.Data),
      }}
      mutationFn={mutationFn}
    >
      <div className="space-y-4">
        <FormField name="siteName">
          <TextInput placeholder="My Site" aria-label="Site name" />
        </FormField>

        <FormField name="seo.title">
          <TextInput placeholder="SEO title" aria-label="SEO title" />
        </FormField>

        <FormField name="seo.description">
          <TextInput placeholder="SEO description" aria-label="SEO description" />
        </FormField>

        <FormField name="maintenanceMode">
          <BooleanInput aria-label="Maintenance mode" />
        </FormField>

        <Button type="submit">Save</Button>
      </div>
    </FormProvider>
  </div>
)
```

**Field name rules:**
- Flat name (`siteName`) → `{ siteName: "..." }` in the submitted JSON
- Dot-notation (`seo.title`) → `{ seo: { title: "..." } }` in the submitted JSON
- `FormProvider` handles the conversion automatically — no manual nesting required

---

## Step 6 — Register the route in router.tsx

Add a specific route **before** the generic `content-types/:slug` catch-all:

```tsx
// apps/web/src/router.tsx

const SiteSettingsPanel = lazy(() =>
  import('@/pages/admin/panels/SiteSettingsPanel').then((m) => ({
    default: m.SiteSettingsPanel,        // ← named export, not default
  })),
)
```

Then inside the `<Route path="/admin">` block:

```tsx
{/* Custom panel — must appear before the generic :slug catch-all */}
<Route
  path="content-types/site-settings"
  element={
    <Suspense fallback={<PanelFallback />}>
      <SiteSettingsPanelWrapper />
    </Suspense>
  }
/>
```

Because `SiteSettingsPanel` expects a `ContentType` prop, wrap it with a resolver:

```tsx
function SiteSettingsPanelWrapper() {
  const { data: contentTypes = [], isLoading } = useContentTypes()
  const ct = contentTypes.find((c) => c.Slug === 'site-settings')
  if (isLoading) return <PanelFallback />
  if (!ct) return <p className="text-muted-foreground">Content type not found.</p>
  return <SiteSettingsPanel contentType={ct} />
}
```

> **Tip:** If you want the generic panel to handle a content type, skip steps 2–6. The generic panel renders all field keys found in the document automatically. Custom panels are only needed when you want specific field names, labels, or rich input types (e.g. `RichTextInput`, `MediaInput`, `JsonInput`).

---

## Available input components

| Component | Import path | Use for |
|-----------|-------------|---------|
| `TextInput` | `@/components/form/inputs/TextInput` | Short text, URLs |
| `NumberInput` | `@/components/form/inputs/NumberInput` | Integers, decimals |
| `BooleanInput` | `@/components/form/inputs/BooleanInput` | Toggles, flags |
| `RichTextInput` | `@/components/form/inputs/RichTextInput` | Long-form HTML content (CKEditor) |
| `JsonInput` | `@/components/form/inputs/JsonInput` | Arbitrary JSON (CodeMirror) |
| `MediaInput` | `@/components/form/inputs/MediaInput` | Image / file upload via Cloudinary |

All inputs are `FormField`-compatible and follow the Shadcn UI styling conventions.
