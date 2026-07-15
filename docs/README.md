# Bus Booking System — Distributed Realtime Booking Platform

## Overview

Bus Booking System is a backend-focused learning project designed to simulate a real-world distributed booking platform.

The project focuses on:

- Realtime seat booking
- Concurrent booking handling
- Distributed locking
- Event-driven architecture
- Async processing
- Strong consistency in transactions
- Clean Architecture in Go
- Service communication using gRPC
- Realtime communication using WebSocket
- Background workers and booking expiration

This project is intentionally designed as a **backend-heavy system** with a simple static frontend used only for testing APIs and realtime flows.

---

# Goals

This project is built for learning and practicing:

## Backend & Distributed System Concepts

- Clean Architecture
- Domain-driven modular design
- REST API with Gin
- WebSocket realtime communication
- Redis distributed locking
- PostgreSQL transaction consistency
- Concurrent booking handling
- Background workers
- Event-driven architecture
- gRPC service communication
- Kafka asynchronous messaging
- Cache strategies
- Booking expiration workflows

---

# Tech Stack

| Technology | Purpose |
|---|---|
| Go (Golang) | Main programming language |
| Gin | REST API framework |
| PostgreSQL | Primary database |
| Redis | Locking, caching, pub/sub |
| WebSocket | Realtime seat updates |
| gRPC | Internal service communication |
| Kafka | Async event streaming & DLQ |
| log/slog | JSON structured logging |
| Docker | Local infrastructure |
| HTML + Vanilla JS | Simple frontend for testing |

---

# System Architecture

```txt
                           +----------------+
                           | Static Frontend|
                           | HTML + JS      |
                           +--------+-------+
                                    |
                                    v
                         +----------+----------+
                         | API Gateway / Gin   | <--- HTTP Health (/health)
                         +----------+----------+
                                    |
                +-------------------+-------------------+
                | (gRPC)                                | (gRPC)
                v                                       v
        +-------+--------+                     +--------+-------+
        | Booking Service|<----gRPC---------->| Bus Service    |
        | - Health Check |                     | - Health Check |
        +-------+--------+                     +--------+-------+
            |        |
            |        +-------------------+
            v                            v
    +-------+--------+           +-------+--------+
    | Redis           |           | PostgreSQL     |
    | - Locks         |           | - Outbox Table |
    | - Cache         |           +-------+--------+
    | - Pub/Sub       |                   |
    +----------------+                    | (Atomic Write)
                                          v
                                  [Outbox Worker]
                                          |
                                          v (Publish Events)
                                  +----------------+
                                  | Kafka Brokers  |
                                  | - booking.topic|
                                  | - payment.topic|
                                  | - payment.dlq  |
                                  +----------------+
```

---

# Core Features

# 1. Bus Management

Manage bus trips and seat configurations.

## Features

- Create bus trips
- Configure seat counts
- Define departure and destination
- Manage departure schedule
- Track available seats

---

# 2. Realtime Seat Booking

Users can select seats in realtime.

## Features

- Live seat updates using WebSocket
- Seat locking with expiration
- Prevent double booking
- Multi-seat booking support

---

# 3. Booking Workflow

## Booking Flow

```txt
User selects seat
    ->
Seat locked in Redis
    ->
Realtime broadcast to other users
    ->
User enters information
    ->
Fake payment processing
    ->
Booking confirmation
    ->
Database transaction commit
```

---

# 4. Fake Payment Simulation

A fake payment workflow is implemented to simulate real production booking systems.

## Booking States

```txt
PENDING
PENDING_PAYMENT
CONFIRMED
CANCELLED
EXPIRED
```

---

# 5. Booking Expiration

Locked seats automatically expire if users do not complete payment.

## Example

```txt
Seat A1 locked
TTL: 5 minutes

If payment not completed:
    ->
Booking expired
    Seat released
        ->
    Realtime broadcast triggered
```

---

# 6. Transactional Outbox Pattern

To prevent **Dual-Write** issues (where a DB write succeeds but publishing to Kafka fails or vice-versa), the system implements the **Transactional Outbox Pattern**:

- **Atomic Transactions:** Domain actions (Confirming, Expiring, or Cancelling bookings) execute inside a PostgreSQL transaction. In the same transaction, a payload is written to the `outbox` table in a `PENDING` state.
- **Outbox Worker:** A background runner polls the `outbox` table every `500ms`, retrieves pending events, publishes them to Kafka, and marks them as `PROCESSED`.
- **Automatic Pruning:** An hourly background job runs within the worker to delete `PROCESSED` outbox events older than 24 hours, keeping database storage clean.

---

# 7. Saga Choreography Pattern

For distributed transaction coordination, the system implements an **Event-Driven Saga Choreography**:

- **Happy Path:**
  1. Booking is created (`PENDING_PAYMENT`).
  2. Booking Service writes the `booking.created` event to the Outbox.
  3. Payment Service consumes the event and simulates processing.
  4. Payment succeeds and emits `payment.processed` with status `SUCCESS`.
  5. Booking Service consumes the success event and confirms the booking.
- **Compensating Transaction (Payment Failure):**
  - If the simulated payment fails, Payment Service emits `payment.processed` with status `FAILED`.
  - Booking Service consumes the failed event and triggers a **compensating transaction**: cancels the booking and writes a `booking.cancelled` event to the Outbox.
  - Bus Service consumes `booking.cancelled` and releases the seats back to `AVAILABLE`.

---

# 8. Reliable Consumer with Dead Letter Queue (DLQ)

To guarantee resilient and fault-tolerant event processing:
- **Manual Offset Commit:** Consumers fetch messages manually and only commit offsets once the message has been successfully handled.
- **Exponential Backoff:** If event consumption fails due to transient errors, the consumer retries with an exponential backoff delay (`constants.DelayRetry` doubled each time).
- **Dead Letter Queue (DLQ):** If processing fails after exceeding the maximum retry limit (`MaxKafkaEventRetry`), the message is published to `payment.dlq` with detailed headers (error reason, attempt count, failed timestamp) to prevent blocking the partition.

---

# 9. Production-Grade Optimizations

To prepare the application for real production deployments, the following enterprise features were added:
- **Structured JSON Logging:** Switched all standard print/log statements to Go's built-in `log/slog` structured logging, outputting log statements in JSON format for easy ingestion by ELK, Loki, or CloudWatch.
- **Database Connection Pooling:** Configured GORM connection pool parameters (`SetMaxIdleConns(10)`, `SetMaxOpenConns(100)`, and `SetConnMaxLifetime(time.Hour)`) to avoid exhausting database connections.
- **Health Probes:**
  - Registered official gRPC Health Checking Protocol (`grpc_health_v1`) on all gRPC servers (Booking, Bus, and User services).
  - Exposed a REST `/health` endpoint on the API Gateway for Kubernetes/load balancer readiness and liveness checks.

---

# Realtime System

# WebSocket

Realtime communication is implemented using WebSocket.

## Use Cases

- Seat locked
- Seat released
- Booking confirmed
- Online user tracking
- Realtime seat availability updates

---

# Redis Usage

Redis plays multiple important roles.

## 1. Distributed Locking

Prevent double booking.

### Example

```txt
SETNX seat:A1 temp_user_123 EX 300
```

---

## 2. Cache Layer

Cache frequently accessed data.

### Examples

- Bus details
- Seat availability
- Active trips

---

## 3. Pub/Sub

Synchronize WebSocket events across instances.

---

## 4. TTL Management

Automatically expire temporary locks.

---

# Database Design

# Domains

## Bus

```go
type Bus struct {
    ID
    LicensePlate
    From
    To
    DepartureTime
    TotalSeats
    AvailableSeats
}
```

---

## Seat

```go
type Seat struct {
    ID
    BusID
    Code
    Status
}
```

### Example Codes

```txt
A1 A2 A3
B1 B2 B3
C1 C2 C3
D1 D2 D3
```

---

## Temporary User

Users are anonymous.

```go
type TempUser struct {
    ID
    LastActive
}
```

---

## Booking

```go
type Booking struct {
    ID
    BusID
    TempUserID
    Status
    PaymentStatus
    CreatedAt
}
```

---

## BookingSeat

Supports multi-seat booking.

```go
type BookingSeat struct {
    BookingID
    SeatID
}
```

---

# Seat Status

```txt
AVAILABLE
LOCKED
BOOKED
```

---

# Booking Status

```txt
PENDING
PENDING_PAYMENT
CONFIRMED
CANCELLED
EXPIRED
```

---

# API Design

# Bus APIs

```http
GET    /buses
GET    /buses/:id
GET    /buses/:id/seats
POST   /buses
```

---

# Booking APIs

```http
POST   /bookings/lock
POST   /bookings/confirm
POST   /bookings/cancel
GET    /bookings/:id
```

---

# WebSocket

```http
/ws/buses/:id
```

---

# WebSocket Events

## Seat Locked

```json
{
  "event": "seat_locked",
  "seat": "A1"
}
```

---

## Seat Released

```json
{
  "event": "seat_released",
  "seat": "A1"
}
```

---

## Booking Confirmed

```json
{
  "event": "booking_confirmed",
  "booking_id": "..."
}
```

---

# Service Communication

# gRPC

Used for internal communication between services.

## Example Services

### Bus Service

```proto
service BusService {
  rpc GetBus(GetBusRequest) returns (BusResponse);
  rpc GetSeats(GetSeatsRequest) returns (SeatsResponse);
}
```

---

### Booking Service

```proto
service BookingService {
  rpc CreateBooking(CreateBookingRequest) returns (BookingResponse);
}
```

---

# Event-Driven Architecture

# Kafka Topics

```txt
booking.created
booking.confirmed
booking.expired
seat.locked
seat.released
payment.completed
```

---

# Example Event Flow

```txt
Booking Confirmed
    ->
Publish booking.confirmed
    ->
Notification Service consumes event
    ->
Analytics Service consumes event
```

---

# Transaction Strategy

PostgreSQL is the source of truth.

Redis is used only for temporary state management.

## Booking Transaction

```txt
BEGIN

Validate seats
Check seat availability
Create booking
Insert booking seats
Update seat status

COMMIT
```

If any step fails:

```txt
ROLLBACK
```

---

# Concurrency Handling

One of the main learning goals of this project is handling concurrent booking safely.

## Problems Solved

- Double booking
- Race conditions
- Stale locks
- Realtime synchronization

---

# Clean Architecture

```txt
/internal
    /booking
        /domain
        /usecase
        /repository
        /delivery
            /http
            /grpc
            /ws

    /bus
    /payment
    /user

/pkg
    /postgres
    /redis
    /kafka
    /grpc
    /logger
    /websocket

/cmd
    /api
```

---

# Project Phases

# Phase 1 — Core Foundation

- PostgreSQL setup
- Clean Architecture
- CRUD APIs
- Bus and Seat management

---

# Phase 2 — Realtime Booking

- WebSocket implementation
- Seat locking
- Redis integration
- Live seat updates

---

# Phase 3 — Booking Consistency

- Database transactions
- Concurrent booking handling
- Seat validation
- Expiration worker

---

# Phase 4 — Service Communication

- gRPC integration
- Split services logically
- Internal communication

---

# Phase 5 — Event-Driven Architecture

- Kafka integration
- Async workers
- Booking events
- Notification processing

---

# Advanced Topics

## Future Improvements

- Distributed tracing (OpenTelemetry)
- Prometheus + Grafana metrics
- Rate limiting
- Authentication & Authorization
- Payment gateway integration
- Horizontal WebSocket scaling

---

# Learning Outcomes

By completing this project, developers will gain hands-on experience with:

- Realtime systems
- Distributed locking
- Event-driven architecture
- Consistent transaction handling
- WebSocket scalability
- Redis concurrency patterns
- gRPC service communication
- Kafka asynchronous processing
- Backend system design
- Production-style architecture

---

# Development Philosophy

This project intentionally starts as a:

```txt
Modular Monolith
```

instead of a full microservice architecture.

The goal is to:

- Keep development manageable
- Focus on backend concepts
- Avoid infrastructure complexity too early
- Allow gradual evolution into distributed services

---

# Final Notes

This project is designed primarily as a backend engineering learning platform.

The frontend is intentionally minimal.

The primary focus is:

- System design
- Concurrency
- Distributed consistency
- Realtime communication
- Service architecture
- Scalable backend patterns
