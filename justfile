_default:
  just --list

fmt:
  go fmt ./...

vet:
  go vet ./...

fix:
  go fix ./...

test target="./...":
  go test {{target}}

test_races target="./...":
  go test -race {{target}}

generate:
  go tool templ generate && go tool sqlc generate

build: generate
  go build -o bin/pharmarecall ./cmd/server

build-prod: generate
  CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/pharmarecall ./cmd/server

db_url := "postgres://pharmarecall:pharmarecall@localhost:5432/pharmarecall?sslmode=disable"

migrate *args:
  go tool goose -dir db/migrations -s postgres "{{db_url}}" {{args}}

migrate_create name:
  go tool goose -dir db/migrations -s postgres "{{db_url}}" create {{name}} sql

seed email password:
  go run ./cmd/seed --email {{email}} --password {{password}}

check: fmt vet fix test

openspec *args:
  npx @fission-ai/openspec@latest {{args}}
