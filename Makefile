SHELL := /bin/bash

run:
	source .env && go run cmd/main.go

migrate-up:
	source .env && test -n "$$POSTGRES_DSN" && psql "$$POSTGRES_DSN" -f migrations/001_init_users_tasks.up.sql

migrate-down:
	source .env && test -n "$$POSTGRES_DSN" && psql "$$POSTGRES_DSN" -f migrations/001_init_users_tasks.down.sql

.PHONY: run migrate-up migrate-down