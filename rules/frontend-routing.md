# RULES — Frontend Routing

**Scope:** React Router routes, content-type registry, sidebar navigation, testing.

---

## 1. Routing Rules

### 1.1 Route Structure
| Route | Component | Description |
|---|---|---|
| `/admin/content-type/single-type/:slug` | SingleTypePage | Single-type editor |
| `/admin/content-type/collection-type/:slug` | CollectionListPage | Collection list |
| `/admin/content-type/collection-type/:slug/new` | CollectionDetailPage | Create entry |
| `/admin/content-type/collection-type/:slug/:id` | CollectionDetailPage | Edit entry |
| `/admin/settings/internationalize` | InternationalizePage | Locale management |
| `/admin/settings/media` | MediaLibraryPage | Media management |
| `/admin/settings/users` | UsersPage | User management |
| `/admin/settings/access-tokens` | AccessTokensPage | Token management |
| `/admin/settings/roles` | RolesPage | Permission matrix |

### 1.2 Custom Panels
- Registered as specific routes **before** generic catch-all
- Receive `ContentType` prop via a wrapper component
- Use lazy loading with named exports

### 1.3 Protected Routes
- `ProtectedRoute` component with optional `minRole` prop
- Settings links filtered by role in sidebar

---

## 2. Content-Type Registry Rules

### 2.1 Purpose
- Metadata-only module at `src/content-type-registry/index.ts`
- Maps content-type slugs to custom panel configuration
- Registry overrides take precedence over dynamic behavior

### 2.2 Column Override
- If a registry entry defines `columns`, the column chooser button is hidden
- Non-registry content types use the dynamic column chooser

---

## 3. Sidebar Rules

### 3.1 Content Types
- Eager metadata load from API
- Lazy component loading via `React.lazy`
- Grouped by kind (single-type, collection-type)

### 3.2 Settings Links (Role-Filtered)
| Link | Visible To |
|---|---|
| Media Library | All authenticated |
| Users | admin+ |
| Access Tokens | super_admin only |
| Roles | super_admin only |
| Internationalize | super_admin only |

---

## 4. Testing Rules (Frontend-Specific)

### 4.1 Test Stack
- Vitest + React Testing Library + MSW
- Tests in co-located `__tests__/` directories

### 4.2 Test Patterns
- Test user behavior, not implementation details
- Use `screen.getByRole`, `screen.getByText`, `screen.getByLabelText`
- Mock API calls with MSW, not by mocking hooks
- Verify query invalidation happens after mutations

### 4.3 Required Tests
- Every new component needs test file
- Every new hook needs test file
- Page tests: renders correctly, handles loading/error states
- Form tests: submit, validation, dirty state
- Dialog tests: open/close, confirm/cancel actions

---

## 5. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | Lazy-load panels with `React.lazy` |
| **Always** | Co-locate tests in `__tests__/` directories |
