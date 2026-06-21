# Plan — UI Design System (Strapi-Inspired)

Spec: [specs/ui-design-system.md](../specs/ui-design-system.md)

---

## Dependency Graph

```
T1 (Color tokens)
 ├──→ T2 (Button enhancement)  ← needs new tokens (primary-hover, success, etc.)
 ├──→ T3 (Status badges)       ← needs success/warning tokens
 └──→ T4 (Sidebar context)     ← needs sidebar-* tokens
        │
        ├──→ T5 (Sidebar components) ← needs SidebarContext
        │     │
        │     └──→ T6 (Sidebar responsive overlay) ← needs Sidebar + SidebarContext
        │
        └──→ T7 (AdminLayout integration) ← needs Sidebar + SidebarContext
              │
              ├──→ T8 (Breadcrumbs + TopBar) ← needs AdminLayout updated
              │
              └──→ T9 (Sticky action bar) ← needs Button enhancement (T2) + TopBar (T8)
                    │
                    └──→ T10 (Page-level integration) ← needs action bar + cards
                          │
                          └──→ T11 (Dark mode verification) ← needs all above
```

### Parallelizable Groups

- **Independent**: T1 (must go first — all others depend on it)
- **After T1**: T2, T3, T4 can run in parallel
- **After T4**: T5
- **After T5**: T6, T7 can run in parallel
- **After T7**: T8
- **After T2 + T8**: T9
- **After T9**: T10
- **After T10**: T11

---

## Tasks

### T1 — Color Tokens: Migrate to Indigo Primary

**Files**: `apps/web/src/index.css`

**What**: Replace current neutral grayscale `:root` and `.dark` token values with indigo-based system. Add new tokens: `--primary-hover`, `--primary-active`, `--primary-muted`, `--success`, `--success-foreground`, `--warning`, `--warning-foreground`, `--sidebar-muted`. Update `@theme inline` block to register new color tokens.

**Changes**:
- `:root` block: `--primary` → `oklch(0.488 0.243 264.376)`, `--background` → `oklch(0.968 0.001 247)`, `--ring` → match primary, `--sidebar` → `oklch(0.205 0 0)` (dark bg), all sidebar tokens updated
- `.dark` block: `--primary` → `oklch(0.623 0.214 259)`, sidebar tokens updated
- `@theme inline`: add `--color-primary-hover`, `--color-primary-active`, `--color-primary-muted`, `--color-success`, `--color-success-foreground`, `--color-warning`, `--color-warning-foreground`, `--color-sidebar-muted`

**Acceptance criteria**:
- `npm run build` passes with no Tailwind errors
- All existing UI still renders (no broken color references)
- New tokens `bg-primary-hover`, `text-success`, `text-warning`, `text-sidebar-muted` resolve in Tailwind

**Verify**: `cd apps/web && npm run build` — zero errors

---

### T2 — Button Enhancement: Success Variant + Loading State

**Files**: `apps/web/src/components/ui/button.tsx`

**Depends on**: T1 (needs `--success`, `--primary-hover`, `--primary-active` tokens)

**What**:
1. Add `success` variant to `buttonVariants` CVA config
2. Update `default` variant to use `hover:bg-primary-hover active:bg-primary-active`
3. Add `loading` prop: when true, replaces children with `Loader2` spinner, adds `aria-busy="true"` + `pointer-events-none`
4. Add optional `loadingText` prop
5. Update `lg` size to `h-10 px-4` (spec says 40px, currently `h-9`)

**Interface change**:
```tsx
interface ButtonProps extends ButtonPrimitive.Props, VariantProps<typeof buttonVariants> {
  loading?: boolean
  loadingText?: string
}
```

**Acceptance criteria**:
- `<Button variant="success">Publish</Button>` renders green bg
- `<Button loading>Save</Button>` shows spinner, is not clickable, has `aria-busy`
- `<Button loading loadingText="Saving...">Save</Button>` shows spinner + text
- Existing button usages unchanged (variant defaults still work)
- `npm run build` passes

**Verify**: `cd apps/web && npm run build` + manual visual check in browser

---

### T3 — Status Badges: Draft / Published / Modified

**Files**: `apps/web/src/components/ui/badge.tsx`

**Depends on**: T1 (needs `--success`, `--warning`, `--primary-muted` tokens)

**What**: Add three semantic badge variants to `badgeVariants` CVA config:
- `draft`: `bg-warning/10 text-warning border-warning/20`
- `published`: `bg-success/10 text-success border-success/20`
- `modified`: `bg-primary-muted text-primary border-primary/20`

**Acceptance criteria**:
- `<Badge variant="draft">Draft</Badge>` renders amber-toned badge
- `<Badge variant="published">Published</Badge>` renders green-toned badge
- `<Badge variant="modified">Modified</Badge>` renders indigo-toned badge
- Existing badge variants unchanged
- `npm run build` passes

**Verify**: `cd apps/web && npm run build`

---

### T4 — Sidebar Context + localStorage Persistence

**Files**: New `apps/web/src/components/sidebar/SidebarContext.tsx`

**Depends on**: T1 (sidebar tokens must exist)

**What**: Create `SidebarContext` providing:
- `collapsed: boolean` — sidebar expanded/collapsed state
- `toggle: () => void` — toggle collapsed
- `isMobile: boolean` — true when viewport `<1024px`
- `mobileOpen: boolean` — overlay sidebar open on mobile
- `setMobileOpen: (open: boolean) => void`

Persist `collapsed` to `localStorage` key `"sidebar-collapsed"`. Use `matchMedia` listener for `isMobile`.

**Acceptance criteria**:
- `useSidebar()` returns correct `collapsed` state
- Toggling persists to localStorage
- Refreshing page restores collapsed state
- `isMobile` reacts to viewport resize
- `npm run build` passes

**Verify**: `cd apps/web && npm run build`

---

### T5 — Sidebar Components: Brand, Group, SubGroup, Item, Toggle

**Files**: New files in `apps/web/src/components/sidebar/`
- `Sidebar.tsx` — root `<aside>` with width transition
- `SidebarBrand.tsx` — logo/brand area, links to `/admin`
- `SidebarGroup.tsx` — collapsible section (Content Manager, Settings)
- `SidebarSubGroup.tsx` — labeled sub-section (Single Types, Collection Types)
- `SidebarItem.tsx` — individual `NavLink` with active state
- `SidebarCollapseToggle.tsx` — expand/collapse button at footer
- `index.ts` — barrel export

**Depends on**: T4 (needs `SidebarContext`)

**What**:
- `Sidebar`: dark bg (`bg-sidebar`), `w-64` expanded / `w-16` collapsed, `transition-[width] duration-200 ease-in-out`, flex column, border-r
- `SidebarBrand`: logo icon + "AbyssOfTime CMS" text (hidden when collapsed)
- `SidebarGroup`: Lucide icon + label + chevron toggle, children collapsible, expand/collapse persisted to localStorage per group key
- `SidebarSubGroup`: section label (e.g., "Single Types") + indented children
- `SidebarItem`: `NavLink` with active state detection, `text-sidebar-foreground` default, `bg-sidebar-accent text-sidebar-primary` when active with left border indicator
- `SidebarCollapseToggle`: button at sidebar footer, `ChevronsLeft`/`ChevronsRight` icon
- When collapsed: show icon-only items, group labels hidden, hover shows popover with sub-menu items

**Acceptance criteria**:
- Sidebar renders with Content Manager group (Single Types + Collection Types sub-groups) and Settings group
- Clicking group header expands/collapses children
- Active nav item has indigo left border + highlighted text
- Collapse toggle shrinks sidebar to 64px rail with icons only
- Hovering rail icons shows popover with sub-menu
- Role-gated items (Users, Access Tokens, Roles) excluded from DOM for insufficient roles
- `npm run build` passes

**Verify**: `cd apps/web && npm run build` + manual browser check

---

### T6 — Sidebar Responsive: Mobile Overlay

**Files**: `apps/web/src/components/sidebar/Sidebar.tsx`, `SidebarContext.tsx`

**Depends on**: T5

**What**: When `isMobile` is true:
- Sidebar rendered as fixed overlay with backdrop (`bg-black/50`)
- Opened via `setMobileOpen(true)` (triggered by hamburger in TopBar)
- Clicking backdrop or any nav item closes overlay
- Sidebar overlay uses `z-40`
- No rail mode on mobile — always full-width when open

**Acceptance criteria**:
- At `<1024px`, sidebar not visible by default
- Opening overlay shows full sidebar with backdrop
- Clicking nav item navigates AND closes overlay
- Clicking backdrop closes overlay
- At `≥1024px`, sidebar reverts to normal expand/collapse behavior

**Verify**: Browser resize to mobile width + manual interaction

---

### T7 — AdminLayout Integration: Wire New Sidebar

**Files**: `apps/web/src/pages/admin/layout/AdminLayout.tsx`

**Depends on**: T5 (Sidebar components), T4 (SidebarContext)

**What**:
- Replace current `<Sidebar />` import with new sidebar
- Wrap layout in `<SidebarProvider>`
- Main content area adjusts width based on sidebar collapsed state (`ml-64` vs `ml-16`)
- Remove old `apps/web/src/pages/admin/layout/Sidebar.tsx` (replaced entirely)

**Acceptance criteria**:
- Admin panel renders with new collapsible sidebar
- Content area fills remaining width
- All existing navigation routes still work
- Old Sidebar.tsx removed
- `npm run build` passes

**Verify**: `cd apps/web && npm run build` + navigate all routes in browser

---

### ✅ Checkpoint 1: Foundation + Sidebar Complete

Verify before proceeding:
- [ ] All new color tokens resolve in Tailwind (`npm run build`)
- [ ] Button success + loading states work
- [ ] Status badges render correctly
- [ ] Sidebar collapses/expands with persisted state
- [ ] Rail mode shows icons + popover sub-menus
- [ ] Mobile overlay works at `<1024px`
- [ ] All existing routes navigate correctly
- [ ] `cd apps/web && npm run test` — existing tests pass

---

### T8 — Breadcrumbs + TopBar Rebuild

**Files**:
- New `apps/web/src/hooks/useBreadcrumbs.ts`
- New `apps/web/src/components/ui/breadcrumb.tsx`
- Update `apps/web/src/pages/admin/layout/TopBar.tsx`

**Depends on**: T7 (AdminLayout must use new sidebar for hamburger toggle)

**What**:
1. `useBreadcrumbs()` hook: derive breadcrumb segments from `useMatches()` (react-router). Map route patterns to labels:
   - `/admin` → "Home"
   - `/admin/content-type/single-type/:slug` → "Content Manager" / slug name
   - `/admin/content-type/collection-type/:slug` → "Content Manager" / slug name
   - `/admin/settings/*` → "Settings" / page name
2. `Breadcrumb` component: renders segments with chevron separators, last item non-clickable
3. TopBar: hamburger button (mobile only, triggers `setMobileOpen(true)`), breadcrumbs left, user badge + logout right

**Acceptance criteria**:
- Breadcrumbs update on route change
- Last breadcrumb is plain text
- Hamburger only visible on mobile
- Hamburger opens sidebar overlay
- `npm run build` passes

**Verify**: Navigate multiple routes, check breadcrumbs update correctly

---

### T9 — Sticky Action Bar for Content Pages

**Files**:
- New `apps/web/src/pages/admin/layout/StickyActionBar.tsx`
- Update `apps/web/src/pages/admin/panels/content-type/ContentDetailLayout.tsx`
- Update `apps/web/src/components/content-type/ContentTypeLayout.tsx`

**Depends on**: T2 (Button loading/success), T8 (TopBar complete)

**What**:
1. `StickyActionBar`: `sticky top-0 z-30`, glassmorphism bg (`bg-background/80 backdrop-blur-sm`), `h-16 border-b`, renders title + status badge left, action buttons right
2. Update `ContentDetailLayout`: replace current header div with `StickyActionBar`, move back link above it
3. Update `ContentTypeLayout`: use same sticky pattern for consistency
4. Wire button states to form lifecycle:
   - Save: `variant="default"` when dirty, `variant="secondary"` + disabled when clean, loading spinner when submitting
   - Publish: `variant="success"`, disabled when dirty (must save first)

**Acceptance criteria**:
- Action bar sticks to top on scroll
- Save button disabled when form is clean
- Save button shows spinner during submission
- Publish button uses green success variant
- Back link visible on detail pages
- Status badge uses semantic colors (draft=amber, published=green)
- `npm run build` passes

**Verify**: Open content editing page, scroll, verify sticky behavior + button states

---

### T10 — Card Wrappers + Page Polish

**Files**:
- New `apps/web/src/components/ui/card.tsx`
- Update `apps/web/src/pages/admin/panels/content-type/ContentTypeBuilder.tsx`
- Update content editing pages to wrap form sections in cards

**Depends on**: T9

**What**:
1. `Card` component: `bg-card border border-border rounded-lg shadow-sm`, with `CardHeader` (`px-6 py-4 border-b`) and `CardContent` (`p-6`) sub-components
2. Wrap `ContentTypeBuilder` form fields in `Card` container
3. Apply consistent spacing: `gap-6` between page sections, `gap-4` between fields within cards, `px-6` page padding

**Acceptance criteria**:
- Form fields wrapped in white card containers with subtle shadow
- Consistent spacing between cards and fields
- Collection list pages use same card pattern for table container
- `npm run build` passes

**Verify**: Visual check all content pages

---

### T11 — Dark Mode Verification Pass

**Files**: Potentially `apps/web/src/index.css` and component files if fixes needed

**Depends on**: T10 (all components must exist)

**What**: Toggle dark mode and verify every new component:
- Color tokens render correct dark values
- Sidebar: dark bg in light mode, darker bg in dark mode, text contrast OK
- Buttons: all variants visible with sufficient contrast
- Badges: semantic colors readable in dark
- Cards: dark card surface distinct from dark background
- Sticky action bar: glassmorphism effect works in dark
- Breadcrumbs: text readable
- Focus rings visible on dark backgrounds

**Acceptance criteria**:
- No contrast issues (WCAG AA minimum: 4.5:1 for text)
- No invisible elements in dark mode
- All transitions smooth in both modes
- `npm run build` passes
- `npm run test` passes

**Verify**: Toggle dark mode, check every page and component

---

### ✅ Checkpoint 2 (Final): Full Design System Complete

- [ ] All admin pages use new design system
- [ ] Sidebar expand/collapse/rail/mobile all work
- [ ] TopBar breadcrumbs update on navigation
- [ ] Sticky action bar on content editing pages
- [ ] Button loading + success states functional
- [ ] Status badges use semantic colors
- [ ] Form sections in card containers
- [ ] Dark mode fully working
- [ ] `npm run build` — zero errors
- [ ] `npm run test` — all tests pass
- [ ] No hardcoded colors in component files
