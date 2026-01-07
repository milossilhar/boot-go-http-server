# GoLang Server - Development
default:
    @just --list

alias r := run
# runs the server
@run:
    go run .

alias t := test
alias tv := testv
@test:
    go test ./internal/auth
@testv:
    go test -v ./internal/auth

# migrates database using goose
@migrate:
    goose -dir ./sql/schema postgres "postgres://postgres:xucUAF7hs54H@localhost:5432/golang" up

# rollbacks database using goose
@rollback:
    goose -dir ./sql/schema postgres "postgres://postgres:xucUAF7hs54H@localhost:5432/golang" down

# generates sqlc files
@sqlc:
    sqlc generate