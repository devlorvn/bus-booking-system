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
	METRICS_PORT=9091 go run cmd/bus-service/main.go

dev-booking-service:
	METRICS_PORT=9092 go run cmd/booking-service/main.go

dev-payment-service:
	METRICS_PORT=9094 go run cmd/payment-service/main.go

dev-notification-service:
	METRICS_PORT=9095 go run cmd/notification-service/main.go

dev-ws-service:
	METRICS_PORT=8082 go run cmd/ws-service/main.go

dev-user-service:
	METRICS_PORT=9093 go run cmd/user-service/main.go

dev-all:
	@make -j 7 dev-user-service dev-bus-service dev-booking-service dev-payment-service dev-notification-service dev-ws-service dev-api-gateway

build:
	go build -o bin/api cmd/api/main.go

seed:
	go run cmd/seed/main.go

gen-proto:
	go run scripts/gen_proto.go

deploy-local:
	act -W .github/workflows/cd-local.yml -e deployment/local/event.json --secret-file deployment/local/.secret --bind

deploy-local-fast:
	powershell -ExecutionPolicy Bypass -File scripts/deploy_local.ps1

prune-swarm:
	ssh -o StrictHostKeyChecking=no -i deployment/local/id_rsa manager@192.168.220.128 "bash -s" < scripts/run_prune.sh