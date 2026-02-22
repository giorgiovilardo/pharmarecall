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

- Server-rendered, no SPA, no JS framework
- Single binary deployment — all assets (oat.ink CSS) and migrations embedded via `embed.FS`
- Domain-driven project layout under `internal/`
- Three roles: admin, pharmacy owner, pharmacy personnel

## Project Commands

- `just openspec <args>` — run OpenSpec commands (never `openspec` directly)
- `just test` — run all tests (with data race detection)
- `just fmt` — format all Go code
- `just vet` — run go vet
- `just generate` — run `templ generate` and `sqlc generate` (also runs automatically before build)
- Docker Compose for PostgreSQL

## Development Workflow

**Strict TDD cycle. No exceptions.**

1. Write ONE failing test for the observable behaviour you're implementing
2. Write the minimum code to make that test pass
3. Refactor if needed
4. Repeat

Edge cases come LATER — first get the happy path working. Do NOT write multiple failing tests at once. Do NOT write code before the test exists.

All code MUST be formatted (`just fmt`) and pass `just vet` before committing.

All tests MUST use the `_test` package suffix (e.g., `package auth_test` not `package auth`). Tests only exercise the public API of a module — no testing internal/unexported functions.

## OpenSpec

Active change: `pharmarecall-mvp` — all artifacts complete (proposal, design, specs, tasks). Resume implementation with `/opsx:apply`.

Artifacts location: `openspec/changes/pharmarecall-mvp/`

## Code Style

- **Interfaces near consumers**: declare interfaces where they are used, not where they are implemented
- **Small interfaces**: prefer single-method interfaces
- **No shared interfaces by default**: only extract an interface to a common package if it's genuinely needed everywhere (e.g., a `Clock` interface wrapping `Now()`). Start with a local interface in the consumer package.
- **Manual mocks in tests**: no mocking frameworks. Write simple structs that implement the interface in test files.
- **Error wrapping**: always wrap errors with context using `fmt.Errorf("doing something: %w", err)`. Never bare `return err`.
- **Table-driven tests**: use table-driven tests as the default pattern, especially for calculation and validation logic.
- **Structured logging**: use `log/slog` from stdlib. No `log.Println` or `fmt.Printf` for logging.
- **Use `new(expr)` for pointer values**: Go 1.26 allows `new(someExpression)` — use it instead of helper functions or address-of workarounds for optional pointer fields.
- **Prefer stdlib `slices`, `strings`, `maps` packages**: use `slices.Contains`, `slices.SortFunc`, `strings.Cut`, `maps.Keys` etc. instead of hand-rolled loops. Don't reinvent what the stdlib provides.
- **Run `go fix ./...`**: Go 1.26 modernizers can automatically update code to use newer language and library features. Run it periodically.

## Restrictions

- **Never run git commands.** I manage the repository myself. No commits, no pushes, no branch operations, no staging.
- **Frontend code MUST use `/oatsmith`**: when writing or reviewing any HTML/CSS/Templ templates, always invoke the `/oatsmith` skill. Never write frontend markup without it.

## Conventions

- Use `just` as the task runner
- SQL migrations are plain `.sql` files under `db/migrations/`, sequential numbering
- No PostgreSQL enums. Use basic data types (text/varchar) with CHECK constraints for constrained values (e.g., roles, statuses, fulfillment).
- Pharmacy scoping: all patient/prescription/notification queries MUST filter by pharmacy_id
- Order calculation is on-demand (no cron), based on consumption rate and box start date
- All database writes MUST use a transaction
