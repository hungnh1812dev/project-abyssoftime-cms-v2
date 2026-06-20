# Archive — Phases 0–5: Foundation Build

> All phases completed. Archived from `tasks/todo.md`.

---

## Phase 0 — Foundation
- [x] T0.1 Monorepo scaffold (dir tree, go mod init, npm create vite, SPEC.md copy)
- [x] T0.2 Domain layer — entities (User, Document, ContentType, MediaAsset) + repository interfaces
- [x] T0.3 Infrastructure base — MongoDB client, pkg/errors, pkg/jwt stubs, /health endpoint
- [x] T0.4 Frontend base — Vite + React Router + TanStack Query + Shadcn + api.ts + queryClient.ts
- [x] ✅ Checkpoint 0: /health → 200, React app loads, go vet + npm build pass

## Phase 1 — Auth
- [x] T1.1 User MongoDB repository (Create, FindByEmail, FindById) + unit tests
- [x] T1.2 JWT package (GenerateAccessToken, GenerateRefreshToken, ValidateToken)
- [x] T1.3 Auth usecase (Register, Login, RefreshToken, Logout) + tests ≥ 80% coverage
- [x] T1.4 Auth HTTP handlers + JWT middleware (HttpOnly cookie for refresh token)
- [x] T1.5 FE: Axios API client with JWT interceptors + 401 auto-refresh
- [x] T1.6 FE: Login + Register pages (react-hook-form, useMutation)
- [x] T1.7 FE: Auth context + ProtectedRoute + AdminRoute guards
- [x] ✅ Checkpoint 1: Register → Login → admin shell; refresh token survives tab close; 401 on unauth requests

## Phase 2 — Form System
- [x] T2.1 FormProvider + FormField (dot-notation → nested JSON, no React.Children.map)
- [x] T2.2 TextInput + NumberInput + BooleanInput (Shadcn-wrapped, FormField-compatible)
- [x] T2.3 RichTextInput (CKEditor, controlled via Controller)
- [x] T2.4 JsonInput (CodeMirror, validates JSON, outputs parsed object)
- [x] ✅ Checkpoint 2: Test panel with all inputs; submit shows nested JSON; lint + TS pass

## Phase 3 — Core CMS (old single-record model — migrated by Phase A–C)
- [x] T3.1 ContentType MongoDB repository (Create, FindByID, FindBySlug, FindAll, Update, Delete)
- [x] T3.2 ContentType usecase (CRUD, unique slug, validate kind) + tests
- [x] T3.3 ContentType REST handlers (GET/POST/PUT/DELETE /api/content-types, admin only)
- [x] T3.4 Document MongoDB repository (CRUD + UpdateStatus; cascade via usecase)
- [x] T3.5 Document usecase (CRUD + draft/publish state machine + cascade delete) + tests
- [x] T3.6 Document REST handlers (GET/POST/PUT/DELETE + publish/unpublish; guest = read-only)
- [x] T3.7 FE: TanStack Query hooks for documents + content types (invalidateQueries on all mutations)
- [x] T3.8 FE: Admin layout (AdminLayout, Sidebar with content-type list, TopBar)
- [x] T3.9 FE: Single-type panel example (FormProvider + useDocument + useUpdateDocument + Draft/Publish)
- [x] T3.10 FE: Collection-type list + detail panel (CollectionListPage + CollectionDetailPanel)
- [x] ✅ Checkpoint 3: Edit → publish → UI updates; delete → cascade; usecase coverage ≥ 80%

## Phase 4 — Media Upload
- [x] T4.1 MediaAsset MongoDB repository + StorageAdapter interface (DeleteByDocumentRef)
- [x] T4.2 Cloudinary adapter (cloudinary-go/v2, Upload → SecureURL + PublicID)
- [x] T4.3 Media usecase + POST /api/media/upload handler
- [x] T4.4 FE: MediaInput component + useUploadMedia hook
- [x] ✅ Checkpoint 4: Upload image → thumbnail; submit → Cloudinary URL in doc; delete → MediaAsset removed

## Phase 5 — CI/CD + Docker
- [x] T5.1 Dockerfiles (multi-stage Go + nginx/React)
- [x] T5.2 docker-compose.yml (mongodb + api + web, single Render entrypoint)
- [x] T5.3 GitHub Actions (api-test + web-lint + docker build + Render deploy hook)
- [x] ✅ Checkpoint 5 (Final): docker-compose up healthy; CI green on push to main; Render deploys
