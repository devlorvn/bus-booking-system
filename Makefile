include .env

MIGRATE_POSTGRES_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)

migrate-up:
	migrate -path migrations -database "$(MIGRATE_POSTGRES_URL)" up

migrate-down:
	migrate -path migrations -database "$(MIGRATE_POSTGRES_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

dev:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

seed:
	go run cmd/seed/main.go

gen-proto:
	./scripts/gen_proto.sh
