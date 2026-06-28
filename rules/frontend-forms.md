# RULES — Frontend Forms

**Scope:** FormProvider, field rendering, dot-notation deserialization, form state management.

---

## 1. FormProvider Rules

### 1.1 Lifecycle
| Moment | Behavior |
|---|---|
| Initial load | Fields rendered, pre-filled from server data |
| Clean state | Save button disabled (`isDirty === false`) |
| After edit | Save button enabled (`isDirty === true`) |
| Failed save | `toast.error(msg)`, form stays edited |
| Successful save | `toast.success('Saved')` → invalidate → reset → Save disabled |

### 1.2 FormProvider Props
- `query`: TanStack Query config for initial data fetch
- `mutationFn`: function to save data

### 1.3 Field Name Rules
- Flat name (`siteName`) → `{ siteName: "..." }`
- Dot-notation (`seo.title`) → `{ seo: { title: "..." } }`
- Array indexing (`skills.0.name`) → `{ skills: [{ name: "..." }] }`
- `FormProvider` handles conversion automatically

### 1.4 Invariants
- **NEVER** use `React.Children.map` or recursive child scanning
- **NEVER** use drag-and-drop or dynamic form engine
- FormProvider manages loading, submitting, isDirty — **NEVER** duplicate this state

---

## 2. Boundaries Summary

| Rule | Detail |
|---|---|
| **Always** | FormProvider manages loading/submitting/isDirty |
| **Never** | Use `React.Children.map` in FormProvider |
| **Never** | Use drag-and-drop or dynamic form engine |
