include .env

MIGRATE_POSTGRES_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)

migrate-up:
	migrate -path migrations -database "$(MIGRATE_POSTGRES_URL)" up

migrate-down:
	migrate -path migrations -database "$(MIGRATE_POSTGRES_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

dev-api-gateway:
	go run cmd/api/main.go

dev-bus-service:
	go run cmd/bus-service/main.go

dev-booking-service:
	go run cmd/booking-service/main.go

dev-payment-service:
	go run cmd/payment-service/main.go

dev-notification-service:
	go run cmd/notification-service/main.go

dev-ws-service:
	go run cmd/ws-service/main.go

dev-user-service:
	go run cmd/user-service/main.go

dev-all:
	@make -j 7 dev-user-service dev-bus-service dev-booking-service dev-payment-service dev-notification-service dev-ws-service dev-api-gateway

build:
	go build -o bin/api cmd/api/main.go

seed:
	go run cmd/seed/main.go

gen-proto:
	go run scripts/gen_proto.go

deploy-local:
	act -e deployment/local/event.json --secret-file deployment/local/.secret