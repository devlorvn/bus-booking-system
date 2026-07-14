Tiếp tục với **Choreography-based Saga (Phương án 2)**, đây là mô hình thiết kế tối ưu nhất cho giao tiếp giữa các Microservices khi cần thực hiện các giao dịch phân tán (Distributed Transactions) mà không muốn các service bị ràng buộc cứng (tight coupling) bởi các cuộc gọi API đồng bộ (gRPC/REST).

Dưới đây là thiết kế chi tiết và các bước triển khai cụ thể cho mô hình Saga trong hệ thống Bus Booking của chúng ta.

---

## I. Nguyên Lý Hoạt Động Của Saga trong Luồng Hủy Vé

Thay vì **Booking Service** phải chủ động đi gọi **Bus Service** để giải phóng ghế, luồng đi của sự kiện sẽ được đảo ngược (Event-Driven):

```mermaid
sequenceDiagram
    autonumber
    participant Payment as Payment Service
    participant Booking as Booking Service
    participant Kafka as Kafka Broker
    participant Bus as Bus Service

    Payment->>Kafka: Publish PaymentProcessedEvent (FAILED)
    Kafka->>Booking: Consume PaymentProcessedEvent (FAILED)
    Note over Booking: Cập nhật Booking sang FAILED
    Booking->>Kafka: Publish BookingCancelledEvent
    Note over Booking: Kết thúc nhiệm vụ Booking Service!

    Kafka->>Bus: Consume BookingCancelledEvent
    Note over Bus: Tự giải phóng ghế trong DB Bus
```