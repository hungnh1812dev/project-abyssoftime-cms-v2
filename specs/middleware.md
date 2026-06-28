# SPEC — Middleware Module

## 1. Overview

The middleware module provides cross-cutting HTTP middleware applied globally to all Gin routes. It handles CORS origin whitelisting, request body size limiting, and security response headers. These middleware functions are configured at the router level and applied before any handler logic executes.

---

## 2. File Map

All paths relative to `apps/api/`.

```
internal/delivery/http/middleware/cors.go             # CORS middleware
internal/delivery/http/middleware/cors_test.go
internal/delivery/http/middleware/bodylimit.go        # Request body size limit
internal/delivery/http/middleware/bodylimit_test.go
internal/delivery/http/middleware/security_headers.go  # Security response headers
internal/delivery/http/middleware/security_headers_test.go
```

---

## 3. Middleware

### CORS (`middleware/cors.go`)
- Checks `Origin` header against `CORS_ORIGINS` whitelist
- Sets `Access-Control-Allow-Origin`, `Allow-Methods`, `Allow-Headers`, `Allow-Credentials`
- Handles `OPTIONS` preflight with 204
- Never uses `Access-Control-Allow-Origin: *`

### Body Size Limit (`middleware/bodylimit.go`)
- Applied globally with `BODY_LIMIT_BYTES` default (10 MB)
- Media upload exempt (uses its own multipart limit)
- Uses `http.MaxBytesReader`

### Security Headers (`middleware/security_headers.go`)
Sets on every response:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 0`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Content-Security-Policy: default-src 'self'`

---

## 4. Boundaries

| Rule | Detail |
|---|---|
| **Always** | Apply body size limit globally; media upload uses its own multipart limit |
| **Always** | Set security headers on every response |
| **Never** | Use `Access-Control-Allow-Origin: *` — always explicit origin whitelist |

---

## 5. Changelog

| Date | Change | Source |
|------|--------|--------|
| v1.0 | Initial project setup with middleware | §1–§6 |
| v1.1 | Gin migration (Phase A) | §11.4 |
