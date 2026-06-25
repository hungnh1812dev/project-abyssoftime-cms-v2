# SPEC: Collapsible Component Fields

**Status:** Draft
**Module:** Frontend (apps/web)
**References:** [rules/frontend.md](../rules/frontend.md), [rules/content.md](../rules/content.md)

---

## 1. Objective

Add expand/collapse functionality to all component fields (both non-repeatable and repeatable) in the document edit form. When collapsed, each component displays its name plus the current value of its first text-type field as a hint, reducing visual clutter for deeply nested schemas.

**Target users:** Content editors working with content types that contain multiple or deeply nested components.

**Problem:** All component fields are always fully expanded, making long forms with nested components difficult to scan and navigate. Editors must scroll past large sections of irrelevant fields to reach the one they want to edit.

---

## 2. Scope

### In Scope
- Expand/collapse toggle on non-repeatable component `<fieldset>` elements
- Expand/collapse toggle on each individual repeatable component entry (#1, #2, ...)
- Hint text (first text field value) displayed when collapsed
- Default expand/collapse state based on depth and component type
- Clickable header area with chevron icon indicator

### Out of Scope
- Persisting expand/collapse state across navigations (resets to defaults on every load)
- Collapse for primitive fields (text, number, boolean, media, richtext, json)
- Drag-and-drop reordering (existing rule: NEVER)
- Bulk expand/collapse all button (can be added later)
- Keyboard shortcuts for expand/collapse

---

## 3. Default Expand/Collapse Rules

| Component Type | Depth | Default State |
|---|---|---|
| Non-repeatable component | depth = 0 (top-level) | **Expanded** |
| Non-repeatable component | depth >= 1 (nested) | **Collapsed** |
| Repeatable entry item | Any depth | **Collapsed** |

**Depth definition:** `depth` is the parameter already passed through `renderField()` in `renderSchemaField.tsx`. Top-level fields rendered by `ContentTypeBuilder` start at `depth = 0`. Each nesting level increments by 1.

**Note on "level" terminology:** The user refers to "first level" as `depth = 0` and "level >= 2" as `depth >= 1`. This spec uses `depth` consistently to match the codebase.

---

## 4. UI Design

### 4.1 Collapsed State — Non-Repeatable Component

```
+---------------------------------------------------------------+
| > componentName — First text field value here...              |
+---------------------------------------------------------------+
```

- Chevron icon (`ChevronRight` from lucide-react) on the left, rotates 90 degrees when expanded
- Component name in bold/medium weight (existing legend style)
- Separator dash (`—`) between name and hint text
- Hint text: current value of the first `text`-type child field, truncated with ellipsis if too long
- If no text field exists or value is empty, show name only (no dash, no hint)
- Entire header row is clickable to toggle

### 4.2 Expanded State — Non-Repeatable Component

```
+---------------------------------------------------------------+
| v componentName — First text field value here...              |
|---------------------------------------------------------------|
| [child fields rendered in grid as today]                      |
+---------------------------------------------------------------+
```

- Chevron rotated down (`ChevronRight` with `rotate-90` transform)
- Header shows same hint text as collapsed state
- Child fields rendered below in the existing 6-column grid layout
- All existing depth-based styling (border color, background color) preserved

### 4.3 Collapsed State — Repeatable Entry Item

```
+---------------------------------------------------------------+
| > #1 — First text field value here...     [Up] [Down] [Delete]|
+---------------------------------------------------------------+
```

- Each entry item has its own expand/collapse toggle
- Index indicator (`#1`, `#2`) remains visible
- Hint text from the first text-type child field of that entry
- Move up/down and delete buttons remain visible and functional when collapsed
- The outer repeatable `<fieldset>` wrapper (with label + "Add entry" button) is always expanded

### 4.4 Expanded State — Repeatable Entry Item

```
+---------------------------------------------------------------+
| v #1 — First text field value here...     [Up] [Down] [Delete]|
|---------------------------------------------------------------|
| [child fields rendered in grid as today]                      |
+---------------------------------------------------------------+
```

- Same as current UI but with the chevron toggle added to the entry header

### 4.5 Chevron Animation

- CSS transition on transform: `transition-transform duration-200`
- Collapsed: `rotate-0` (pointing right)
- Expanded: `rotate-90` (pointing down)

---

## 5. Hint Text Rules

### 5.1 Source
- Scan the component's `fields` array (the schema definition, not form values)
- Find the **first** child field where `type === 'text'`
- Read the current form value for that field using `useFormContext().watch()`

### 5.2 Display
- Show raw text value (no HTML rendering)
- Truncate to max 60 characters with `...` suffix
- Apply muted text color: `text-muted-foreground`
- Separator: ` — ` (space-emdash-space) between name/index and hint

### 5.3 Edge Cases
- No text-type child field exists → show name/index only, no separator
- Text field exists but value is empty/undefined → show name/index only, no separator
- Value is whitespace-only → treat as empty
- For repeatable entries, read from `{name}.{index}.{firstTextField}`

---

## 6. Files to Modify

| File | Change |
|---|---|
| `apps/web/src/pages/admin/panels/content-type/renderSchemaField.tsx` | Replace non-repeatable component `<fieldset>` with collapsible version; pass `depth` for default state |
| `apps/web/src/components/form/inputs/RepeatableComponentField.tsx` | Add per-entry expand/collapse toggle with hint text; all entries collapsed by default |

### 6.1 No New Files Required

The collapsible behavior is inline local state (`useState`) within the existing components. No new component file, hook, or context is needed.

### 6.2 No New Dependencies

- `ChevronRight` icon: already available via `lucide-react` (same package used for `ArrowUp`, `ArrowDown`, `Trash2`, `Plus`)
- `useFormContext().watch()`: already available via `react-hook-form`

---

## 7. Implementation Details

### 7.1 Non-Repeatable Component (in `renderSchemaField.tsx`)

The current non-repeatable component block (lines 66-75) renders a `<fieldset>` with all children always visible. Replace with:

1. Local state: `useState(depth >= 1)` — collapsed if nested, expanded if top-level
2. Find first text field: `field.fields?.find(child => child.type === 'text')?.name`
3. Watch hint value: `useFormContext().watch(fieldName + '.' + firstTextFieldName)`
4. Render clickable header with chevron + name + hint
5. Conditionally render child grid based on expanded state

**Important:** Because `renderField` is currently a plain function (not a component), the non-repeatable component section must be extracted into a small React component (e.g., `CollapsibleFieldset`) to use hooks (`useState`, `useFormContext`). This component lives in the same file — no new file.

### 7.2 Repeatable Entry Item (in `RepeatableComponentField.tsx`)

The current entry map (lines 34-76) renders each item always expanded. Modify:

1. Extract entry rendering into a sub-component (e.g., `RepeatableEntry`) to use hooks
2. Local state per entry: `useState(true)` — always collapsed by default
3. Find first text field: `childFields.find(child => child.type === 'text')?.name`
4. Watch hint value: `watch(`${name}.${index}.${firstTextFieldName}`)`
5. Add chevron before `#N` index indicator
6. Conditionally render child grid based on expanded state
7. Keep move/delete buttons always visible regardless of collapse state

### 7.3 Hint Text Helper

A small pure function (in whichever file uses it, or shared if both need it):

```typescript
function getHintText(fields: FieldDefinition[]): string | undefined {
  const firstTextField = fields.find(field => field.type === 'text');
  return firstTextField?.name;
}

function formatHint(value: unknown): string {
  if (typeof value !== 'string') return '';
  const trimmed = value.trim();
  if (!trimmed) return '';
  return trimmed.length > 60 ? trimmed.slice(0, 60) + '...' : trimmed;
}
```

---

## 8. Accessibility

- Chevron toggle: `aria-expanded="true|false"` on the clickable element
- Clickable header: `role="button"`, `tabIndex={0}`, responds to Enter/Space keypress
- Child content wrapper: `aria-hidden` when collapsed (or use `hidden` attribute)
- Existing ARIA labels on move/delete buttons are preserved

---

## 9. Testing Strategy

### 9.1 Unit Tests (Vitest + React Testing Library)

**File:** `apps/web/src/pages/admin/panels/content-type/__tests__/renderSchemaField.test.tsx`

| Test Case | Description |
|---|---|
| Top-level component expanded by default | Depth 0 non-repeatable component renders children visible |
| Nested component collapsed by default | Depth >= 1 non-repeatable component hides children |
| Toggle expand/collapse | Click header toggles children visibility |
| Hint text displayed | First text field value shown in collapsed header |
| Empty hint text | No text field → no separator or hint shown |
| Chevron rotation | `aria-expanded` attribute toggles correctly |

**File:** `apps/web/src/components/form/inputs/__tests__/RepeatableComponentField.test.tsx`

| Test Case | Description |
|---|---|
| Entry collapsed by default | Each repeatable entry starts collapsed |
| Toggle entry expand/collapse | Click entry header toggles that entry only |
| Hint text on entry | First text field value shown in entry header |
| Controls visible when collapsed | Move/delete buttons remain accessible |
| Add entry | New entry appears collapsed |
| Independent toggle | Expanding one entry does not affect others |

---

## 10. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Reset expand/collapse state to defaults on form load |
| **Always** | Show move/delete buttons on repeatable entries even when collapsed |
| **Always** | Use `aria-expanded` on toggle elements |
| **Always** | Preserve existing depth-based color cycling |
| **Always** | Keep child grid layout (`grid-cols-6`) unchanged |
| **Never** | Persist collapse state to storage or server |
| **Never** | Collapse primitive fields (only component-type fields) |
| **Never** | Hide the "Add entry" button when repeatable entries are collapsed |
| **Never** | Use `React.Children.map` or recursive child scanning |
| **Never** | Add external state management (context, store) for collapse state |
