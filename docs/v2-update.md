# Lộ Trình Cải Tiến Hệ Thống - Phiên Bản V2 Update

Tài liệu này chi tiết hóa kế hoạch triển khai của **Bước 5** trong lộ trình nâng cấp hệ thống (V2). Chúng ta sẽ giải quyết triệt để vấn đề ranh giới Domain của **Lock Expiration**, tách độc lập **Notification/WebSocket Service**, và nâng cấp **Payment Service** hỗ trợ tích hợp Stripe.

---

## 1. Trạng Thái Hiện Tại & Thách Thức

Sau khi hoàn thành Bước 4, hệ thống đã giao tiếp bất đồng bộ qua Kafka cho thanh toán và đồng bộ WebSocket. Tuy nhiên, vẫn còn 2 điểm yếu lớn về mặt kiến trúc:

1. **Ranh giới Domain bị vi phạm ở API Gateway:**
   - `lockExpirationWorker` và `expireBookingUsecase` đang được chạy trực tiếp tại API Gateway (`cmd/api/main.go`). API Gateway phải kết nối trực tiếp vào Postgres DB của Booking để cập nhật trạng thái hết hạn.
   - Cơ chế bắt sự kiện hết hạn hiện tại dựa trên **Redis Keyspace Notifications (`__keyevent@0__:expired`)**. Đây là cơ chế *At-most-once* (gửi và quên), nếu API Gateway bị restart đúng lúc key hết hạn, sự kiện hủy vé sẽ bị mất vĩnh viễn, dẫn đến treo ghế trong DB.

2. **WebSocket Hub chưa được tách rời:**
   - WebSocket Hub và `WsConsumer` vẫn đang chạy cùng tiến trình với API Gateway. Khi lượng người dùng truy cập đồng thời tăng cao, API Gateway sẽ bị quá tải RAM/CPU do phải duy trì hàng ngàn kết nối WebSocket lâu dài.

---

## 2. Kiến Trúc Mục Tiêu (Target Architecture)

Chúng ta sẽ tách WebSocket thành **Notification Service** độc lập và chuyển toàn bộ logic hủy vé hết hạn về **Booking Service** chạy nền bằng **Redis ZSet (Sorted Set)** làm Delay Queue đảm bảo tin cậy tuyệt đối (*At-least-once*).

```mermaid
graph TD
    %% Clients
    Client([Web Client]) <-->|WebSocket| NotificationService[Notification/WS Service]
    Client -->|HTTP / REST| APIGateway[API Gateway]

    %% Sync Calls
    APIGateway -- gRPC --> BusService[Bus Service]
    APIGateway -- gRPC --> BookingService[Booking Service]

    %% Message Broker
    subgraph Kafka Cluster
        TopicBooking[booking-events]
        TopicPayment[payment-events]
        TopicWS[booking-websocket-events]
    end

    %% Expiration & Delay Queue
    RedisZSet[(Redis ZSet Expiration Queue)] <-->|Poll & Push| ExpirationWorker[ZSet Expiration Worker]
    subgraph Booking Service [Booking Service Process]
        BookingService --> DB_Booking[(Postgres: Bookings)]
        ExpirationWorker --> BookingService
    end

    %% Payment Integration
    subgraph Payment Service [Payment Service Process]
        PaymentWorker[Payment Worker] --> StripeAPI[Stripe Gateway API]
    end

    %% Event Flows
    BookingService -- Publish booking.created --> TopicBooking
    TopicBooking --> PaymentWorker
    PaymentWorker -- Publish payment.processed --> TopicPayment
    TopicPayment --> BookingService
    
    %% Realtime WS updates
    BookingService -- Publish seat_locked/unlocked --> TopicWS
    BusService -- Publish seat_locked/unlocked --> TopicWS
    TopicWS --> NotificationService
```

---

## 3. Lộ Trình Triển Khai Chi Tiết (Implementation Roadmap)

### Phase 5.1: Tách Biệt WebSocket (Notification Service)
Chúng ta sẽ chuyển toàn bộ logic WebSocket ra khỏi API Gateway.

1. **Khởi tạo Notification Service:**
   - Tạo thư mục `cmd/notification-service/`.
   - Di chuyển `internal/booking/delivery/ws/` sang một package dùng chung hoặc khởi tạo trực tiếp trong `cmd/notification-service/` (do service này chỉ làm nhiệm vụ giao tiếp WebSocket).
2. **Cập nhật API Gateway:**
   - Xóa `ws.Hub`, `ws.NewWsConsumer` và router `/ws/buses/:id` khỏi `cmd/api/main.go`.
   - API Gateway chỉ giữ vai trò là stateless REST Gateway (routing HTTP request tới Bus Service và Booking Service).
3. **Cấu hình Client kết nối:**
   - Cập nhật frontend (`web/js/services/websocket.js`) kết nối trực tiếp tới cổng của Notification Service (ví dụ: `ws://localhost:8082/ws/buses/:id`) thay vị trí cũ ở API Gateway.

---

### Phase 5.2: Tái Cấu Trúc Lock Expiration & Redis ZSet Delay Queue
Chuyển logic hết hạn về đúng Booking Service và đảm bảo độ tin cậy.

1. **Chuyển code Expiration Worker:**
   - Di chuyển `lockExpirationWorker` và `expireBookingUsecase` từ `cmd/api/main.go` sang `cmd/booking-service/main.go`.
   - API Gateway không còn kết nối trực tiếp tới DB của Booking hay chứa logic liên quan đến DB Booking nữa.
2. **Implement Redis ZSet Delay Queue:**
   - Viết thư viện helper trong `pkg/redis` hỗ trợ:
     - `PushToDelayQueue(queueName string, id string, expireAt time.Time)`: Đẩy ID booking vào ZSet với score là timestamp hết hạn.
     - `PollFromDelayQueue(queueName string)`: Dùng script Lua hoặc lệnh `ZRANGEBYSCORE` để quét lấy ra các ID đã quá hạn và xử lý.
   - Khi Confirm Booking thành công, Booking Service sẽ đẩy `bookingID` vào ZSet thay vì chỉ tạo key Redis expire thông thường.
3. **Đảm bảo tính tin cậy (At-least-once):**
   - Delay Queue Worker của Booking Service sẽ poll liên tục từ ZSet.
   - Sau khi thực thi `expireBookingUsecase.Execute` cập nhật DB thành công và bắn event lên Kafka, lúc đó mới xóa ID booking khỏi ZSet. 
   - Nếu sập ứng dụng giữa chừng, dữ liệu trong ZSet vẫn còn nguyên và sẽ được xử lý tiếp khi app start lại.

---

## 4. Nâng Cấp Payment Service Thực Tế (Stripe Integration)
Thay thế bộ giả lập thanh toán (Fake Payment Processor) bằng tích hợp Stripe thực tế.

1. **Cấu hình Stripe Client:**
   - Đăng ký tài khoản Sandbox Stripe để lấy API Keys (`STRIPE_SECRET_KEY`).
   - Cài đặt Stripe Go SDK (`github.com/stripe/stripe-go/v72`).
2. **Luồng thanh toán nâng cấp:**
   - Khi nhận `BookingCreatedEvent`, Payment Service sẽ gọi API Stripe để tạo một **PaymentIntent** (hoặc tạo session Checkout Stripe và trả về URL cho client).
   - Client tiến hành thanh toán trên trang Stripe.
3. **Webhook Handler:**
   - Phát triển một API Webhook tại Payment Service để nhận callback từ Stripe khi giao dịch thành công/thất bại.
   - Webhook nhận event `charge.succeeded` hoặc `payment_intent.succeeded` từ Stripe, sau đó publish sự kiện tương ứng (`SUCCESS`/`FAILED`) lên Kafka topic `payment-events`.

---

## 5. Bảng Phân Chia Công Việc Dự Kiến

| Công việc | Vị trí thay đổi | Độ ưu tiên |
| :--- | :--- | :--- |
| **5.1.1** Khởi chạy `cmd/notification-service` | New Service | High |
| **5.1.2** Tách WS Hub ra khỏi API Gateway | `cmd/api/main.go` | High |
| **5.2.1** Chuyển `lockExpirationWorker` về Booking Service | `cmd/booking-service/main.go` | High |
| **5.2.2** Thay thế Redis Expire bằng ZSet Delay Queue | `pkg/redis`, `booking-service` | Critical |
| **5.3.1** Tích hợp Stripe SDK vào Payment Service | `cmd/payment-service` | Medium |
| **5.3.2** Viết Webhook API xử lý kết quả Stripe | `cmd/payment-service` | Medium |
