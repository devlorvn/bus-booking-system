Đây là một thắc mắc **rất chính xác và rất hay** về thiết kế hệ thống phân tán (Distributed Systems)!

Câu trả lời ngắn gọn:
👉 **Đúng vậy, khi dùng `SKIP LOCKED` với nhiều Worker chạy song song, bạn KHÔNG THỂ đảm bảo tính FIFO tuyệt đối trên toàn hệ thống (Global Strict FIFO).**

Vì Worker 2 có thể gửi record 11 lên Kafka xong **trước** khi Worker 1 kịp gửi record 1 (do lag mạng, CPU, v.v.).

---

### 1. Bản chất vấn đề: Global FIFO vs Aggregate FIFO

Trong thực tế hệ thống phân tán, **Global Strict FIFO (Thứ tự tuyệt đối của tất cả mọi người trên toàn hệ thống)** là một "bẫy hiệu năng" (performance bottleneck).

- **Tại sao không cần Global FIFO?**
  - Việc vé của **Khách hàng A** được xử lý trước hay sau vé của **Khách hàng B** vài miligiây **không quan trọng**.
  - Điều **DUY NHẤT quan trọng** là: Tất cả các event của **CÙNG 1 ĐƠN ĐẶT VÉ (Cùng `booking_id`)** phải đúng thứ tự:
    $$\text{BookingCreated} \longrightarrow \text{BookingPendingPayment} \longrightarrow \text{BookingConfirmed}$$

---

### 2. Thực tế các hệ thống lớn xử lý vấn đề này như thế nào?

Có **4 chiến lược thực tế** được áp dụng tùy vào quy mô hệ thống:

#### 🟢 Cách 1: Phân luồng theo Key trên Message Broker (Kafka Partition Key) — _Phổ biến nhất_

Dù các Outbox Worker đọc record khỏi DB theo thứ tự nào, khi publish lên Kafka, họ **phải truyền Message Key** là `AggregateID` (ví dụ `booking_id`):

- Kafka sẽ dùng thuật toán `Hash(booking_id) % total_partitions` để đẩy tất cả event của `Booking_A` vào **cùng một Partition**.
- Kafka **đảm bảo tính FIFO tuyệt đối 100% trên từng Partition**.
- Phía Consumer chỉ dùng **1 thread đọc cho 1 Partition**, do đó Consumer chắc chắn sẽ nhận `BookingCreated` trước `BookingConfirmed` của đơn vé đó.

```
Worker 1 nhặt Record 1 (Booking A - Created)   ───> Kafka Partition 1 (Order: Created -> Confirmed)
Worker 2 nhặt Record 2 (Booking A - Confirmed) ───> Kafka Partition 1
```

---

#### 🟢 Cách 2: Phân chia Outbox theo Partition / Sharding ở cấp DB

Nếu muốn chính các Outbox Worker đọc DB cũng không bao giờ bị chéo event của cùng 1 đơn vé:

- Mỗi Worker được phân công xử lý một tập các `AggregateID` nhất định bằng thuật toán Hash:
  $$\text{Worker ID} = \text{Hash}(booking\_id) \pmod N$$
- Worker 1 chỉ fetch các event có `hash(aggregate_id) % 2 == 0`.
- Worker 2 chỉ fetch các event có `hash(aggregate_id) % 2 == 1`.
- **Kết quả**: Tất cả event của `Booking_A` sẽ luôn luôn do **duy nhất Worker 1** đảm nhiệm từ đầu đến cuối, đảm bảo tuyệt đối FIFO.

---

#### 🟢 Cách 3: Thiết kế Consumer Idempotent & Chống lỗi lệch thứ tự (Out-of-Order Resilient)

Phía Consumer không bao giờ tin tưởng 100% vào thứ tự tin nhắn đến. Họ áp dụng:

1. **State Machine Check**:
   - Nếu Consumer nhận được event `BookingConfirmed` nhưng kiểm tra DB thấy trạng thái booking chưa từng sang `PENDING_PAYMENT`, Consumer sẽ đẩy event đó vào **Retry Queue** hoặc **Delay Queue** để chờ `BookingCreated` đến trước xử lý xong.
2. **Event Versioning / Sequence Number**:
   - Mỗi event được gắn `version: 1`, `version: 2`. Consumer lưu `last_processed_version`. Nếu nhận event `v2` trước `v1`, Consumer sẽ reject/re-queue `v2`.

---

#### 🟢 Cách 4: Single Outbox Worker Instance (Nếu lượng traffic vừa/nhỏ)

- Nếu hệ thống không có hàng triệu giao dịch/giây, **chỉ cần chạy DUY NHẤT 1 instance của Outbox Worker** (dùng Leader Election qua Redis lock / Kubernetes StatefulSet).
- Vì chỉ có 1 Worker chạy đơn luồng (hoặc dùng goroutine pool có ordering), tính FIFO tuyệt đối từ DB ra Kafka luôn được bảo đảm 100%.

---

### 💡 Tóm lại với dự án `booking-system` của bạn:

1. Trong dự án này, cách tốt nhất là **dùng `booking_id` làm Key khi Publish Kafka**:
   ```go
   // Gửi Message Key = bookingID.String() vào Kafka
   kafkaWriter.WriteMessages(ctx, kafka.Message{
       Key:   []byte(event.BookingID.String()),
       Value: payloadBytes,
   })
   ```
2. Việc sử dụng `FOR UPDATE SKIP LOCKED` giúp các Outbox Worker chạy **tối đa hiệu năng (Horizontal Scaling)** mà không lo bắn trùng lặp event hay nghẽn DB. Kafka sẽ chịu trách nhiệm giữ đúng thứ tự FIFO cho từng `booking_id`.
