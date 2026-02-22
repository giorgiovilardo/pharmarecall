_default:
  just --list

fmt:
  go fmt ./...

vet:
  go vet ./...

fix:
  go fix ./...

test *args:
  go test -race ./... {{args}}

generate:
  templ generate && sqlc generate

build: generate
  go build -o bin/pharmarecall ./cmd/server

build-prod: generate
  CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/pharmarecall ./cmd/server

check: fmt vet fix test

openspec *args:
  npx @fission-ai/openspec@latest {{args}}
