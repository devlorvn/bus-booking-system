Edited main.go
Edited main.go
Edited main.go
Viewed kafka%20payment%20migration.md:72-85

Chính xác rồi bạn! Dưới đây là mô hình chi tiết từng bước của luồng đi (flow) hiện tại để bạn dễ hình dung và lưu lại:

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant BookingService as Booking Service
    participant BusService as Bus Service
    participant PaymentService as Payment Service

    Client->>BookingService: 1. ConfirmBooking (HTTP/gRPC)
    Note over BookingService: Tạo booking: PENDING_CONFIRMATION<br/>Lưu Outbox: booking_created
    BookingService-->>Client: Trả về trạng thái PENDING_CONFIRMATION

    Note over BookingService: OutboxWorker publish:<br/>booking_created (Kafka)

    BusService->>BusService: 2. BookingWorker nhận booking_created
    alt Giữ ghế thành công
        BusService-->>BookingService: Gửi event seats_reserved (Kafka)
        Note over BookingService: SagaResultWorker nhận:<br/>Cập nhật DB: PENDING_PAYMENT<br/>Lưu Outbox: booking_pending_payment
        Note over BookingService: OutboxWorker publish:<br/>booking_pending_payment (Kafka)

        PaymentService->>PaymentService: 3. Nhận booking_pending_payment
        Note over PaymentService: Giả lập xử lý thanh toán
        PaymentService-->>BookingService: Gửi kết quả thanh toán (Topic: payment-events)

        BookingService->>BookingService: 4. PaymentWorker nhận kết quả
        alt Thanh toán thành công (SUCCESS)
            Note over BookingService: Cập nhật DB: PAID / COMPLETED<br/>Lưu Outbox: booking_confirmed (Gửi WS client)
            Note over BookingService: Gọi gRPC confirm giữ ghế vĩnh viễn ở Bus Service
        else Thanh toán thất bại (FAILED)
            Note over BookingService: Cập nhật DB: FAILED<br/>Lưu Outbox: booking_cancelled (Gửi WS & Bus Service)
            Note over BookingService: OutboxWorker gửi booking_cancelled<br/>để Bus Service giải phóng ghế
        end
    else Giữ ghế thất bại (Hết vé/Trùng ghế)
        BusService-->>BookingService: Gửi event seats_reservation_failed (Kafka)
        Note over BookingService: SagaResultWorker nhận:<br/>Cập nhật DB: FAILED<br/>Lưu Outbox: booking_cancelled (Gửi WS giải phóng lock)
    end
```

### Các trạng thái và sự kiện đi kèm:

1. **Đặt vé ban đầu**:
   - DB Booking: `PENDING_CONFIRMATION`.
   - Sự kiện Kafka phát đi: `booking_created` (Header `event_type`).
2. **Xác nhận giữ ghế bất đồng bộ**:
   - Nếu thành công:
     - DB Booking chuyển thành: `PENDING_PAYMENT`.
     - Sự kiện Kafka phát đi: `booking_pending_payment`.
   - Nếu thất bại:
     - DB Booking chuyển thành: `FAILED`.
     - Sự kiện Kafka phát đi: `booking_cancelled` (để thông báo Client qua WebSocket hủy vé).
3. **Thanh toán**:
   - `payment-service` lắng nghe sự kiện `booking_pending_payment`, xử lý xong gửi phản hồi qua topic `payment-events`.
   - `booking-service` nhận kết quả thanh toán:
     - Thành công: Chuyển thành `PAID` và gửi request gRPC cập nhật ghế vĩnh viễn bên Bus Service.
     - Thất bại: Chuyển thành `FAILED` và gửi sự kiện `booking_cancelled` để Bus Service hủy giữ ghế và hoàn lại ghế trống.
