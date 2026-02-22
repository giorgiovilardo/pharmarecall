## Context

PharmaRecall is a greenfield Go web application for Italian pharmacies to manage recurring prescriptions. The system tracks patients, their recurring prescriptions (with consumption rates and box quantities), and calculates daily orders so pharmacies can prepare packages proactively.

The tech stack is Go 1.26 + PostgreSQL 18. The application is server-rendered — no SPA, no JS framework. Pharmacy personnel are the only users; patients are a domain concept, not application users.

## Goals / Non-Goals

**Goals:**

- Simple, server-rendered web app that pharmacy personnel can use immediately
- Accurate daily order calculation based on prescription consumption rates
- Single binary deployment (all assets and migrations embedded) with dockerized PostgreSQL for development
- Clean separation between domain logic and HTTP/persistence layers

**Non-Goals:**

- Patient-facing UI or patient authentication
- Payment processing
- ERP/CMS integration
- Order delivery logistics
- Real-time push notifications (email/SMS is future work; MVP uses in-app notifications)
- Mobile-specific UI (responsive CSS is sufficient)

## Decisions

### 1. Server-rendered HTML with Templ + oat.ink

**Choice**: Use [Templ](https://templ.guide) for type-safe Go HTML templates and [oat.ink](https://oat.ink/) as the CSS component library.

**Rationale**: Templ provides compile-time type safety for templates and integrates naturally with Go. oat.ink is an ~8KB zero-dependency semantic CSS library — it styles native HTML elements directly, which pairs well with server-rendered markup. No JS build pipeline needed. Static assets (oat.ink CSS) are embedded into the binary via `embed.FS`.

**Alternatives considered**:
- `html/template` — no type safety, error-prone string-based templates
- HTMX + Templ — adds interactivity complexity not needed for MVP

### 2. Standard library HTTP router (net/http)

**Choice**: Use Go's standard `net/http` with the ServeMux pattern matching (Go 1.22+).

**Rationale**: Go 1.22+ ServeMux supports method-based routing and path parameters natively. No need for a third-party router for this application's complexity.

**Alternatives considered**:
- Chi — good, but unnecessary given modern stdlib capabilities
- Gin — too opinionated, adds framework weight

### 3. PostgreSQL with raw SQL via pgx

**Choice**: Use `jackc/pgx` for PostgreSQL access with hand-written SQL queries.

**Rationale**: The domain model is straightforward (5-6 tables). Raw SQL keeps queries explicit and avoids ORM mapping complexity. pgx is the most performant and feature-complete Go PostgreSQL driver.

**Alternatives considered**:
- sqlc — good for code generation from SQL, but adds a build step; can adopt later
- GORM — too heavy, hides query behavior, complicates debugging

### 4. Session management with SCS

**Choice**: Use [alexedwards/scs](https://github.com/alexedwards/scs) for session management, with the PostgreSQL store (`scs/pgxstore`) and secure HTTP-only cookies.

**Rationale**: SCS is a mature, well-tested session library for Go that provides middleware for `net/http`, automatic cookie management, and a pluggable store backend. Using `pgxstore` keeps sessions in PostgreSQL with no extra infrastructure. Saves us from writing session CRUD, expiry cleanup, and cookie handling ourselves.

**Alternatives considered**:
- Hand-rolled sessions — more code to write and maintain for no benefit
- JWT — suited for APIs, adds complexity for server-rendered apps (token refresh, storage)
- gorilla/sessions — less actively maintained, no built-in pgx store

### 5. Three user roles with role-based access

**Choice**: Three roles — admin, pharmacy owner, pharmacy personnel. Admin is seeded at deployment. Admin creates pharmacies and their owners. Owners manage their personnel.

**Rationale**: Centralised deployment model where one admin onboards pharmacies. Keeps self-registration out of the MVP, reducing abuse surface. Role is stored as a column on the users table, with middleware enforcing access per route group.

### 6. Domain-driven project layout

**Choice**: Organize code by domain concern, not by technical layer.

```
cmd/server/          — entrypoint
internal/
  pharmacy/          — pharmacy + personnel domain
  patient/           — patient domain
  prescription/      — recurring prescription + order calculation
  auth/              — authentication, sessions
  web/               — HTTP handlers, middleware, Templ components
db/
  migrations/        — SQL migration files
```

**Rationale**: Keeps related logic together. The `internal/` package prevents external imports. Domain packages own their types, repository interfaces, and business logic.

### 7. Order calculation approach

**Choice**: Calculate orders on-demand when the dashboard is loaded, based on prescription data (consumption rate, box quantity, last refill date).

**Rationale**: No need for a background scheduler in the MVP. The formula is: estimate when the patient will run out based on daily consumption, and flag prescriptions that need a refill within a configurable lookahead window (e.g., 7 days). This keeps the system stateless and avoids cron complexity.

### 8. Database migrations with goose

**Choice**: Use `pressly/goose` for schema migrations.

**Rationale**: Supports both SQL and Go-code migrations. Migrations are embedded into the binary via `embed.FS`, keeping the single-binary deployment story clean. Sequential numbering is simple to reason about for a small team. Go-code migrations provide an escape hatch for data migrations with business logic if needed later.

**Alternatives considered**:
- golang-migrate — SQL-only, no native `embed.FS` support, less convenient for single-binary deploys

## Risks / Trade-offs

- **Session storage in PostgreSQL** → May need migration to Redis if session volume grows. Mitigation: abstract session storage behind an interface.
- **No background scheduler** → Order calculation happens per-request, which could slow down with thousands of patients. Mitigation: add database indexes on relevant date fields; introduce materialized views or caching if needed.
- **No email/SMS notifications in MVP** → Personnel must check the dashboard. Mitigation: in-app notification list on the dashboard; external notifications are a planned post-MVP capability.
- **Single-tenant design** → Each pharmacy is isolated by data, but the application is shared. Mitigation: row-level filtering by pharmacy ID on all queries.
