# PharmaRecall

Web application for Italian pharmacies to manage patients with recurring prescriptions. Tracks refill schedules, calculates daily orders, and helps pharmacies prepare packages proactively.

Pharmacy personnel log in, register patients, record their recurring prescriptions (medication, units per box, daily consumption), and the system calculates when each box will run out. A dashboard shows what needs to be prepared each day. Orders move through a simple lifecycle: pending, prepared, fulfilled.

## Tech stack

Go 1.26, PostgreSQL 18, server-rendered HTML with [Templ](https://templ.guide) templates and [oat.ink](https://oat.ink/) CSS. Single binary deployment with embedded migrations and static assets.

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

Run migrations:

```
just migrate up
```

Build and run:

```
just build
./bin/pharmarecall
```

The server starts on `http://localhost:8080`.

## Running everything in Docker

```
docker compose up
```

This builds the app image (multi-stage, runs from scratch) and starts both the app and database. The app waits for the database to be healthy before starting. You need a `config.toml` in the project root — it gets mounted into the container.

Note: when running fully dockerized, the database host in `config.toml` should be `pharmarecall-db` (the compose service name) instead of `localhost`.

## Common commands

```
just test                     # run all tests
just test ./internal/web/...  # run tests for a specific package
just test_races               # run all tests with race detection
just check                    # fmt + vet + fix + test
just generate                 # regenerate templ and sqlc code
just migrate up               # apply pending migrations
just migrate down             # roll back one migration
just migrate status           # show migration status
just migrate_create <name>    # create a new migration file
just build                    # build the binary
just build-prod               # build a production binary (stripped, static)
```

## Project layout

```
cmd/server/        entrypoint
internal/
  auth/            authentication, sessions
  config/          configuration loading
  web/             HTTP handlers, middleware, templates
db/
  migrations/      SQL migration files (goose)
  queries/         SQL query files (sqlc)
static/            static assets (embedded)
```

## TODO

- Proper error pages
- Proper i18n/strings
