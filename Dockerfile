FROM golang:1.26-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG SERVICE_NAME

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o service cmd/${SERVICE_NAME}/main.go

From alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/service .

COPY --from=builder /app/configs/ configs/

EXPOSE 8080

ENTRYPOINT ["./service"]
