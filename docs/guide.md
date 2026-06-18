# Feature Guide

Consolidated reference for all CMS features: panels, schema-as-code, draft/publish, localization, media, GraphQL, and config.

See [local-dev.md](local-dev.md) for environment setup and run commands.

---

## Table of Contents

1. [Panel walkthrough](#panel-walkthrough)
2. [Schema-as-code (content-type definitions)](#schema-as-code)
3. [Draft / publish workflow](#draft--publish-workflow)
4. [Localization (i18n)](#localization-i18n)
5. [Media upload and auto-thumbnail](#media-upload-and-auto-thumbnail)
6. [GraphQL API](#graphql-api)
7. [Config reference](#config-reference)

---

## Panel walkthrough

A panel is a hard-coded React page that edits a specific content type. The generic panel (`ContentTypePanelPage`) handles any content type automatically, but you can override it with a custom panel that controls exactly which fields appear and how.

This section walks through creating a custom **"Site Settings"** single-type panel as a concrete example.

### Overview — 6 steps

1. Define the content type as a JSON file (schema-as-code)
2. Create the panel component file
3. Fetch the document with a query hook
4. Wire mutations for save / publish
5. Build the form with `FormProvider` + `FormField` + inputs
6. Register the panel as a route in `router.tsx`

### Step 1 — Define the content type

Content-type **structure** is defined as code, not through the UI or API — there is no "create content type" button or endpoint. Add a JSON file under `apps/api/content-types/`:

```json
// apps/api/content-types/site-settings.json
{
  "slug": "site-settings",
  "name": "Site Settings",
  "kind": "single",
  "fields": [
    { "name": "siteName", "type": "text" },
    { "name": "seo.title", "type": "text" },
    { "name": "seo.description", "type": "text" },
    { "name": "maintenanceMode", "type": "boolean" }
  ]
}
```

Restart the API (`make dev-api`). On boot, `content_type.Sync` reads every file in this directory and reconciles it into MongoDB — creating the `ContentType` and its per-content-type document collection (`documents_site-settings`) automatically.

The sidebar shows **Site Settings** once synced. Clicking it loads the generic panel. For single types, the page shows an empty form until the first explicit Save. Steps 2–6 replace the generic panel with a custom one.

> Content **data** (the entries themselves) is unaffected by this — saving, publishing, and (for collection types) creating/deleting entries all still go through the normal UI and `/api/content-types/{slug}/documents` endpoints. Only the type's structure is JSON-only.

### Step 2 — Create the panel file

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

### Step 3 — Fetch the document

```tsx
import { useDocuments } from '@/hooks/useDocuments'

export function SiteSettingsPanel({ contentType }: Props) {
  const { data: docs, isLoading } = useDocuments(contentType.Slug)
  const doc = docs?.[0]

  if (isLoading) return <p className="text-muted-foreground">Loading…</p>
  if (!doc) return <p className="text-muted-foreground">No document yet — save to create one.</p>

  // continue in Step 4
}
```

### Step 4 — Wire save and publish mutations

```tsx
import {
  useUpdateDocument,
  usePublishDocument,
  useUnpublishDocument,
} from '@/hooks/useDocuments'
import { Button } from '@/components/ui/button'

export function SiteSettingsPanel({ contentType }: Props) {
  const { data: docs, isLoading } = useDocuments(contentType.Slug)
  const doc = docs?.[0]

  const { mutateAsync: updateDoc } = useUpdateDocument()
  const publish = usePublishDocument()
  const unpublish = useUnpublishDocument()

  if (isLoading) return <p className="text-muted-foreground">Loading…</p>
  if (!doc) return <p className="text-muted-foreground">No document yet — save to create one.</p>

  const mutationFn = (data: Record<string, unknown>) =>
    updateDoc({ contentTypeSlug: contentType.Slug, id: doc.DocumentID, data })

  // continue in Step 5
}
```

### Step 5 — Build the form

Wrap fields in `FormProvider`. Each field uses `FormField` with a dot-notation `name` — nested names (e.g. `seo.title`) produce nested JSON on submit.

```tsx
import { FormProvider } from '@/components/form/FormProvider'
import { FormField } from '@/components/form/FormField'
import { TextInput } from '@/components/form/inputs/TextInput'
import { BooleanInput } from '@/components/form/inputs/BooleanInput'
import { api } from '@/lib/api'
import type { Document } from '@/types/cms'

// Status is tri-state: "draft" (never published), "modified" (draft has
// unpublished changes), "published" (draft and published are in sync).
const canPublish = doc.Status !== 'published'
const canUnpublish = doc.Status !== 'draft'

return (
  <div className="space-y-6">
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-xl font-semibold">{contentType.Name}</h1>
        <span className="text-sm text-muted-foreground capitalize">{doc.Status}</span>
      </div>
      <div className="flex gap-2">
        {canPublish && (
          <Button
            onClick={() => publish.mutate({ contentTypeSlug: contentType.Slug, id: doc.DocumentID })}
            disabled={publish.isPending}
          >
            Publish
          </Button>
        )}
        {canUnpublish && (
          <Button
            variant="outline"
            onClick={() => unpublish.mutate({ contentTypeSlug: contentType.Slug, id: doc.DocumentID })}
            disabled={unpublish.isPending}
          >
            Unpublish
          </Button>
        )}
      </div>
    </div>

    <FormProvider
      query={{
        queryKey: ['documents', 'detail', contentType.Slug, doc.DocumentID, 'data'],
        queryFn: () =>
          api.get<Document>(`/api/content-types/${contentType.Slug}/documents/${doc.DocumentID}`).then((r) => r.data.Data),
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

### Step 6 — Register the route in router.tsx

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

### Available input components

| Component | Import path | Use for |
|-----------|-------------|---------|
| `TextInput` | `@/components/form/inputs/TextInput` | Short text, URLs |
| `NumberInput` | `@/components/form/inputs/NumberInput` | Integers, decimals |
| `BooleanInput` | `@/components/form/inputs/BooleanInput` | Toggles, flags |
| `RichTextInput` | `@/components/form/inputs/RichTextInput` | Long-form HTML content (CKEditor) |
| `JsonInput` | `@/components/form/inputs/JsonInput` | Arbitrary JSON (CodeMirror) |
| `MediaInput` | `@/components/form/inputs/MediaInput` | Image / file upload |

All inputs are `FormField`-compatible and follow the Shadcn UI styling conventions.

---

## Schema-as-code

Content-type **structure** (fields, types, `kind`) is defined in JSON files under `apps/api/content-types/*.json` — never created or edited via the API or UI.

### How sync works

On every API startup, `usecase/content_type.Sync` reads all definition files and reconciles them against MongoDB:

| Event | Sync action |
|-------|-------------|
| New file | Creates the `ContentType` and its per-content-type document collection (`documents_<slug>`) with indexes |
| Changed file (fields added/changed) | Updates the `ContentType` schema in place; ensures the collection exists |
| Field removed from a file | Drops the field from the schema; existing entry data is untouched (orphaned key stays in MongoDB) |
| File deleted | Deletes the `ContentType`, cascade-deletes all its entries (draft + published), and drops the collection |

Sync is one-directional: JSON files → MongoDB. Nothing the UI or API does writes back to the files.

### JSON schema format

```json
{
  "slug": "blog-post",
  "name": "Blog Post",
  "kind": "collection",
  "fields": [
    { "name": "title", "type": "text" },
    { "name": "body", "type": "richtext" },
    { "name": "published", "type": "boolean" }
  ]
}
```

`kind` is `"single"` (at most one entry, created on first save) or `"collection"` (many entries).

System fields (`createdAt`, `updatedAt`, `publishedAt`, `createdBy`, `updatedBy`, `publishedBy`, `locale`) are injected automatically on every record — never declared in the schema.

---

## Draft / publish workflow

Every content entry follows a **draft → publish** model. Two separate records — a `draft` record and a `published` record — share the same `documentId` within a per-content-type MongoDB collection (`documents_<slug>`).

### States (computed, never stored)

| Status | Meaning |
|--------|---------|
| `draft` | No published record exists yet |
| `modified` | Draft has changes since the last publish (`draft.updatedAt > published.updatedAt`) |
| `published` | Draft and published are in sync |

### Operations

- **Save**: upserts the `draft` record. Never touches the `published` record. Draft is invisible to the public read API.
- **Publish**: copies `draft.data` → `published` record (creates it if absent). Public read API serves this immediately.
- **Unpublish**: deletes the `published` record, reverting status to `draft`. Public read API returns 404 again until the next Publish.

Unpublish is exposed in `SingleTypePanel` and `CollectionDetailPanel` whenever `status !== 'draft'`.

### Public read API

`GET /api/public/content-types/:slug/documents/:documentId` resolves the `published` record only. If no published record exists for the requested `(documentId, locale)` pair, it returns 404 — the draft is never visible to readers.

---

## Localization (i18n)

Locale variants extend the draft/publish model: each `documentId` may have an independent draft+published pair **per locale**, sharing the same `documentId` but with distinct `locale` values, within the same per-content-type collection.

### Key rules

- Supported locales are fixed by `SUPPORTED_LOCALES` config (e.g. `en,vi`). Saving a draft with an unsupported locale is rejected.
- Each locale variant has its own computed `status`. Publishing `en` never changes `vi`.
- The public read API resolves `published` for a given `(documentId, locale)` pair. The 404-when-unpublished rule applies per locale.
- Localization is whole-entry — all fields are localized together. There is no per-field `localized` flag.

### Admin UI

A locale switcher appears in `SingleTypePanel`, `CollectionDetailPanel`, and `ContentTypePanelPage`. The editor picks the active locale before saving or publishing.

### FE hooks

All document hooks accept a `locale` parameter:

```ts
const { data } = useDocuments(contentTypeSlug)
const { mutate } = useUpdateDocument()
mutate({ contentTypeSlug, id: documentId, locale: 'vi', data })
```

`useLocales()` fetches the supported locales list from `GET /api/locales` (public, no auth).

---

## Media upload and auto-thumbnail

### Upload

`POST /api/media/upload` (multipart/form-data, admin only). Returns a `MediaAsset` with `sourceURL` and `thumbnailURL`.

Use `MediaInput` in a panel form:

```tsx
<FormField name="coverImage">
  <MediaInput aria-label="Cover image" />
</FormField>
```

`MediaInput` renders both the original and a second thumbnail preview when `thumbnailURL !== sourceURL`.

### Auto-thumbnail toggle

Controlled by `MEDIA_AUTO_THUMBNAIL` (default `true`):

| Value | Behavior |
|-------|---------|
| `true` | Storage adapter generates a resized thumbnail at upload time. `thumbnailURL` is stored separately from `sourceURL`. |
| `false` | Only the original is uploaded. `thumbnailURL` falls back to the provider's native on-the-fly capability, or `sourceURL` if unavailable. |

Both S3 and Cloudinary adapters implement this behind the same `StorageAdapter` interface — no adapter-specific branching in the media usecase.

### Storage provider selection

`STORAGE_PROVIDER` (env var) selects the active adapter: `s3` or `cloudinary`. Cloudinary performs real eager thumbnail generation; S3 always sets `thumbnailURL == sourceURL` (no native eager thumbnail).

---

## GraphQL API

The CMS exposes a GraphQL endpoint alongside REST — not a replacement for it.

### Endpoint

Mounted at `GRAPHQL_PATH` (default `/graphql`). Access via standard HTTP POST:

```sh
curl -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -d '{"query": "{ publishedDocument(contentTypeSlug: \"blog\", documentId: \"abc\", locale: \"en\") { data } }"}'
```

### Schema overview

```graphql
type Query {
  publishedDocument(contentTypeSlug: String!, documentId: ID!, locale: String): Document
  contentTypes: [ContentType!]!
}

type Mutation {
  saveDocument(contentTypeSlug: String!, documentId: ID!, locale: String, data: JSON!): Document! @auth
  publishDocument(contentTypeSlug: String!, documentId: ID!, locale: String): Document! @auth
  unpublishDocument(contentTypeSlug: String!, documentId: ID!, locale: String): Document! @auth
  deleteDocument(contentTypeSlug: String!, documentId: ID!): Boolean! @auth
}
```

All GraphQL operations require `contentTypeSlug` to route to the correct per-content-type collection, plus `documentId` to identify the specific entry.

`Query` operations are public (no auth required). `Mutation` operations require the same JWT bearer auth as REST, enforced via the `@auth` directive.

### Resolvers

Resolvers call the existing usecases (`document`, `content_type`, `media`) — no business logic is duplicated. This means the same validation, locale checks, and draft/publish semantics apply to both REST and GraphQL paths.

### Regenerating after schema changes

Edit `apps/api/graphql/schema.graphqls`, then:

```sh
make graphql-generate
```

Generated code in `apps/api/graphql/generated/` is never hand-edited.

---

## Config reference

All environment variables are loaded once at boot into a typed `Config` struct (`internal/config`). No `os.Getenv` calls exist anywhere else in the codebase.

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | API listen port | `8080` |
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017/cms` |
| `JWT_SECRET` | Secret for signing JWTs | *(required, no default)* |
| `CLOUDINARY_CLOUD_NAME` | Cloudinary account name | *(required for Cloudinary media)* |
| `CLOUDINARY_API_KEY` | Cloudinary API key | *(required for Cloudinary media)* |
| `CLOUDINARY_API_SECRET` | Cloudinary API secret | *(required for Cloudinary media)* |
| `CONTENT_TYPES_DIR` | Directory of JSON content-type definition files synced on boot | `content-types` |
| `STORAGE_PROVIDER` | Active media storage adapter | `s3` or `cloudinary` |
| `SUPPORTED_LOCALES` | Comma-separated locale codes accepted when saving drafts | `en,vi` |
| `MEDIA_AUTO_THUMBNAIL` | Toggle server-side thumbnail generation at upload time | `true` |
| `GRAPHQL_PATH` | Mount path for the GraphQL endpoint | `/graphql` |
| `VITE_API_URL` | API base URL used by the Vite dev-server proxy | `http://localhost:8080` |

Copy `.env.example` → `.env` and fill in required values. Never commit `.env`.

For full setup instructions, see [local-dev.md](local-dev.md).
