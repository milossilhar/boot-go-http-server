# boot-go-http-server

Small HTTP server built while following Boot.dev’s **Learn HTTP Servers in Go** course, which covers routing, middleware, JSON APIs, auth/JWTs, webhooks, and production-minded patterns in Go ([course link](https://www.boot.dev/courses/learn-http-servers-golang)).

## Getting started

Prereqs:
- Go (see `go.mod` for version)
- [just](https://github.com/casey/just) command runner
- PostgreSQL (used by migrations)
- Optional: `goose` and `sqlc` installed if you want to run DB tasks

## Useful `just` commands

- `just run` — start the server (`go run .`)
- `just test` — run auth package tests
- `just testv` — verbose auth tests
- `just migrate` — apply database migrations via goose (uses local Postgres connection string in `.justfile`)
- `just rollback` — roll back the last migration
- `just sqlc` — regenerate SQLC code from queries

You can list all available tasks with `just --list`.
