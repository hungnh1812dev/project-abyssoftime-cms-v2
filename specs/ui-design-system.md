# SPEC — UI Design System (Strapi-Inspired)

A Strapi-inspired design system for AbyssOfTime CMS admin panel. Adopts Strapi's key layout patterns — collapsible sidebar with grouped navigation, sticky action bars, structured page headers — while maintaining the project's own brand identity on the Shadcn + Tailwind foundation.

---

## 1. Objective

Redesign the admin panel UI to match Strapi CMS's usability patterns:

- **Collapsible sidebar** with icon-only rail mode, grouped navigation sections (Content Manager, Content-Type Builder, Settings), and expandable sub-menus
- **Button component** with complete state coverage (default, hover, active, focus, disabled, loading) and Strapi-aligned variants
- **Sticky header/action bar** with breadcrumbs, page titles, and persistent Save/Publish controls
- **Color system** migrated to Strapi-inspired indigo primary with full light + dark mode tokens
- **Typography and spacing** standardized across the admin panel

### Target Users

Developers and content editors using the AbyssOfTime CMS admin panel.

### Non-Goals

- Dynamic form engine or drag-and-drop UI
- Component library published as a standalone package
- Pixel-perfect Strapi replication — we adopt patterns, not copy the product

---

## 2. Color Tokens

Replace the current neutral grayscale primary with a Strapi-inspired indigo-based system. All tokens are CSS custom properties consumed via Tailwind.

### Light Theme

```
--primary:              oklch(0.488 0.243 264.376)   /* Indigo 600 — primary actions */
--primary-foreground:   oklch(0.985 0 0)             /* White text on primary */
--primary-hover:        oklch(0.432 0.232 265)        /* Indigo 700 — hover state */
--primary-active:       oklch(0.380 0.220 266)        /* Indigo 800 — pressed state */
--primary-muted:        oklch(0.932 0.032 264)        /* Indigo 50 — subtle backgrounds */

--secondary:            oklch(0.97 0 0)               /* Gray 50 — secondary surfaces */
--secondary-foreground: oklch(0.205 0 0)              /* Dark text on secondary */

--success:              oklch(0.627 0.194 145)        /* Green 600 — published, active */
--success-foreground:   oklch(0.985 0 0)
--warning:              oklch(0.750 0.183 55)         /* Amber 500 — draft, modified */
--warning-foreground:   oklch(0.205 0 0)
--destructive:          oklch(0.577 0.245 27.325)     /* Red 500 — delete, errors */

--background:           oklch(0.968 0.001 247)        /* Gray 25 — page background */
--foreground:           oklch(0.205 0 0)              /* Near-black body text */
--muted:                oklch(0.97 0 0)
--muted-foreground:     oklch(0.556 0 0)

--border:               oklch(0.922 0 0)
--input:                oklch(0.922 0 0)
--ring:                 oklch(0.488 0.243 264.376)    /* Matches primary for focus rings */

--card:                 oklch(1 0 0)                  /* White card surface */
--card-foreground:      oklch(0.145 0 0)

--sidebar:              oklch(0.205 0 0)              /* Dark sidebar background */
--sidebar-foreground:   oklch(0.800 0 0)              /* Light gray sidebar text */
--sidebar-primary:      oklch(0.488 0.243 264.376)    /* Indigo active indicator */
--sidebar-primary-foreground: oklch(0.985 0 0)
--sidebar-accent:       oklch(0.269 0 0)              /* Subtle hover on sidebar items */
--sidebar-accent-foreground: oklch(0.985 0 0)
--sidebar-border:       oklch(0.320 0 0)              /* Subtle dividers in sidebar */
--sidebar-muted:        oklch(0.556 0 0)              /* Sidebar secondary text */
```

### Dark Theme

```
--primary:              oklch(0.623 0.214 259)        /* Indigo 400 — lighter primary in dark */
--primary-foreground:   oklch(0.145 0 0)
--primary-hover:        oklch(0.570 0.230 261)
--primary-active:       oklch(0.520 0.240 263)
--primary-muted:        oklch(0.250 0.050 264)

--secondary:            oklch(0.269 0 0)
--secondary-foreground: oklch(0.985 0 0)

--success:              oklch(0.720 0.170 145)
--success-foreground:   oklch(0.145 0 0)
--warning:              oklch(0.800 0.160 55)
--warning-foreground:   oklch(0.145 0 0)
--destructive:          oklch(0.704 0.191 22.216)

--background:           oklch(0.145 0 0)
--foreground:           oklch(0.985 0 0)
--muted:                oklch(0.269 0 0)
--muted-foreground:     oklch(0.708 0 0)

--border:               oklch(1 0 0 / 10%)
--input:                oklch(1 0 0 / 15%)
--ring:                 oklch(0.623 0.214 259)

--card:                 oklch(0.205 0 0)
--card-foreground:      oklch(0.985 0 0)

--sidebar:              oklch(0.145 0 0)
--sidebar-foreground:   oklch(0.800 0 0)
--sidebar-primary:      oklch(0.623 0.214 259)
--sidebar-primary-foreground: oklch(0.985 0 0)
--sidebar-accent:       oklch(0.220 0 0)
--sidebar-accent-foreground: oklch(0.985 0 0)
--sidebar-border:       oklch(1 0 0 / 10%)
--sidebar-muted:        oklch(0.556 0 0)
```

---

## 3. Typography

| Token | Value | Usage |
|-------|-------|-------|
| `--font-sans` | `'Geist Variable', sans-serif` | Body text (keep current) |
| `--font-heading` | `var(--font-sans)` | Headings (same family, weight varies) |
| Page title | `text-2xl font-bold` (24px/700) | Main page heading |
| Section title | `text-lg font-semibold` (18px/600) | Card/section headings |
| Body | `text-sm` (14px/400) | Default body text |
| Caption | `text-xs` (12px/400) | Labels, helper text, sidebar group titles |
| Sidebar nav item | `text-sm font-medium` (14px/500) | Navigation links |

---

## 4. Button Component

Extend the existing `button.tsx` (Shadcn + CVA) with Strapi-aligned state behavior and new variants.

### Variants

| Variant | Resting | Hover | Active/Pressed | Focus | Disabled |
|---------|---------|-------|----------------|-------|----------|
| **default** (primary) | `bg-primary text-primary-foreground` | `bg-primary-hover` | `bg-primary-active scale-[0.98]` | `ring-2 ring-primary/50 ring-offset-2` | `opacity-50 cursor-not-allowed` |
| **secondary** | `bg-secondary text-secondary-foreground` | `bg-secondary` mixed with 5% foreground | `bg-secondary` mixed with 10% foreground, `scale-[0.98]` | `ring-2 ring-primary/50 ring-offset-2` | `opacity-50 cursor-not-allowed` |
| **success** | `bg-success text-success-foreground` | Darken 10% | Darken 20%, `scale-[0.98]` | `ring-2 ring-success/50 ring-offset-2` | `opacity-50 cursor-not-allowed` |
| **danger** (destructive) | `bg-destructive/10 text-destructive` | `bg-destructive/20` | `bg-destructive/30 scale-[0.98]` | `ring-2 ring-destructive/50 ring-offset-2` | `opacity-50 cursor-not-allowed` |
| **ghost** | Transparent | `bg-muted` | `bg-muted` mixed with foreground, `scale-[0.98]` | `ring-2 ring-primary/50 ring-offset-2` | `opacity-50 cursor-not-allowed` |
| **outline** | `border-border bg-background` | `bg-muted` | `bg-muted scale-[0.98]` | `ring-2 ring-primary/50 ring-offset-2` | `opacity-50 cursor-not-allowed` |
| **link** | `text-primary underline-offset-4` | `underline` | `underline text-primary-active` | `ring-2 ring-primary/50` | `opacity-50 cursor-not-allowed` |

### Sizes

| Size | Height | Padding | Icon Size | Font |
|------|--------|---------|-----------|------|
| `xs` | 24px (`h-6`) | `px-2` | 12px | `text-xs` |
| `sm` | 28px (`h-7`) | `px-2.5` | 14px | `text-[0.8rem]` |
| `default` | 32px (`h-8`) | `px-3` | 16px | `text-sm` |
| `lg` | 40px (`h-10`) | `px-4` | 16px | `text-sm` |
| `icon` | 32×32px | `p-0` | 16px | — |
| `icon-xs` | 24×24px | `p-0` | 12px | — |
| `icon-sm` | 28×28px | `p-0` | 14px | — |
| `icon-lg` | 40×40px | `p-0` | 16px | — |

### Loading State

```tsx
<Button loading>
  {/* children replaced with spinner + optional loadingText */}
</Button>
```

Behavior:
- `loading` prop adds `aria-busy="true"` and `pointer-events-none`
- Replaces children with a `Loader2` spinner (lucide) + optional `loadingText`
- Spinner inherits text color, animates with `animate-spin`
- Button retains its width (no layout shift) via `min-w-[current-width]` or fixed width

### Icon Buttons

```tsx
<Button size="icon" variant="ghost" aria-label="Close">
  <X />
</Button>
```

- `aria-label` required when no visible text
- Tooltip wrapper recommended for icon-only buttons

---

## 5. Sidebar Navigation

Replace the current flat `<aside>` sidebar with a Strapi-inspired collapsible navigation.

### Structure

```
┌──────────────────────────┐
│  [Logo] AbyssOfTime CMS  │  ← Brand area (links to /admin)
│  ─────────────────────── │
│                          │
│  ▸ CONTENT MANAGER       │  ← Expandable section
│    ● Single Types        │
│      ○ Homepage          │    ← NavLink items
│      ○ About             │
│    ● Collection Types    │
│      ○ Articles          │
│      ○ Categories        │
│                          │
│  ▸ SETTINGS              │  ← Expandable section
│    ○ Media Library       │
│    ○ Users               │    ← Role-gated (admin+)
│    ○ Access Tokens       │    ← Role-gated (super_admin)
│    ○ Roles               │    ← Role-gated (super_admin)
│                          │
│  ─────────────────────── │
│  [Collapse ‹‹]           │  ← Toggle button at bottom
└──────────────────────────┘
```

### Collapsed (Rail) Mode

```
┌────┐
│ 🏠 │  ← Logo icon only
│ ── │
│ 📄 │  ← Content Manager icon → popover sub-menu on hover
│ ⚙️ │  ← Settings icon → popover sub-menu on hover
│    │
│ ── │
│ ›› │  ← Expand toggle
└────┘
```

### Dimensions

| State | Width | Transition |
|-------|-------|------------|
| Expanded | `256px` (`w-64`) | `transition-[width] duration-200 ease-in-out` |
| Collapsed (rail) | `64px` (`w-16`) | Same transition |

### Sidebar Item States

| State | Style |
|-------|-------|
| Default | `text-sidebar-foreground` on `bg-transparent` |
| Hover | `bg-sidebar-accent text-sidebar-accent-foreground` |
| Active (current page) | `bg-sidebar-accent text-sidebar-primary font-medium` with `2px` left border in `sidebar-primary` color |
| Expanded section header | `text-sidebar-foreground font-semibold`, chevron rotated 90° |
| Collapsed section header | `text-sidebar-muted`, chevron at 0° |
| Disabled / hidden | Not rendered (role-gated items are excluded, not disabled) |

### Section Groups

Each group is a `<Collapsible>` (or custom accordion) with:
- **Header**: Icon + label (expanded) or icon-only (collapsed) + chevron toggle
- **Content**: List of `NavLink` items, indented under header
- **Persistence**: Expand/collapse state persisted to `localStorage` per section key
- **Icons**: Use Lucide icons — `FileText` (Content Manager), `Settings` (Settings)

### Responsive Behavior

| Breakpoint | Behavior |
|------------|----------|
| `≥1024px` (lg) | Sidebar visible, user can toggle expand/collapse |
| `<1024px` | Sidebar hidden by default, hamburger button in TopBar opens as overlay with backdrop |

### Component API

```tsx
// SidebarContext — provides collapsed state to children
interface SidebarContextValue {
  collapsed: boolean
  toggle: () => void
  isMobile: boolean
}

// Sidebar (root)
<Sidebar>
  <SidebarBrand />
  <SidebarContent>
    <SidebarGroup icon={FileText} label="Content Manager" defaultOpen>
      <SidebarSubGroup label="Single Types">
        <SidebarItem to="/admin/..." icon={File}>Homepage</SidebarItem>
      </SidebarSubGroup>
      <SidebarSubGroup label="Collection Types">
        <SidebarItem to="/admin/..." icon={FileStack}>Articles</SidebarItem>
      </SidebarSubGroup>
    </SidebarGroup>
    <SidebarGroup icon={Settings} label="Settings">
      <SidebarItem to="/admin/settings/media">Media Library</SidebarItem>
    </SidebarGroup>
  </SidebarContent>
  <SidebarFooter>
    <SidebarCollapseToggle />
  </SidebarFooter>
</Sidebar>
```

---

## 6. TopBar / Header

Replace the current minimal TopBar with a structured header.

### Structure

```
┌──────────────────────────────────────────────────────────────┐
│  [Breadcrumbs: Home / Content Manager / Articles]            │
│                                        [User badge] [Logout] │
└──────────────────────────────────────────────────────────────┘
```

### Layout

- Fixed height: `h-14` (56px)
- `border-b border-border`
- `flex items-center justify-between px-6`
- Left: Hamburger (mobile only) + breadcrumbs
- Right: User role badge + logout button

### Breadcrumbs

- Auto-derived from current route via a `useBreadcrumbs()` hook
- Separator: `/` or chevron icon
- Last item is plain text (not a link)
- Truncate middle items on small screens with `...`

---

## 7. Page Header & Sticky Action Bar

### Page Header

Every content page has a consistent header area:

```
┌──────────────────────────────────────────────────────────────┐
│  ← Back                                                      │
│                                                              │
│  Article Title                              [Save] [Publish] │
│  Status: Draft                                               │
└──────────────────────────────────────────────────────────────┘
```

### Sticky Action Bar

For content editing pages, the action bar (Save/Publish/Delete buttons) sticks to the top when scrolling:

| Property | Value |
|----------|-------|
| Position | `sticky top-0 z-30` |
| Background | `bg-background/80 backdrop-blur-sm` (glassmorphism effect) |
| Border | `border-b border-border` |
| Height | `h-16` (64px) |
| Padding | `px-6` |

### Action Bar Buttons

| Button | Variant | Condition | Behavior |
|--------|---------|-----------|----------|
| **Save** | `default` (primary) | Enabled when `isDirty === true` | Saves draft, shows loading spinner |
| **Publish** | `success` | Enabled when document has draft content | Saves + publishes |
| **Unpublish** | `outline` | Shown when document is published | Sets to draft |
| **Delete** | `danger` (ghost) | Always visible on existing documents | Confirmation dialog first |
| **Back** | `ghost` with `ArrowLeft` icon | On detail pages | Navigate to list |

### Button State Mapping to Form Lifecycle

| Form State | Save Button | Publish Button |
|------------|-------------|----------------|
| Clean (no changes) | Disabled, `variant="secondary"` | Enabled if draft exists |
| Dirty (has changes) | Enabled, `variant="default"` (indigo) | Disabled (must save first) |
| Submitting | Loading spinner, disabled | Disabled |
| Error | Re-enabled, error toast shown | Re-enabled |

---

## 8. Status Badge

Content status badges use semantic colors:

| Status | Color | Background | Border |
|--------|-------|------------|--------|
| **Draft** | `text-warning` | `bg-warning/10` | `border-warning/20` |
| **Published** | `text-success` | `bg-success/10` | `border-success/20` |
| **Modified** | `text-primary` | `bg-primary-muted` | `border-primary/20` |

Badge component: Extend existing `badge.tsx` with these semantic variants.

---

## 9. Card / Panel Component

Content editing areas use card containers:

```
┌─────────────────────────────────────────┐
│  Section Title                          │  ← card header
│─────────────────────────────────────────│
│                                         │
│  [Form fields inside]                   │  ← card body
│                                         │
└─────────────────────────────────────────┘
```

| Property | Value |
|----------|-------|
| Background | `bg-card` (white in light, dark gray in dark) |
| Border | `border border-border` |
| Radius | `rounded-lg` (`var(--radius)`) |
| Shadow | `shadow-sm` (subtle, Strapi-style) |
| Header padding | `px-6 py-4 border-b border-border` |
| Body padding | `p-6` |

---

## 10. Spacing System

Follow a 4px base grid (consistent with Tailwind):

| Token | Value | Usage |
|-------|-------|-------|
| `gap-1` | 4px | Tight element spacing |
| `gap-2` | 8px | Between related items |
| `gap-3` | 12px | Between form fields |
| `gap-4` | 16px | Between sections within a card |
| `gap-6` | 24px | Page-level section spacing |
| `p-6` | 24px | Card body padding |
| `px-6` | 24px | Page horizontal padding |
| `py-3` | 12px | Sidebar item vertical padding |

---

## 11. Implementation Plan

### Phase 1: Color Tokens & Button Enhancement
1. Update `index.css` with new color token values (primary → indigo, add success/warning/primary-hover/primary-active)
2. Extend `button.tsx` with `success` variant, `loading` prop, and updated state styles
3. Update `badge.tsx` with semantic status variants (draft/published/modified)

### Phase 2: Sidebar Rebuild
4. Create `SidebarContext` with collapsed state + localStorage persistence
5. Build new `Sidebar` component tree (`SidebarBrand`, `SidebarGroup`, `SidebarSubGroup`, `SidebarItem`, `SidebarCollapseToggle`)
6. Integrate Lucide icons for each nav section
7. Add popover sub-menus for collapsed rail mode
8. Implement responsive overlay mode for mobile (`<1024px`)

### Phase 3: TopBar & Action Bar
9. Add `useBreadcrumbs()` hook (derives from react-router match)
10. Rebuild `TopBar` with breadcrumbs and hamburger toggle
11. Build `StickyActionBar` component for content editing pages
12. Update `ContentTypeLayout` and `ContentDetailLayout` to use sticky action bar

### Phase 4: Page-Level Polish
13. Wrap form sections in card components
14. Apply spacing system consistently across all admin pages
15. Verify dark mode rendering for all new tokens and components

---

## 12. File Structure

```
apps/web/src/
├── components/
│   ├── ui/
│   │   ├── button.tsx           ← Extend with success, loading
│   │   ├── badge.tsx            ← Extend with status variants
│   │   ├── card.tsx             ← New: Card/Panel wrapper
│   │   └── breadcrumb.tsx       ← New: Breadcrumb component
│   └── sidebar/
│       ├── SidebarContext.tsx    ← Collapsed state + provider
│       ├── Sidebar.tsx          ← Root sidebar component
│       ├── SidebarBrand.tsx     ← Logo/brand area
│       ├── SidebarGroup.tsx     ← Expandable section
│       ├── SidebarSubGroup.tsx  ← Sub-section (Single Types, etc.)
│       ├── SidebarItem.tsx      ← Individual nav link
│       └── SidebarCollapseToggle.tsx
├── pages/admin/layout/
│   ├── AdminLayout.tsx          ← Update to use new sidebar
│   ├── TopBar.tsx               ← Rebuild with breadcrumbs
│   └── StickyActionBar.tsx      ← New: sticky action bar
├── hooks/
│   └── useBreadcrumbs.ts        ← New: route-derived breadcrumbs
└── index.css                    ← Update color tokens
```

---

## 13. Boundaries

### Always
- Use CSS custom properties for all colors — never hardcode hex/oklch in component files
- Button `loading` state must prevent double-submission (`pointer-events-none` + `aria-busy`)
- Sidebar collapse state persisted to `localStorage`
- All interactive elements must have visible focus indicators (ring)
- Role-gated sidebar items: exclude from DOM, don't render as disabled
- Transitions: use `duration-200 ease-in-out` for sidebar and state changes

### Ask Before
- Adding new color tokens beyond those specified here
- Changing button sizes or adding new size variants
- Modifying the sidebar section structure (adding/removing top-level groups)

### Never
- Hardcode colors in components — always reference CSS tokens via Tailwind
- Use CSS-in-JS or styled-components — Tailwind + CVA only
- Add animations longer than 300ms for UI state changes
- Remove existing Shadcn component patterns — extend them
- Use `z-index` above 50 — reserve `z-30` for sticky bar, `z-40` for sidebar overlay
