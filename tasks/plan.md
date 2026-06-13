# Plan — personal-cms (project-abyssoftime-cms-v2)

## Context
Greenfield Personal Headless CMS. No existing code. Building a monorepo at
`~/workspace/dev/project-abyssoftime/project-abyssoftime-cms-v2/` with:
- `apps/api` — Go backend, Clean Architecture, MongoDB, REST, JWT auth
- `apps/web` — React/Vite frontend, TanStack Query, react-hook-form, Shadcn UI

Decisions locked in:
- Go module path: `project-abyssoftime-cms-v2/api`
- Media storage: Cloudinary
- Render deploy: single service running docker-compose
- API build order: REST first (GraphQL deferred)
- DB: MongoDB only (PostgreSQL deferred)

---

## Dependency Graph

```
Phase 0 (Foundation)
  T0.1 Monorepo scaffold ────────────────────────────┐
  T0.2 Domain entities + repo interfaces ─────────── all backend phases
  T0.3 MongoDB connection + pkg/errors + pkg/jwt ─── Phase 1, 3
  T0.4 Frontend base (Vite + Router + TanStack + Shadcn) ── all FE phases

Phase 1 (Auth)          depends on T0.x
  T1.1 User entity + UserRepository (interface + Mongo impl)
  T1.2 JWT pkg (access + refresh token gen/validate)
  T1.3 Auth usecase (register / login / refresh / logout)
  T1.4 Auth handlers + JWT middleware
  T1.5 FE: API client (Axios + JWT interceptors + token refresh)
  T1.6 FE: Login + Register pages (react-hook-form)
  T1.7 FE: Auth context + protected route + role guard

Phase 2 (Form System)   depends on T0.4
  T2.1 FormProvider + FormField (core form context)
  T2.2 TextInput + NumberInput + BooleanInput
  T2.3 RichTextInput (CKEditor)
  T2.4 JsonInput (CodeMirror)

Phase 3 (Core CMS)      depends on T0.x, T1.x, T2.x
  T3.1 ContentType entity + repo interface + Mongo impl
  T3.2 ContentType usecase (CRUD)
  T3.3 ContentType REST handlers
  T3.4 Document entity + repo interface + Mongo impl (cascade delete)
  T3.5 Document usecase (CRUD + draft/publish)
  T3.6 Document REST handlers
  T3.7 FE: TanStack Query hooks (all document + contentType hooks)
  T3.8 FE: Admin layout (shell + sidebar + nav)
  T3.9 FE: Single-type panel example (uses FormProvider + inputs)
  T3.10 FE: Collection-type panel (list page + detail panel)

Phase 4 (Media)         depends on T3.x
  T4.1 MediaAsset entity + StorageAdapter interface + Mongo impl
  T4.2 Cloudinary storage adapter
  T4.3 Media usecase + REST handler
  T4.4 FE: MediaInput + useUploadMedia

Phase 5 (CI/CD)         depends on all phases
  T5.1 Dockerfiles (api + web)
  T5.2 docker-compose.yml (api + web + mongodb)
  T5.3 GitHub Actions (lint → test → build → Render deploy hook)
```

---

## Phase 0 — Foundation

### T0.1 — Monorepo Scaffold
Create the directory tree:
```
project-abyssoftime-cms-v2/
├── apps/api/           → go mod init project-abyssoftime-cms-v2/api
├── apps/web/           → npm create vite@latest . -- --template react-ts
├── tasks/              → plan.md, todo.md
└── SPEC.md             (copy from agent-skills)
```
**Acceptance:** `cd apps/api && go run ./cmd/server` starts without errors and returns 200 on `GET /health`. `cd apps/web && npm run dev` opens the Vite default page.

### T0.2 — Domain Layer
Files: `apps/api/internal/domain/entity/*.go`, `apps/api/internal/domain/repository/*.go`

Entities to define (structs + field tags for MongoDB):
- `User` (id, email, passwordHash, role, createdAt)
- `Document` (documentId, contentTypeId, status [draft|published], data map, createdAt, updatedAt)
- `ContentType` (documentId, name, slug, kind [single|collection])
- `MediaAsset` (documentId, url, publicId, contentTypeId, documentRef)

Repository interfaces (pure Go interfaces, zero imports of DB packages):
- `UserRepository`, `DocumentRepository`, `ContentTypeRepository`, `MediaAssetRepository`

**Acceptance:** `go build ./internal/domain/...` succeeds. No DB imports in this package.

### T0.3 — Infrastructure Base + Shared Packages
- `pkg/errors/` — domain error types (`ErrNotFound`, `ErrUnauthorized`, `ErrConflict`)
- `pkg/jwt/` — token structs (defined here, implementation in T1.2)
- `internal/infrastructure/mongodb/client.go` — MongoDB connection factory
- `cmd/server/main.go` — wire MongoDB connection; expose `/health`

**Acceptance:** `go test ./pkg/...` passes. Server connects to MongoDB on startup.

### T0.4 — Frontend Base
In `apps/web`:
- Install: `@tanstack/react-query`, `react-router-dom`, `axios`, `react-hook-form`
- Install Shadcn UI + TailwindCSS per Shadcn CLI
- `src/lib/queryClient.ts` — `QueryClient` singleton with `staleTime: 30_000`
- `src/lib/api.ts` — Axios instance (base URL from env var, interceptors stubbed)
- `src/main.tsx` — wrap app in `<QueryClientProvider>` + `<BrowserRouter>`
- `src/router.tsx` — placeholder routes (`/login`, `/admin/*`)

**Acceptance:** `npm run build` succeeds with zero TypeScript errors. App loads in browser, no console errors.

---

## ✅ Checkpoint 0
- `GET /health` → 200
- React app loads in browser, router works
- `go vet ./...` and `npm run lint` both pass

---

## Phase 1 — Auth

### T1.1 — User Entity + Repository
- `internal/infrastructure/mongodb/user_repository.go` — implements `UserRepository`
- Methods: `Create`, `FindByEmail`, `FindById`
- Password: store bcrypt hash; never store plain text
- Write unit test with mock `UserRepository` (table-driven)

**Acceptance:** `go test ./internal/usecase/auth/...` passes with mock repo.

### T1.2 — JWT Package
File: `pkg/jwt/jwt.go`
- `GenerateAccessToken(userID, role string) (string, error)` — 15 min TTL
- `GenerateRefreshToken(userID string) (string, error)` — 7 day TTL
- `ValidateToken(token string) (*Claims, error)`
- Secret loaded from env var (never hard-coded)

**Acceptance:** `go test ./pkg/jwt/...` passes (generate → validate round-trip).

### T1.3 — Auth Usecase
File: `internal/usecase/auth/auth_usecase.go`
- `Register(email, password string) (*User, error)` — hashes password, calls repo
- `Login(email, password string) (accessToken, refreshToken string, error)` — validates creds, returns both tokens
- `RefreshToken(refreshToken string) (newAccessToken string, error)`
- `Logout(userID string) error`

No DB imports. Depends only on `UserRepository` interface and `pkg/jwt`.

**Acceptance:** `go test ./internal/usecase/auth/...` ≥ 80% coverage, all with mock repo.

### T1.4 — Auth Handlers + Middleware
Files in `internal/delivery/http/`:
- `handler/auth_handler.go` — `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`
- `middleware/auth.go` — validate `Authorization: Bearer` header; attach `userID` + `role` to context
- Refresh token travels in `HttpOnly` cookie (set/read in handler, never in response body)
- Handler maps domain errors to HTTP status codes

**Acceptance:** `go test ./internal/delivery/http/handler/...` with `httptest`; login → 200 with access token in body + refresh cookie set.

### T1.5 — Frontend API Client + Token Management
File: `src/lib/api.ts`
- Axios instance with `Authorization: Bearer <access_token>` request interceptor
- 401 response interceptor → call `POST /auth/refresh` → retry original request once
- Token stored in React context (in-memory, not localStorage)

### T1.6 — Login + Register Pages
Files: `src/pages/auth/LoginPage.tsx`, `src/pages/auth/RegisterPage.tsx`
- Both use `react-hook-form` directly (no `FormProvider` — auth forms are outside the CMS panel system)
- `useMutation` for submit; on success → store access token in auth context → navigate to `/admin`
- Validation: email format, password min length 8

**Acceptance:** Can register → login → land on `/admin` (placeholder). Invalid credentials show error toast.

### T1.7 — Auth Context + Route Guards
Files: `src/hooks/useAuth.ts`, `src/router.tsx`
- `<ProtectedRoute>` — redirects to `/login` if no access token
- `<AdminRoute>` — renders 403 if role is not `admin`
- Token refresh on app load (call `/auth/refresh` from the stored cookie)

---

## ✅ Checkpoint 1
- Register new user → Login → land on admin shell
- Refresh token: close and reopen browser tab → still logged in
- Unauthenticated `GET /api/documents` → 401
- Guest role → admin-only route → 403

---

## Phase 2 — Form System

### T2.1 — FormProvider + FormField
Files: `src/components/form/FormProvider.tsx`, `src/components/form/FormField.tsx`

`FormProvider`:
- Wraps `react-hook-form`'s `<FormProvider>`
- Props: `queryFn` (TanStack `useQuery` config), `mutationFn` (TanStack `useMutation` config), `onSuccess?`
- Internally calls `useQuery` → feeds result into `useForm({ defaultValues })`
- On submit: serializes dot-notation field names into nested JSON, calls `mutationFn`
- Exposes `loading` (from `isFetching`) and `submitting` (from `isPending`) to children via context
- **MUST NOT** use `React.Children.map` or any recursive child scanning

`FormField`:
- Consumes `useFormContext()` → extracts `register`, `control`, `formState`
- Passes them as props to its single child via `React.cloneElement`
- Handles error message display
- Supports dot-notation `name` (e.g., `block.house.title`)

**Acceptance:** Unit test — render `<FormProvider> + <FormField name="a.b"> + <input>`, submit → assert payload is `{ a: { b: value } }`. No `React.Children.map` in source.

### T2.2 — TextInput + NumberInput + BooleanInput
Files: `src/components/form/inputs/TextInput.tsx`, `NumberInput.tsx`, `BooleanInput.tsx`

- `TextInput` — wraps Shadcn `Input`; optional `multiline` prop uses `Textarea`
- `NumberInput` — wraps Shadcn `Input` with `type="number"`; `step`, `min`, `max` props
- `BooleanInput` — wraps Shadcn `Switch`; uses `control` (Controller) not `register`

Each component: accepts `register` and `control` props injected by `FormField`.

**Acceptance:** Vitest — each input renders, accepts value, fires onChange; FormField injects props correctly.

### T2.3 — RichTextInput (CKEditor)
File: `src/components/form/inputs/RichTextInput.tsx`
- Install `@ckeditor/ckeditor5-react` + classic build
- Controlled via `Controller` from react-hook-form (`control` prop)
- Output: HTML string stored in form state

**Acceptance:** Component renders CKEditor; typing updates form value; submit captures HTML.

### T2.4 — JsonInput (CodeMirror)
File: `src/components/form/inputs/JsonInput.tsx`
- Install `@codemirror/lang-json` + `@uiw/react-codemirror`
- Controlled via `Controller`; validates JSON on change (highlight syntax errors)
- Output: parsed JSON object (not string) in form state

**Acceptance:** Valid JSON → form value is parsed object. Invalid JSON → inline error shown, form blocks submission.

---

## ✅ Checkpoint 2
- Render a test panel page with all 4 input types
- Submit form with nested names → console.log shows correctly nested JSON
- `npm run lint` passes; no TypeScript errors
- `go test ./...` still passes

---

## Phase 3 — Core CMS

### T3.1 — ContentType Repo + Mongo Impl
File: `internal/infrastructure/mongodb/content_type_repository.go`
- Implements `ContentTypeRepository`: `Create`, `FindByID`, `FindBySlug`, `FindAll`, `Update`, `Delete`
- `documentId` = MongoDB `_id` stringified (ObjectID.Hex())

### T3.2 — ContentType Usecase
File: `internal/usecase/content_type/content_type_usecase.go`
- CRUD operations; slug must be unique (check via repo before create)
- Validate `kind` is `single` or `collection`

**Acceptance:** `go test ./internal/usecase/content_type/...` with mock repo.

### T3.3 — ContentType REST Handlers
Routes (all require `admin` role via middleware):
- `GET /api/content-types`
- `POST /api/content-types`
- `GET /api/content-types/:id`
- `PUT /api/content-types/:id`
- `DELETE /api/content-types/:id`

### T3.4 — Document Repo + Mongo Impl
File: `internal/infrastructure/mongodb/document_repository.go`
- Implements `DocumentRepository`: `Create`, `FindByID`, `FindByContentType`, `Update`, `UpdateStatus`, `Delete`
- Cascade rule enforced at **usecase layer**: usecase calls `MediaAssetRepository.DeleteByDocumentRef` before `DocumentRepository.Delete`

### T3.5 — Document Usecase
File: `internal/usecase/document/document_usecase.go`
- `Create`, `Update`, `GetOne`, `GetAll`, `Delete`, `Publish`, `Unpublish`
- Status machine: `draft` ↔ `published`
- `Delete` calls `MediaAssetRepository.DeleteByDocumentRef` first

**Acceptance:** Table-driven tests covering status transitions and cascade delete sequence.

### T3.6 — Document REST Handlers
Routes (JWT middleware; `guest` role can only GET):
- `GET /api/documents?contentType=:slug`
- `POST /api/documents`
- `GET /api/documents/:id`
- `PUT /api/documents/:id`
- `DELETE /api/documents/:id`
- `POST /api/documents/:id/publish`
- `POST /api/documents/:id/unpublish`

### T3.7 — Frontend TanStack Query Hooks
Files in `src/hooks/`:
- `useContentTypes`, `useContentType`, `useDocumentList`, `useDocument`
- `useCreateDocument`, `useUpdateDocument`, `usePublishDocument`, `useDeleteDocument`

Every mutation `onSuccess` → `queryClient.invalidateQueries` on affected keys.

### T3.8 — Admin Layout
Files: `src/pages/admin/layout/AdminLayout.tsx`, `Sidebar.tsx`, `TopBar.tsx`
- Sidebar: `useContentTypes` → list of panel links
- TopBar: user info + logout

### T3.9 — Single-Type Panel Example
File: `src/pages/admin/panels/SingleTypePanel.tsx`
- `<FormProvider>` with `useDocument` + `useUpdateDocument`
- Draft / Publish action buttons

**Acceptance:** Edit → Save → UI reflects immediately without page reload.

### T3.10 — Collection-Type Panel
Files: `src/pages/admin/panels/CollectionListPage.tsx`, `CollectionDetailPanel.tsx`
- List: `useDocumentList` → table + Edit/Delete actions
- Detail: `FormProvider` pattern same as T3.9

---

## ✅ Checkpoint 3
- Edit single-type document → publish → status badge updates without reload
- Collection: create 2 entries, delete 1 → list auto-updates
- Delete document → MediaAssets removed from DB
- `go test ./...` ≥ 80% on usecase packages

---

## Phase 4 — Media Upload

### T4.1 — MediaAsset Repo + Storage Interface
- `domain/repository/storage_adapter.go` — `StorageAdapter` interface
- `internal/infrastructure/mongodb/media_asset_repository.go` — `DeleteByDocumentRef`

### T4.2 — Cloudinary Adapter
File: `internal/infrastructure/storage/cloudinary_adapter.go`
- `github.com/cloudinary/cloudinary-go/v2`
- Returns `SecureURL` + `PublicID`; credentials from env vars

### T4.3 — Media Usecase + Handler
- `POST /api/media/upload` (multipart/form-data, admin only)
- Returns `{ url, publicId, documentId }`

### T4.4 — MediaInput + useUploadMedia
File: `src/components/form/inputs/MediaInput.tsx`
- Hidden `<input type="file">` behind Shadcn `Button`
- On select → upload → set URL in form state
- Shows thumbnail for images

---

## ✅ Checkpoint 4
- Upload image → thumbnail appears → submit → Cloudinary URL in document data
- Delete document → MediaAsset removed from DB

---

## Phase 5 — CI/CD + Docker

### T5.1 — Dockerfiles
- `apps/api/Dockerfile`: multi-stage Go build → alpine runtime, port 8080
- `apps/web/Dockerfile`: multi-stage Node build → nginx serve, proxies `/api`

### T5.2 — docker-compose.yml
Services: `mongodb` (mongo:7 + volume), `api`, `web`
Single entrypoint for Render.com: `docker-compose up --build`

### T5.3 — GitHub Actions
```
on: push (main), pull_request
jobs:
  api-test:   go vet + go test ./...
  web-lint:   npm ci + npm run lint + npm run build
  docker:     docker-compose build
  deploy:     curl Render deploy hook (main only)
```

---

## ✅ Checkpoint 5 (Final)
- `docker-compose up` → all services healthy
- Push to main → CI green → Render deploys
- `go test ./...` passes in CI; `npm run build` passes in CI
