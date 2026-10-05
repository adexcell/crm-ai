# Production Standards & Engineering Guidelines

## Core Principles
1. **Zero Cut Corners**: Every line of code written must be production-ready. Proper error handling, validation, transactions, connection pooling, graceful shutdowns, logging, and security headers are mandatory from day one.
2. **Dual Backend Parity**:
   - Both `backend-python` (FastAPI) and `backend-go` (Go) serve the exact same business logic, database tables, and API contracts.
   - Any API endpoint implemented in FastAPI must have an exact twin in Go with matching request/response schemas and HTTP status codes.
3. **Multi-Tenancy Isolation**:
   - Organization ID (`org_id` / `organization_id`) scoping is strictly enforced on all queries, events, and cache keys.
   - Cross-tenant data leakage is a critical blocker.
4. **Unified API Envelope**:
   - Success: `{"ok": true, "data": <payload>, "error": null}`
   - Failure: `{"ok": false, "data": null, "error": {"code": string, "message": string, "details": object}}`
5. **Frontend Design Excellence (Vue 3)**:
   - Modern, high-end aesthetics (slate/dark palette, vibrant refined accents, micro-animations, glassmorphism touches, clean typography).
   - Component-Driven Development: All UI primitives and reusable widgets must be isolated and cataloged in Storybook.
