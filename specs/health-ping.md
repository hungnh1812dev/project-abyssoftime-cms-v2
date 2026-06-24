# Spec: Background Health Ping Service

## 1. Objective

Keep the CMS frontend aware of API availability by continuously pinging the `/health` endpoint from the browser. When the API is unreachable, block the UI with a full-screen overlay informing the user that the CMS is connecting to the service. When the API recovers, automatically dismiss the overlay so the user can continue where they left off.

**Secondary goal:** On Render.com free-tier, the API spins down after 15 minutes of inactivity. The 14-minute ping interval keeps the service warm and prevents cold-start delays for active users.

### Target users

All CMS users — the service runs on every page (login, register, admin) as soon as the app mounts.

### Non-goals

- Server-side health checks (this is browser-only)
- Health endpoint changes on the backend (the existing `GET /health` → `{ "status": "ok" }` is sufficient)
- Retry with exponential backoff (fixed intervals are intentional for keep-alive)

---

## 2. Behavior

### 2.1 Ping Lifecycle

```
App mounts
  → Immediate ping to GET /health
  → If SUCCESS: schedule next ping in 14 minutes
  → If FAILURE: show overlay, schedule retry in 10 seconds
  → On recovery (failure → success): hide overlay, resume 14-minute cycle
  → On tab hidden (document.visibilityState === 'hidden'): pause ping timer
  → On tab visible again: immediate ping, then resume normal cycle
```

### 2.2 Ping Intervals

| State | Interval | Rationale |
|---|---|---|
| Healthy | 14 minutes | Keep Render.com free-tier alive (15 min timeout) |
| Unhealthy | 10 seconds | Fast recovery detection without hammering the server |

### 2.3 Ping Request

- **Method:** `GET`
- **URL:** `${VITE_API_URL}/health`
- **Timeout:** 5 seconds (do not wait for the default axios timeout)
- **Auth:** None (the `/health` endpoint is public, no Bearer token)
- **Important:** Use a standalone `fetch` or dedicated axios instance — do NOT use the main `api` axios instance from `lib/api.ts` to avoid triggering the 401-refresh interceptor on health checks

### 2.4 Success Criteria

- HTTP 200 response received within 5 seconds
- Any other status code or network error = failure

### 2.5 Visibility API Integration

- When the browser tab becomes hidden (`document.visibilitychange` → `hidden`): clear the ping timer to avoid unnecessary background requests
- When the tab becomes visible again: fire an immediate ping, then schedule the next one based on result
- This prevents wasted requests when the user is not actively using the CMS

---

## 3. UI — Connection Overlay

### 3.1 Overlay Behavior

| Trigger | Action |
|---|---|
| First ping fails (on mount or any subsequent) | Show full-screen overlay |
| Ping succeeds after failure | Auto-dismiss overlay (no reload) |
| Ping succeeds (healthy state) | Overlay stays hidden |

### 3.2 Overlay Design

- **Position:** Fixed, full viewport, `z-50` (above all content)
- **Background:** Semi-transparent backdrop (`bg-background/80 backdrop-blur-sm`)
- **Content (centered):**
  - Animated spinner (Tailwind `animate-spin`)
  - Text: **"Connecting to service..."**
  - Subtext: *"The server may be starting up. This can take up to 30 seconds."*
- **Interaction:** Blocks all clicks (overlay covers everything)
- **Accessibility:** `role="alert"`, `aria-live="assertive"`, `aria-busy="true"`

### 3.3 Component Location

- `apps/web/src/components/ConnectionOverlay.tsx`
- Named export: `ConnectionOverlay`
- Receives `visible: boolean` prop
- Uses Tailwind transitions for fade in/out

---

## 4. Architecture

### 4.1 New Files

| File | Purpose |
|---|---|
| `src/context/HealthContext.tsx` | React context + provider: manages ping loop, exposes `isApiHealthy` state |
| `src/components/ConnectionOverlay.tsx` | Full-screen blocking overlay UI |

### 4.2 Modified Files

| File | Change |
|---|---|
| `src/main.tsx` | Wrap app with `HealthProvider`, render `ConnectionOverlay` |

### 4.3 Context API

```typescript
interface HealthContextValue {
  isApiHealthy: boolean;
}
```

- `HealthProvider` wraps the entire app tree (outermost provider after `QueryClientProvider`)
- Exposes `useHealthStatus()` hook for any component that needs to read API health
- The provider itself renders `ConnectionOverlay` internally based on state

### 4.4 Provider Internals

```
HealthProvider
  ├── state: isApiHealthy (boolean, initial: true — optimistic)
  ├── ref: timerRef (stores setTimeout ID)
  ├── effect: on mount → ping immediately
  ├── effect: on visibilitychange → pause/resume
  ├── cleanup: on unmount → clear timer, remove listener
  └── renders: children + <ConnectionOverlay visible={!isApiHealthy} />
```

**Initial state is `true` (optimistic):** The overlay only appears after a confirmed failure. This prevents a flash of the overlay on fast connections.

### 4.5 Ping Function (not a TanStack Query)

This is intentionally NOT a TanStack Query. Rationale:
- The ping is a side-effect loop (fire-and-forget with scheduling), not request-response data fetching
- It has custom retry timing (10s fail / 14m success) that doesn't map to TanStack Query's retry config
- It should not populate the query cache or trigger re-renders across the app
- It uses `fetch` or a standalone axios instance to avoid the auth interceptor

---

## 5. Project Structure (affected)

```
apps/web/src/
├── components/
│   └── ConnectionOverlay.tsx          # NEW — overlay UI
├── context/
│   ├── AuthContext.tsx                 # existing
│   └── HealthContext.tsx               # NEW — ping loop + context
└── main.tsx                           # MODIFIED — add HealthProvider
```

---

## 6. Code Style

- TypeScript strict mode, no `any`
- Named exports only (no `export default`)
- All variable/function names 3+ characters
- Shadcn UI + TailwindCSS for overlay styling
- No comments unless the "why" is non-obvious

---

## 7. Testing Strategy

### 7.1 HealthContext Tests (`src/context/__tests__/HealthContext.test.tsx`)

| Test Case | Description |
|---|---|
| Initial healthy state | Provider renders children without overlay |
| Ping failure shows overlay | Mock fetch to reject → overlay becomes visible |
| Recovery hides overlay | Mock fetch to fail then succeed → overlay appears then disappears |
| Retry interval on failure | After failure, next ping fires in ~10s (use fake timers) |
| Success interval | After success, next ping fires in ~14m (use fake timers) |
| Cleanup on unmount | Timer is cleared, no lingering intervals |
| Visibility pause | Simulate `visibilitychange` to hidden → timer cleared |
| Visibility resume | Simulate `visibilitychange` to visible → immediate ping fires |

### 7.2 ConnectionOverlay Tests (`src/components/__tests__/ConnectionOverlay.test.tsx`)

| Test Case | Description |
|---|---|
| Visible when `visible=true` | Overlay renders with spinner and text |
| Hidden when `visible=false` | Overlay not in DOM or has hidden attribute |
| Accessibility attributes | Has `role="alert"`, `aria-live`, `aria-busy` |
| Blocks interaction | Overlay has pointer-events-auto on the backdrop |

### 7.3 Test Tools

- Vitest + React Testing Library
- `vi.useFakeTimers()` for interval testing
- `vi.stubGlobal('fetch', ...)` for mocking health endpoint
- No MSW needed (direct fetch mock is simpler for a single endpoint)

---

## 8. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Use standalone `fetch` or dedicated axios instance for `/health` — never the main `api` client |
| **Always** | Clear timer on unmount and on tab hidden |
| **Always** | Fire immediate ping on mount and on tab-visible |
| **Always** | Use `document.visibilityState` API to pause/resume |
| **Always** | Initial state is `true` (optimistic) — overlay only on confirmed failure |
| **Never** | Use TanStack Query for the ping loop (it's a side-effect, not data fetching) |
| **Never** | Send auth headers on health ping (endpoint is public) |
| **Never** | Ping when tab is hidden (waste of resources) |
| **Never** | Reload the page on recovery — auto-dismiss the overlay |
| **Never** | Run this service on the server / SSR — browser-only |
| **Ask first** | Changing ping intervals from 10s/14m |
| **Ask first** | Adding more health-check endpoints beyond `/health` |

---

## 9. Acceptance Criteria

1. On page load, CMS pings `GET /health` immediately
2. If healthy, pings again every 14 minutes
3. If unhealthy, pings again every 10 seconds
4. On failure, a full-screen blocking overlay appears with "Connecting to service..." message
5. On recovery, overlay auto-dismisses without page reload
6. Ping pauses when browser tab is hidden, resumes with immediate ping when tab is visible
7. Health check does not use the main `api` axios instance (no auth interceptor)
8. All tests pass (`make test-web`)
9. No TypeScript errors (`tsc --noEmit`)
