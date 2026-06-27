# Báo Cáo Đánh Giá Mã Nguồn (Code Review Report)
## Dự Án: Bus Booking System

Bản báo cáo này đánh giá mã nguồn hiện tại của dự án theo hai khía cạnh cốt lõi: **Clean Architecture** (Kiến trúc sạch) và **Go Concurrency** (Xử lý đồng thời với Goroutines & Channels). Ngoài ra, báo cáo cũng đề cập đến các vấn đề về **Hiệu năng & Hệ thống phân tán** phát hiện được trong quá trình duyệt code.

---

## 1. Đánh Giá Kiến Trúc (Clean Architecture)

### 📌 Các Điểm Sáng (Strengths)
1. **Phân lớp rõ ràng (Layer Separation)**: 
   * Dự án phân chia cấu trúc thư mục rất tốt theo mô hình Clean Architecture / Hexagonal Architecture:
     * **Domain (`internal/.../domain`)**: Chứa các thực thể cốt lõi (`Bus`, `Seat`, `Booking`, `User`) hoàn toàn độc lập với các thư viện bên ngoài hay cơ sở dữ liệu.
     * **Usecase (`internal/.../usecase`)**: Chứa logic nghiệp vụ ứng dụng, chỉ phụ thuộc vào các Interfaces (Ports) được định nghĩa sẵn.
     * **Delivery (`internal/.../delivery`)**: Chứa adapter đầu vào như HTTP (Gin), WebSocket, hoặc Background Worker.
     * **Infrastructure (`pkg/...`)**: Chứa các triển khai cụ thể cho database (PostgreSQL/GORM) và Redis.
2. **Dependency Inversion (Đảo ngược phụ thuộc)**: Các usecase giao tiếp với tầng ngoài (database, redis) thông qua các Port interfaces (ví dụ: `BusProvider`, `SeatLockPort`), giúp dễ dàng viết Unit Test (mocking) và thay thế công nghệ lưu trữ.
3. **Transaction Management sạch**: Cách triển khai `Transaction` quản lý qua Go `context.Context` (trong `pkg/database/transaction.go` và `dbFromContext` của repositories) rất thông minh. Nó giúp tầng Usecase điều khiển được transaction mà không hề bị lộ chi tiết GORM (GORM types) lên tầng nghiệp vụ.

---

### ⚠️ Các Điểm Yếu & Khắc Phục (Architectural Issues & Recommendations)

#### Vấn đề 1: Gọi Network/External Service bên trong Database Transaction (Anti-pattern)
* **Vị trí**: `internal/booking/usecase/confirm_booking/usecase.go` (Dòng 170-176)
```go
err = u.tx.Execute(ctx, func(txCtx context.Context) error {
    // ... lưu Booking và BookingSeats ...
    err = u.publisher.PublishPaymentRequested(ctx, bookingID) // <-- Giao tiếp mạng
    if err != nil {
        return err
    }
    return nil
})
```
* **Rủi ro**:
  1. **Nghẽn kết nối database**: Việc thực hiện một cuộc gọi mạng (qua Redis/Kafka/RabbitMQ) bên trong transaction block sẽ giữ kết nối database mở lâu hơn (chờ mạng phản hồi). Nếu mạng chậm hoặc Broker bị treo, kết nối DB sẽ bị chiếm dụng dẫn đến cạn kiệt Connection Pool.
  2. **Lỗi Dual Write (Không nhất quán dữ liệu)**: Nếu sự kiện `PublishPaymentRequested` gửi thành công sang Broker, nhưng ngay sau đó việc commit database ở cuối block `u.tx.Execute` bị thất bại (ví dụ: lỗi trùng khóa, mất mạng với DB), transaction sẽ bị rollback. Tuy nhiên, sự kiện thanh toán **đã được gửi đi**, dẫn đến các service khác xử lý thanh toán cho một Booking không hề tồn tại trong DB!
* **Khắc phục**: Luôn đẩy các hành động không thuộc DB (như gửi Event, gọi API bên thứ ba) ra **ngoài** block Transaction. Chỉ publish event sau khi transaction đã commit thành công.
```go
// Tách biệt việc lưu DB và gửi event
err = u.tx.Execute(ctx, func(txCtx context.Context) error {
    // Chỉ thực hiện các câu lệnh DB tại đây
    return nil
})
if err != nil {
    return nil, err
}

// Publish event bên ngoài transaction
if err := u.publisher.PublishPaymentRequested(ctx, bookingID); err != nil {
    log.Printf("Failed to publish payment requested event: %v", err)
    // Tùy chọn: có thể ghi log lỗi để đối soát thủ công hoặc trả về mã lỗi riêng biệt
}
```

#### Vấn đề 2: Lỗi Thứ Tự Publish Event trong `LockSeatUsecase`
* **Vị trí**: `internal/booking/usecase/lock_seat/usecase.go` (Dòng 60-87)
* **Mô tả**: Usecase duyệt qua danh sách ghế và gọi `u.eventPublisher.PublishSeatLocked(...)` trước khi thực sự gọi `u.seatLockPort.AcquireSeatLocks(...)` để giữ chỗ trên Redis.
* **Rủi ro**: Nếu việc lock ghế trên Redis thất bại (ví dụ: ghế đã bị người khác lock trước đó, hoặc lỗi kết nối Redis), hệ thống vẫn trả về lỗi cho người dùng, nhưng các client WebSocket khác **đã nhận được sự kiện ghế đã khóa** từ trước đó. Gây không nhất quán trạng thái UI nghiêm trọng.
* **Khắc phục**: Chỉ publish event sau khi đã lock thành công tất cả các ghế trong Redis.
```go
// 1. Thực hiện lock trên Redis trước
err = u.seatLockPort.AcquireSeatLocks(ctx, input.BusID, input.SeatCodes, input.TempUserID.String())
if err != nil {
    return nil, err
}

// 2. Chỉ khi lock thành công mới publish event thông báo cho các client khác
for _, seat := range seats {
    _ = u.eventPublisher.PublishSeatLocked(
        input.BusID.String(),
        seat.ID.String(),
        seat.SeatCode,
        input.TempUserID.String(),
    )
}
```

#### Vấn đề 3: Thiếu Orchestration (Lắp ráp code) cho `ConfirmBookingUsecase`
* **Mô tả**: `ConfirmBookingUsecase` được viết khá hoàn chỉnh ở tầng usecase, nhưng hoàn toàn **chưa được khởi tạo hay đăng ký vào router** trong `cmd/api/main.go`. Endpoint xác nhận đặt vé chưa khả dụng cho client gọi tới.
* **Khắc phục**: Khởi tạo và truyền dependencies cho `ConfirmBookingUsecase` trong `main.go`, sau đó bổ sung API route (ví dụ: `POST /api/bookings/confirm`) để đón nhận request.

---

## 2. Đánh Giá Xử Lý Đồng Thời (Concurrency & Go Routines)

Đây là khía cạnh xuất hiện nhiều lỗi tiềm tàng nhất, có thể gây sập hệ thống (panic) hoặc rò rỉ tài nguyên (goroutine leak).

### 🚨 Lỗi 1: Panic "close of closed channel" trong WebSocket Hub (Critical)
* **Vị trí**: `internal/booking/delivery/ws/hub.go` và `client.go`
* **Nguyên nhân**:
  1. Trong `hub.go` hàm `handleBroadcast`: Nếu kênh gửi (`client.Send`) bị đầy (do client xử lý chậm), hub sẽ thực hiện block `default` để giải phóng client:
     ```go
     default:
         close(client.Send) // <-- Đóng channel lần 1
         delete(room, client)
     ```
  2. Kênh `client.Send` bị đóng sẽ làm vòng lặp `WritePump` của client đó kết thúc. 
  3. Do đó, hàm `ReadPump` của client cũng sẽ thoát và gọi hàm đóng kết nối ở khối `defer`:
     ```go
     defer func() {
         c.Hub.unregister <- c // <-- Gửi client vào kênh unregister
         c.Conn.Close()
     }()
     ```
  4. Hub nhận được client từ kênh `unregister` và gọi `handleUnregister(client)`:
     ```go
     func (h *Hub) handleUnregister(client *Client) {
         room, exists := h.rooms[client.BusID]
         // ...
         delete(room, client)
         close(client.Send) // <-- Đóng channel lần 2! Gây PANIC ứng dụng!
     }
     ```
* **Khắc phục**:
  Cách an toàn nhất là chỉ đóng channel `client.Send` tại một nơi duy nhất (`handleUnregister`) và kiểm tra xem client có thực sự còn tồn tại trong phòng hay không trước khi xóa và đóng:
  ```go
  func (h *Hub) handleUnregister(client *Client) {
      room, exists := h.rooms[client.BusID]
      if !exists {
          return
      }
      
      // Chỉ đóng và xóa nếu client vẫn còn trong phòng (chưa bị xóa ở nơi khác)
      if _, ok := room[client]; ok {
          delete(room, client)
          close(client.Send)
      }
  
      if len(room) == 0 {
          delete(h.rooms, client.BusID)
      }
  }
  ```

---

### 🚨 Lỗi 2: Rò Rỉ Goroutine (Goroutine Leak) Khi Dừng Ứng Dụng
* **Vị trí**: `internal/booking/delivery/ws/hub.go` và `client.go`
* **Nguyên nhân**:
  1. Hàm `hub.Run()` chạy bằng một vòng lặp vô hạn `for { select { ... } }` mà không có cơ chế lắng nghe Context cancellation để thoát.
  2. Khi ứng dụng tắt hoặc khởi động lại (graceful shutdown), nếu `hub.Run` dừng hoặc thoát trước, kênh `unregister` (không có buffer) sẽ không còn ai lắng nghe.
  3. Khi đó, tất cả các Goroutine `ReadPump` đang chạy của client sẽ bị nghẽn vĩnh viễn ở dòng `c.Hub.unregister <- c` trong khối `defer`, không bao giờ giải phóng được bộ nhớ.
* **Khắc phục**:
  1. Truyền `context.Context` vào `hub.Run(ctx)` và xử lý trường hợp `<-ctx.Done()` để giải phóng hub.
  2. Sử dụng `select` kèm timeout hoặc non-blocking send khi đẩy dữ liệu vào kênh `unregister` ở `ReadPump` để đảm bảo goroutine luôn được giải phóng:
  ```go
  defer func() {
      select {
      case c.Hub.unregister <- c:
      case <-time.After(100 * time.Millisecond): // Tránh block vô hạn nếu hub đã dừng
      }
      c.Conn.Close()
  }()
  ```

---

### 🚨 Lỗi 3: Vòng Lặp Vô Hạn Tiêu Tốn 100% CPU Khi Mất Kết Nối Redis (Critical)
* **Vị trí**: `internal/booking/delivery/worker/lock_expiration.go` (Dòng 33-63)
```go
ch := pubsub.Channel()
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case msg := <-ch: // <-- Đọc từ channel của Redis
        if msg == nil {
            continue // <-- Gây lặp vô tận nếu channel bị đóng
        }
        // ...
    }
}
```
* **Nguyên nhân**: Trong Go, khi một channel bị đóng (ví dụ: khi kết nối Redis bị đứt hoặc subscription bị ngắt), việc đọc từ channel đó (`<-ch`) sẽ ngay lập tức trả về giá trị mặc định (zero value) của kiểu dữ liệu (ở đây là `nil` đối với con trỏ `*redis.Message`) và **không block nữa**. Khối `if msg == nil { continue }` sẽ liên tục chạy mà không hề dừng lại, tạo ra một vòng lặp vô hạn chiếm dụng 100% CPU của luồng ứng dụng.
* **Khắc phục**: Kiểm tra trạng thái đóng của channel bằng cú pháp nhận 2 giá trị (`msg, ok := <-ch`):
```go
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case msg, ok := <-ch:
        if !ok {
            log.Println("Redis pubsub channel closed, stopping worker.")
            return errors.New("redis pubsub channel closed")
        }
        if msg == nil {
            continue
        }
        // ... xử lý event bình thường
    }
}
```

---

## 3. Hiệu Năng & Hệ Thống Phân Tán (Performance & Distributed Locking)

### 🏎️ Vấn đề 1: Lỗi Hiệu Năng N+1 Queries Lên Redis
* **Vị trí**: `internal/bus/usecase/bus.go` hàm `GetSeats` (Dòng 106-114)
```go
for _, seat := range seats {
    isLocked, err := u.seatPort.IsSeatLocked(ctx, busID, seat.SeatCode)
    // ...
}
```
* **Mô tả**: Với mỗi ghế của một xe bus (thường từ 30 đến 45 ghế), hệ thống thực hiện một truy vấn độc lập đến Redis để kiểm tra trạng thái khóa. Điều này tạo ra 30-45 kết nối/round-trip mạng tuần tự đến Redis cho mỗi lượt xem sơ đồ ghế của người dùng.
* **Khắc phục**: Chuyển sang dạng truy vấn lô (Batch Query) bằng lệnh `MGet` của Redis để lấy trạng thái của tất cả các ghế trong **chỉ 1 network round-trip**.
  1. Định nghĩa thêm phương thức vào `SeatPort` interface:
     ```go
     GetLockedSeats(ctx context.Context, busID uuid.UUID, seatCodes []string) (map[string]bool, error)
     ```
  2. Triển khai trong `LockSeatRepository` dùng `MGet`:
     ```go
     func (r *LockSeatRepository) GetLockedSeats(ctx context.Context, busID uuid.UUID, seatCodes []string) (map[string]bool, error) {
         if len(seatCodes) == 0 {
             return nil, nil
         }
         keys := make([]string, len(seatCodes))
         for i, code := range seatCodes {
             keys[i] = buildSeatLockKey(busID, code)
         }
         results, err := r.client.MGet(ctx, keys...).Result()
         if err != nil {
             return nil, err
         }
         lockedMap := make(map[string]bool)
         for i, val := range results {
             if val != nil {
                 lockedMap[seatCodes[i]] = true
             }
         }
         return lockedMap, nil
     }
     ```

---

### 🏎️ Vấn đề 2: Race Condition Khi Giải Phóng Khóa Phân Tán (Distributed Lock Leak)
* **Vị trí**: `pkg/redis/seat_lock_repository.go` hàm `ReleaseSeatLocks` (Dòng 87-98)
```go
owner, err := r.client.Get(ctx, key).Result()
// ...
if owner != tempUserID {
    continue
}
_, err = r.client.Del(ctx, key).Result() // <-- KHÔNG ATOMIC!
```
* **Mô tả**: Tiến trình đọc giá trị khóa (`Get`) và tiến trình xóa khóa (`Del`) diễn ra không nguyên tử (non-atomic). 
* **Rủi ro**: Nếu khóa của User A sắp hết hạn, User A thực hiện giải phóng khóa. Ngay sau câu lệnh `Get` kiểm tra đúng `owner == tempUserID`, khóa hết hạn trên Redis và User B nhanh chóng chiếm lấy khóa của chính chiếc ghế đó. Lúc này câu lệnh `Del` của User A chạy và **xóa nhầm khóa đang hợp lệ của User B**!
* **Khắc phục**: Sử dụng một đoạn mã script Lua ngắn để thực hiện việc kiểm tra và xóa khóa một cách nguyên tử (Atomic Delete) trên Redis:
```go
var releaseLockScript = redis.NewScript(`
    if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("del", KEYS[1])
    else
        return 0
    end
`)

func (r *LockSeatRepository) ReleaseSeatLocks(ctx context.Context, busID uuid.UUID, seatCodes []string, tempUserID string) error {
    for _, seatCode := range seatCodes {
        key := buildSeatLockKey(busID, seatCode)
        _, err := releaseLockScript.Run(ctx, r.client, []string{key}, tempUserID).Result()
        if err != nil && !errors.Is(err, goredis.Nil) {
            return err
        }
    }
    return nil
}
```

---

## 4. Tổng Kết Bảng Đánh Giá & Mức Độ Ưu Tiên

| Lỗi / Vấn đề | Phân Loại Lớp | Mức Độ Nghiêm Trọng | Ảnh Hưởng Thực Tế |
| :--- | :--- | :--- | :--- |
| **Panic "close of closed channel"** | WebSocket Delivery | **🚨 Khẩn cấp (Critical)** | Có thể làm crash/sập hoàn toàn tiến trình API server khi tải cao hoặc kết nối không ổn định. |
| **100% CPU Busy Loop khi mất Redis** | Background Worker | **🚨 Khẩn cấp (Critical)** | Gây nghẽn CPU của Server, dẫn tới treo toàn bộ các API khác khi Redis gặp sự cố ngắn hạn. |
| **Rò rỉ Goroutine** | WebSocket Delivery | **⚠️ Cao (High)** | Làm tăng dung lượng RAM sử dụng của server theo thời gian, cuối cùng gây lỗi OOM (Out Of Memory). |
| **Lỗi thứ tự Event & Network in Tx** | Usecase / DB Transaction | **⚠️ Cao (High)** | Không nhất quán trạng thái hiển thị ghế cho người dùng khác; nghẽn kết nối cơ sở dữ liệu. |
| **Race Condition xóa nhầm lock** | Infrastructure (Redis) | **⚠️ Trung bình (Medium)** | Một người dùng có thể vô tình hủy quyền đặt ghế của người khác trong lúc mạng chập chờn. |
| **N+1 Queries trên Redis** | Usecase / Performance | **⚠️ Trung bình (Medium)** | Làm chậm tốc độ phản hồi danh sách ghế của chuyến xe do tốn nhiều round-trip kết nối. |
| **Chưa đấu nối ConfirmBooking** | Usecase Orchestration | **ℹ️ Thấp (Low)** | Tính năng xác nhận đặt vé chưa chạy thực tế được từ client. |
