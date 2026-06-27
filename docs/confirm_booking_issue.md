Dưới đây là phân tích chi tiết về vấn đề Race Condition khi xác nhận đặt vé (Confirm Booking), cùng với giải pháp Trade-off tối ưu nhất cho dự án của bạn và cách chúng ta vừa triển khai trực tiếp vào mã nguồn.

1. Ý kiến của tôi về lỗi Race Condition này
Vấn đề cốt lõi ở đây là sự phối hợp trạng thái giữa hai hệ thống lưu trữ độc lập:

Redis: Đóng vai trò là Optimistic Guard (Khóa tạm thời với TTL ngắn hạn để tối ưu hiệu năng đọc/ghi và mang lại UX mượt mà).
Postgres: Đóng vai trò là Source of Truth (Trạng thái đặt vé vĩnh viễn và chính xác tuyệt đối).
Do không có giao dịch phân tán (distributed transaction) để đảm bảo tính nguyên tử (atomic) trên cả 2 hệ thống này cùng lúc, việc giải phóng khóa trên Redis trước khi transaction Postgres commit sẽ tạo ra một khoảng thời gian sơ hở (dangerous window):

Lúc này, Redis đã bị xóa khóa (trở thành AVAILABLE).
Nhưng Postgres vẫn chưa commit xong transaction (trạng thái seat vẫn là AVAILABLE đối với các transaction đọc khác do cơ chế MVCC / Read Committed của Postgres).
Hệ quả: User khác có thể ngay lập tức nhảy vào lock ghế đó trên Redis và tiến hành confirm, dẫn đến tình trạng tranh chấp nặng nề. Ngoài ra, trong bảng booking_seats của bạn, khóa chính là cặp (booking_id, seat_id), tức là không có ràng buộc duy nhất (UNIQUE) cho riêng seat_id. Do đó, nếu không có cơ chế khóa ở mức ghế (seat status) trong Postgres, lỗi double-booking (ghế bị bán 2 lần) chắc chắn sẽ xảy ra.
2. Giải pháp Best Trade-off đề xuất (Hybrid Solution)
Giải pháp tối ưu và an toàn nhất là sử dụng Postgres làm chốt chặn cuối cùng nguyên tử (Atomic Database-level guard) và thay đổi thứ tự giải phóng khóa Redis:

Không giải phóng khóa Redis trước khi Postgres commit. Chỉ giải phóng sau khi Postgres transaction đã thành công hoàn toàn.
Thực hiện khóa và kiểm tra trạng thái ghế trực tiếp trong Database Transaction bằng một câu lệnh cập nhật nguyên tử duy nhất (Atomic Update):
sql
UPDATE seats SET status = 'BOOKED' 
WHERE bus_id = ? AND seat_code IN ? AND status = 'AVAILABLE';
Tại sao tối ưu? Lệnh UPDATE này sẽ tự động kích hoạt cơ chế khóa ghi (Exclusive Lock - FOR UPDATE ngầm định) lên các dòng ghế trong Postgres. Các transaction đồng thời khác muốn ghi đè lên các ghế này sẽ phải xếp hàng chờ.
Nếu số dòng bị ảnh hưởng (RowsAffected) nhỏ hơn số ghế cần đặt, nghĩa là đã có ít nhất một ghế bị người khác đặt trước (trạng thái không còn là AVAILABLE). Khi đó ta chủ động trả về lỗi và Postgres tự động rollback toàn bộ transaction.
Khả năng chịu lỗi (Fault tolerance): Nếu xóa khóa Redis ở cuối luồng thất bại (ví dụ: mất kết nối mạng với Redis), hệ thống vẫn hoạt động chính xác vì TTL của Redis sẽ tự động hết hạn sau 5 phút, và trạng thái trong Postgres lúc này đã được cập nhật thành BOOKED (không ai có thể đặt lại ghế này nữa).