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

### 3. PostgreSQL with sqlc + pgx

**Choice**: Use [sqlc](https://sqlc.dev/) for type-safe query code generation, backed by `jackc/pgx/v5` with `pgxpool` for connection pooling.

**Rationale**: ~27 queries across 7 tables. sqlc generates type-safe Go code from SQL — no manual row scanning, column/type mismatches caught at build time. SQL stays explicit (you write it yourself, sqlc generates the Go glue). Queries live in `db/queries/*.sql`, schema comes from goose migrations in `db/migrations/`. Generated code uses a `Queries` struct that accepts either a pool or a transaction, fitting the "all writes in a transaction" convention naturally.

**Alternatives considered**:
- Raw pgx — explicit but repetitive; ~27 queries means a lot of manual `rows.Scan` boilerplate
- GORM — too heavy, hides query behavior, complicates debugging

### 4. Session management with SCS

**Choice**: Use [alexedwards/scs](https://github.com/alexedwards/scs) for session management, with the PostgreSQL store (`scs/pgxstore`) and secure HTTP-only cookies.

**Rationale**: SCS is a mature, well-tested session library for Go that provides middleware for `net/http`, automatic cookie management, and a pluggable store backend. Using `pgxstore` keeps sessions in PostgreSQL with no extra infrastructure. Saves us from writing session CRUD, expiry cleanup, and cookie handling ourselves.

**Alternatives considered**:
- Hand-rolled sessions — more code to write and maintain for no benefit
- JWT — suited for APIs, adds complexity for server-rendered apps (token refresh, storage)
- gorilla/sessions — less actively maintained, no built-in pgx store

### 5. CSRF protection with stdlib CrossOriginProtection

**Choice**: Use Go's built-in `http.CrossOriginProtection` middleware (Go 1.25+) combined with `SameSite=Lax` session cookies from SCS.

**Rationale**: `http.CrossOriginProtection` checks `Sec-Fetch-Site` and `Origin` headers to reject cross-origin non-safe requests (POST, PUT, etc.) with 403. No CSRF tokens needed in forms, no third-party dependency. Combined with SameSite cookies, this provides defense-in-depth. CORS is not needed since this is a same-origin server-rendered app.

**Alternatives considered**:
- justinas/nosurf — works but adds a dependency and requires embedding tokens in every form
- gorilla/csrf — less actively maintained, same token-embedding overhead

### 6. Three user roles with role-based access

**Choice**: Three roles — admin, pharmacy owner, pharmacy personnel. Admin is seeded at deployment. Admin creates pharmacies and their owners. Owners manage their personnel.

**Rationale**: Centralised deployment model where one admin onboards pharmacies. Keeps self-registration out of the MVP, reducing abuse surface. Role is stored as a column on the users table, with middleware enforcing access per route group.

### 7. Domain-driven project layout

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

### 8. Order lifecycle tied to depletion cycles

**Choice**: Each prescription generates **orders**, one per depletion cycle. An order represents a single box-to-refill period and has its own status lifecycle: pending → prepared → fulfilled. When an order is fulfilled and a refill is recorded, the prescription starts a new cycle which eventually generates a new order. Fulfilled orders remain as terminal history.

The `orders` table tracks: prescription_id, cycle_start_date (box start), estimated_depletion_date, status, created_at, updated_at. Orders are created on-demand when the dashboard detects a prescription entering the lookahead window and no active (non-fulfilled) order exists for the current cycle.

**Rationale**: Separating orders from prescriptions gives each depletion cycle a clean lifecycle. No stale statuses after refills. Fulfilled orders serve as history. The dashboard query is straightforward: show active orders (pending/prepared) and let personnel advance them. No background scheduler needed — orders are created lazily on dashboard load.

**Lookahead window**: configurable, default 7 days. A prescription generates an order when estimated depletion is within the window.

### 9. Database migrations with goose

**Choice**: Use `pressly/goose` for schema migrations.

**Rationale**: Supports both SQL and Go-code migrations. Migrations are embedded into the binary via `embed.FS`, keeping the single-binary deployment story clean. Sequential numbering is simple to reason about for a small team. Go-code migrations provide an escape hatch for data migrations with business logic if needed later.

**Alternatives considered**:
- golang-migrate — SQL-only, no native `embed.FS` support, less convenient for single-binary deploys

### 10. Configuration with koanf and TOML

**Choice**: Use [koanf](https://github.com/knadh/koanf) with a TOML config file for application configuration.

**Rationale**: koanf is lightweight and flexible. TOML is readable and well-suited for server config. Configuration values: database connection string, server port, session secret, lookahead window.

**Alternatives considered**:
- Environment variables only — harder to manage for multiple settings, no structured nesting
- Viper — heavier, more features than needed

## Risks / Trade-offs

- **Session storage in PostgreSQL** → May need migration to Redis if session volume grows. Mitigation: abstract session storage behind an interface.
- **No background scheduler** → Order calculation happens per-request, which could slow down with thousands of patients. Mitigation: add database indexes on relevant date fields; introduce materialized views or caching if needed.
- **No email/SMS notifications in MVP** → Personnel must check the dashboard. Mitigation: in-app notification list on the dashboard; external notifications are a planned post-MVP capability.
- **Single-tenant design** → Each pharmacy is isolated by data, but the application is shared. Mitigation: row-level filtering by pharmacy ID on all queries.
