# PharmaRecall

## Project Overview

Web application for Italian pharmacies to manage patients with recurring prescriptions. Tracks refill schedules, calculates daily orders, and helps pharmacies prepare packages proactively.

## Tech Stack

- **Language**: Go 1.26
- **Database**: PostgreSQL 18 (dockerized for dev)
- **Templates**: [Templ](https://templ.guide) — type-safe Go HTML templates
- **CSS**: [oat.ink](https://oat.ink/) — semantic, zero-dependency CSS library (~8KB)
- **HTTP**: Go stdlib `net/http` with ServeMux (Go 1.22+)
- **DB driver**: jackc/pgx/v5 with pgxpool
- **Queries**: sqlc — SQL in `db/queries/*.sql`, generated Go code in `internal/db/`
- **Sessions**: alexedwards/scs with pgxstore
- **Migrations**: pressly/goose with embed.FS
- **Auth**: bcrypt password hashing, session-based with HTTP-only cookies
- **Config**: koanf with TOML config file

## Architecture

**Hexagonal (Ports & Adapters)** — business logic is portable and decoupled from HTTP.

- Server-rendered, no SPA, no JS framework
- Single binary deployment — all assets (oat.ink CSS) and migrations embedded via `embed.FS`
- Three roles: admin, pharmacy owner, pharmacy personnel

### Package Layout

```
cmd/server/main.go          — composition root: repos → services → handlers
cmd/seed/main.go             — uses user.Service.SeedAdmin

internal/
  auth/                      — password hashing (bcrypt) + session manager setup
  config/                    — koanf TOML config
  db/                        — sqlc generated (DO NOT EDIT)

  user/                      — DOMAIN: user/authentication business logic
    user.go                  — domain types (User) + domain errors
    port.go                  — driven ports: small single-purpose interfaces + Repository composite
    service.go               — Service (Authenticate, ChangePassword, SeedAdmin)
    service_test.go          — unit tests with tiny manual mocks per port
    pgxrepo.go               — driven adapter: PgxRepository implements ports via pgx/sqlc

  pharmacy/                  — DOMAIN: pharmacy business logic
    pharmacy.go              — domain types + domain errors
    port.go                  — driven ports: small single-purpose interfaces + Repository composite
    service.go               — Service (CreateWithOwner, List, Get, Update, ListPersonnel, CreatePersonnel)
    service_test.go          — unit tests with tiny manual mocks per port
    pgxrepo.go               — driven adapter: PgxRepository implements ports via pgx/sqlc

  web/                       — DRIVING ADAPTER: HTTP → domain
    handler/                 — thin HTTP handlers (parse form → call domain → render)
      login.go, logout.go, change_password.go, admin_dashboard.go, pharmacy.go, personnel.go
    middleware.go            — LoadUser, RequireAuth, RequireAdmin + context helpers
    routes.go                — NewRouter(Handlers struct) → *http.ServeMux
    *.templ                  — templates (accept domain types directly)
```

### Hexagonal Mapping

- **Driving ports**: public methods on `Service` (what handlers call)
- **Driven ports**: small interfaces in `port.go` (what services need from persistence)
- **Driving adapters**: `web/handler/` (HTTP → domain)
- **Driven adapters**: `pgxrepo.go` files (domain → PostgreSQL via pgx/sqlc)

### Key Patterns

- **Closure-based handler DI**: each handler is a `func(...deps) http.HandlerFunc` closing over its deps. No server struct.
- **Small port interfaces**: each port has 1 method, composed into `Repository` for wiring only
- **`NewServiceWith(ServiceDeps{...})`**: test constructor — inject only what you need, rest stays nil
- **`NewService(repo, ...)`**: production constructor — Repository satisfies all ports
- **Domain types, not db types**: handlers and templates use `pharmacy.Pharmacy`, `user.User`, etc. The `db.*` types never leak outside `pgxrepo.go`.
- **`service.go` has zero infrastructure imports**: no pgx, no db, no http
- **Handler-side interfaces**: handlers define consumer-side interfaces (Go idiom) that domain services satisfy
- `main.go` is the composition root: creates repos → services → handlers → router
- `NewRouter()` in `internal/web/routes.go` takes a `Handlers` struct of `http.HandlerFunc` values, knows nothing about interfaces or deps
- Handler tests: construct handlers directly with stubs, wrap in minimal middleware, test via `httptest.NewServer`
- Domain service tests: mock individual port interfaces, not the full Repository

### Where New Features Go

1. **Domain types/errors** → `pharmacy.go` or `user.go` in the domain package
2. **Business logic** → `service.go` method in the domain package (no HTTP, no db imports)
3. **Persistence** → `pgxrepo.go` in the domain package (maps `db.*` ↔ domain types, owns transactions)
4. **HTTP handling** → `web/handler/` (parse form → call service → render template)
5. **Templates** → `web/*.templ` (accept domain types directly)

## Project Commands

- `just openspec <args>` — run OpenSpec commands (never `openspec` directly)
- `just test` — run all tests (`just test ./internal/web/...` to target a package)
- `just test_races` — run all tests with data race detection (use after completing a feature)
- `just fmt` — format all Go code
- `just vet` — run go vet
- `just generate` — run `templ generate` and `sqlc generate` (also runs automatically before build)
- `just migrate <command>` — run goose migrations (`up`, `down`, `status`, `validate`, etc.)
- `just migrate_create <name>` — create a new sequential SQL migration
- Docker Compose for PostgreSQL

## Development Workflow

**Strict TDD cycle. No exceptions.**

1. Write ONE failing test for the observable behaviour you're implementing
2. Write the minimum code to make that test pass
3. Refactor if needed
4. Repeat

Edge cases come LATER — first get the happy path working. Do NOT write multiple failing tests at once. Do NOT write code before the test exists.

When writing a failing test, also create skeleton implementations (functions/methods that return zero values or `errors.New("not implemented")`) so the code compiles. The test must fail for the right **behavioral** reason, not because the function doesn't exist.

**At the end of every feature/section**, run the full quality gate: `just check` (which runs `fmt`, `vet`, `fix`, `test`) and then `just test_races` for data race detection. Fix any issues before moving on.

All code MUST be formatted (`just fmt`) and pass `just vet` before committing.

All tests MUST use the `_test` package suffix (e.g., `package auth_test` not `package auth`). Tests only exercise the public API of a module — no testing internal/unexported functions.

## OpenSpec

No active changes. Start a new one with `/opsx:new`.

Main specs: `openspec/specs/` (7 capabilities)
Archive: `openspec/changes/archive/2026-02-23-pharmarecall-mvp/`

## Code Style

- **Interfaces near consumers**: declare interfaces where they are used, not where they are implemented. Handler-side interfaces in `web/handler/`, driven port interfaces in domain `port.go`.
- **Small interfaces**: prefer single-method interfaces. Compose into `Repository` only for wiring convenience.
- **No shared interfaces by default**: only extract an interface to a common package if it's genuinely needed everywhere (e.g., a `Clock` interface wrapping `Now()`). Start with a local interface in the consumer package.
- **Manual mocks in tests**: no mocking frameworks. Write simple structs that implement the interface in test files. Mock only the port you're testing (not all 6).
- **Domain errors as sentinels**: ALL domain errors — including validation errors — MUST be declared as `var Err... = errors.New(...)` in a single `var` block at the top of the domain type file (`user.go`, `pharmacy.go`, `patient.go`, `prescription.go`). Never use inline `errors.New(...)` in service functions. Handlers check `errors.Is()` to map domain errors to HTTP responses.
- **Error wrapping**: always wrap errors with context using `fmt.Errorf("doing something: %w", err)`. Never bare `return err`. Wrapping (`fmt.Errorf` with `%w`) is the only place where errors are created inside functions — sentinel errors handle all other cases.
- **Table-driven tests**: use table-driven tests as the default pattern, especially for calculation and validation logic.
- **Structured logging**: use `log/slog` from stdlib. No `log.Println` or `fmt.Printf` for logging.
- **Use `new(expr)` for pointer values**: Go 1.26 allows `new(someExpression)` — use it instead of helper functions or address-of workarounds for optional pointer fields.
- **Prefer stdlib `slices`, `strings`, `maps` packages**: use `slices.Contains`, `slices.SortFunc`, `strings.Cut`, `maps.Keys` etc. instead of hand-rolled loops. Don't reinvent what the stdlib provides.
- **Run `go fix ./...`**: Go 1.26 modernizers can automatically update code to use newer language and library features. Run it periodically.

## Restrictions

- **Never run git commands.** I manage the repository myself. No commits, no pushes, no branch operations, no staging.
- **Never edit generated files.** Do not edit `*_templ.go` (templ) or `internal/db/*.go` (sqlc). Reading them is OK to check types; editing goes through the `.templ` or `.sql` source files.
- **Frontend code MUST use `/oatsmith`**: when writing or reviewing any HTML/CSS/Templ templates, always invoke the `/oatsmith` skill. Never write frontend markup without it.

## Conventions

- Use `just` as the task runner
- SQL migrations are plain `.sql` files under `db/migrations/`, sequential numbering
- No PostgreSQL enums. Use basic data types (text/varchar) with CHECK constraints for constrained values (e.g., roles, statuses, fulfillment).
- Pharmacy scoping: all patient/prescription/notification queries MUST filter by pharmacy_id
- Order calculation is on-demand (no cron), based on consumption rate and box start date
- All database writes MUST use a transaction (handled in `pgxrepo.go` adapters)
