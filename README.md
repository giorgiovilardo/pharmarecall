# PharmaRecall

Web application for Italian pharmacies to manage patients with recurring prescriptions. Tracks refill schedules, calculates daily orders, and helps pharmacies prepare packages proactively.

Pharmacy personnel log in, register patients, record their recurring prescriptions (medication, units per box, daily consumption), and the system calculates when each box will run out. A dashboard shows what needs to be prepared each day. Orders move through a simple lifecycle: pending → prepared → fulfilled.

## Tech stack

Go 1.26, PostgreSQL 18, server-rendered HTML with [Templ](https://templ.guide) templates and [oat.ink](https://oat.ink/) CSS (~8KB, semantic, zero-dependency). Single binary deployment with embedded migrations and static assets.

Key libraries: pgx/v5 (database driver + connection pool), sqlc (query codegen), alexedwards/scs with pgxstore (server-side sessions), goose (migrations), koanf (TOML config), bcrypt (password hashing).

All codegen tools (templ, sqlc, goose) are managed as [Go tool dependencies](https://go.dev/doc/modules/managing-dependencies#tools) — no global installs needed.

## Architecture

### Hexagonal (ports & adapters)

Business logic lives in domain packages with **zero infrastructure imports** — no `pgx`, no `net/http`, no `database/sql`. HTTP and database are adapters wired together in the composition root (`cmd/server/main.go`).

```
┌──────────────────────────────────────────────────────────┐
│                     HTTP (driving adapter)                │
│   web/handler/ — parse form → call service → render      │
│   web/middleware.go — auth, role guards, context          │
│   web/*.templ — server-rendered templates                 │
└────────────────────────┬─────────────────────────────────┘
                         │ calls public Service methods
┌────────────────────────▼─────────────────────────────────┐
│                   Domain services                         │
│   user/service.go      — authentication, password mgmt   │
│   pharmacy/service.go  — CRUD, personnel management      │
│   patient/service.go   — CRUD, consensus tracking        │
│   prescription/service.go — CRUD, depletion calc, refill │
│   order/service.go     — dashboard generation, lifecycle  │
│   notification/service.go — in-app alerts                │
└────────────────────────┬─────────────────────────────────┘
                         │ uses small port interfaces
┌────────────────────────▼─────────────────────────────────┐
│               PostgreSQL (driven adapter)                 │
│   */pgxrepo.go — implements ports via pgx + sqlc         │
│   internal/db/ — sqlc-generated query code               │
│   db/migrations/ — goose SQL migrations                  │
└──────────────────────────────────────────────────────────┘
```

**Driving adapters** (HTTP handlers) call public methods on domain services. **Driven adapters** (pgxrepo files) implement small port interfaces that domain services depend on. The composition root wires it all together.

### Key patterns

- **Closure-based handler DI**: each handler is a `func(...deps) http.HandlerFunc` closing over its dependencies. No server struct, no method receivers on a god object.
- **Small port interfaces**: each driven port has 1 method (e.g., `UserFinder`, `PrescriptionCreator`). They compose into a `Repository` interface for wiring convenience only.
- **Consumer-side interfaces**: interfaces are declared where they are consumed (Go idiom). Handler-side interfaces in `web/handler/`, driven port interfaces in each domain's `port.go`.
- **Domain errors as sentinels**: all domain errors are `var Err... = errors.New(...)` in the domain type file. Handlers use `errors.Is()` to map them to HTTP responses.
- **Domain types, not DB types**: handlers and templates use `patient.Patient`, `prescription.Prescription`, etc. The `db.*` types never leak outside `pgxrepo.go`.
- **On-demand order calculation**: no cron jobs. Orders are created lazily when the dashboard is loaded, based on the configurable lookahead window.

### Domain model

```
Pharmacy 1──* User (owner, personnel)
         1──* Patient
                1──* Prescription ──── depletion calculation
                       1──* Order (pending → prepared → fulfilled)
                       1──* RefillHistory
         1──* Notification
```

**Depletion formula**: `depletion_date = box_start_date + floor(units_per_box / daily_consumption)` days. Prescriptions are classified as "ok" (>7 days), "approaching" (≤7 days), or "depleted" (≤0 days).

**Order lifecycle**: when the dashboard is loaded, the system creates orders for prescriptions entering the lookahead window (default: 7 days). Each order is tied to a specific depletion cycle. Recording a refill starts a new cycle and auto-fulfills the previous order.

### Roles and access control

Three roles enforced by middleware:

| Role | Access | Landing page |
|------|--------|--------------|
| **admin** | Manage pharmacies and their personnel | `/admin` |
| **owner** | Manage own pharmacy's personnel + all staff features | `/dashboard` |
| **personnel** | Patients, prescriptions, orders, notifications | `/dashboard` |

All patient/prescription/order data is scoped to a pharmacy — queries always filter by `pharmacy_id`.

Middleware chain: CORS → sessions → LoadUser → LoadNotificationCount → router. Route-level guards (`RequireAuth`, `RequireAdmin`, `RequireOwner`, `RequirePharmacyStaff`) restrict access per role.

## Prerequisites

- Go 1.26+
- Docker (for PostgreSQL)
- [just](https://github.com/casey/just) (task runner)

## Local development

Start the database:

```
docker compose up pharmarecall-db -d
```

Copy the config template and adjust if needed:

```
cp config.toml.dist config.toml
```

Run migrations and seed the admin user:

```
just migrate up
just seed admin@example.com changeme
```

Build and run (migrations also run automatically on startup):

```
just build
./bin/pharmarecall
```

The server starts on `http://localhost:8080`.

## Running everything in Docker

```
docker compose up
```

This builds the app image (multi-stage, runs from `scratch`) and starts both the app and database. The app waits for the database to be healthy before starting. You need a `config.toml` in the project root — it gets mounted into the container.

Note: when running fully dockerized, the database host in `config.toml` should be `pharmarecall-db` (the compose service name) instead of `localhost`.

## Configuration

`config.toml` (see `config.toml.dist` for defaults):

```toml
[server]
port = 8080

[db]
url = "postgres://pharmarecall:pharmarecall@localhost:5432/pharmarecall?sslmode=disable"

[session]
secret = "dev-secret-change-in-production"

[lookahead]
days = 7    # how many days ahead to show approaching prescriptions
```

## Common commands

```
just build                    # build the binary (runs templ + sqlc generate first)
just build-prod               # production build (stripped, static, CGO_ENABLED=0)
just test                     # run all tests
just test ./internal/web/...  # run tests for a specific package
just test_races               # run all tests with race detection
just check                    # fmt + vet + fix + test (full quality gate)
just generate                 # regenerate templ and sqlc code
just migrate up               # apply pending migrations
just migrate down             # roll back one migration
just migrate status           # show migration status
just migrate_create <name>    # create a new migration file
just seed <email> <password>  # seed an admin user
```

## Project layout

```
cmd/
  server/                 entrypoint, composition root
  seed/                   admin user seeding

internal/
  auth/                   password hashing (bcrypt), session manager setup
  config/                 koanf TOML config loading
  db/                     sqlc-generated code (do not edit)
  dbutil/                 shared pgx type conversion helpers (Numeric↔float64, Time→Date)
  depletion/              pure functions for depletion calculations (shared across domains)

  user/                   DOMAIN — authentication, password management
    user.go                 types (User) + sentinel errors
    port.go                 driven port interfaces + Repository composite
    service.go              business logic (Authenticate, ChangePassword, SeedAdmin)
    pgxrepo.go              driven adapter (pgx/sqlc → domain types)

  pharmacy/               DOMAIN — pharmacy CRUD, personnel management
    pharmacy.go             types (Pharmacy, Summary, PersonnelMember, CreateParams)
    port.go                 driven port interfaces
    service.go              business logic (CreateWithOwner, List, Get, Update, personnel ops)
    pgxrepo.go              driven adapter

  patient/                DOMAIN — patient CRUD, consensus tracking
    patient.go              types (Patient, Summary, CreateParams, UpdateParams)
    port.go                 driven port interfaces
    service.go              business logic (Create, List, Get, Update, SetConsensus)
    pgxrepo.go              driven adapter

  prescription/           DOMAIN — prescription CRUD, depletion calculation, refills
    prescription.go         types + depletion formula (EstimatedDepletionDate, DaysRemaining, Status)
    port.go                 driven port interfaces
    service.go              business logic (Create, Get, Update, RecordRefill, ListByPatient)
    pgxrepo.go              driven adapter

  order/                  DOMAIN — order dashboard, status lifecycle
    order.go                types (Order, DashboardEntry) + depletion helpers
    port.go                 driven port interfaces
    service.go              business logic (GenerateOrders, GetDashboard, AdvanceStatus)
    pgxrepo.go              driven adapter

  notification/           DOMAIN — in-app notifications for approaching prescriptions
    notification.go         types (Notification) + depletion helpers
    port.go                 driven port interfaces
    service.go              business logic (Generate, List, MarkRead, MarkAllRead, CountUnread)
    pgxrepo.go              driven adapter

  web/                    DRIVING ADAPTER — HTTP layer
    handler/                thin handlers (parse form → call domain → render)
    middleware.go           LoadUser, RequireAuth, RequireAdmin, RequireOwner, RequirePharmacyStaff
    routes.go               NewRouter(Handlers struct) → *http.ServeMux
    *.templ                 Templ templates (accept domain types directly)

db/
  migrations/             SQL migration files (goose, sequential numbering, 9 migrations)
  queries/                SQL query files for sqlc codegen

static/                   static assets (oat.ink CSS, embedded via embed.FS)
```

## Database schema

9 migrations, applied sequentially:

1. **init** — extensions/baseline
2. **users** — email, password hash, name, role, pharmacy_id
3. **sessions** — scs session store (pgxstore)
4. **pharmacies** — name, address, phone, email
5. **patients** — first/last name, phone, email, delivery address, fulfillment (pickup/shipping), consensus, notes, pharmacy_id
6. **prescriptions** — medication name, units per box, daily consumption, box start date, patient_id
7. **refill_history** — previous box start/end dates, prescription_id
8. **orders** — status (pending/prepared/fulfilled), cycle start/depletion dates, prescription_id
9. **notifications** — pharmacy_id, prescription_id, transition type, read status

No PostgreSQL enums — constrained values use `text` columns with `CHECK` constraints.

## Routes

| Method | Path | Role | Description |
|--------|------|------|-------------|
| GET | `/` | public | Health check |
| GET/POST | `/login` | public | Login |
| POST | `/logout` | auth | Logout |
| GET/POST | `/change-password` | auth | Change own password |
| GET | `/dashboard` | staff | Order dashboard (generates orders on load) |
| GET | `/dashboard/print` | staff | Print-friendly order list |
| GET | `/dashboard/labels` | staff | Batch print labels |
| POST | `/orders/{id}/advance` | staff | Advance order status |
| GET | `/orders/{id}/label` | staff | Print single order label |
| GET | `/notifications` | staff | Notification list |
| POST | `/notifications/{id}/read` | staff | Mark notification as read |
| POST | `/notifications/read-all` | staff | Mark all notifications as read |
| GET | `/admin` | admin | Admin dashboard (pharmacy list) |
| GET/POST | `/admin/pharmacies/...` | admin | Pharmacy CRUD + personnel |
| GET/POST | `/personnel` | owner | Own pharmacy personnel management |
| GET/POST | `/patients` | staff | Patient CRUD |
| GET/POST | `/patients/{id}` | staff | Patient detail + update |
| POST | `/patients/{id}/consensus` | staff | Record patient consensus |
| GET/POST | `/patients/{id}/prescriptions/...` | staff | Prescription CRUD + refill |

## TODO

- Proper error pages
- Proper i18n/strings
