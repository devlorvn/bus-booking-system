oke giờ tôi đang có ý tưởng sẽ làm 1 hệ thống đặt vé xe đơn giản để học về 
golang, websocket và grpc hoặc kafka, 
không cần xác thực đăng kí gì về user, giao diện đơn giản đủ để test, 
+ có một chức năng tạo mở vé xe sẽ có id, biển số xe, ngày đi, nơi đi nơi đến, số vé, số vé trống
+ giao diện sẽ bấm vào chọn 1 xe để chọn vị trí, mỗi xe sẽ có 4 hàng ghế chi đều là A,B,C,D, tuỳ vào số ghế mà chia đều
+ sau khi chọn xong thì đến bước xác nhận cần các thông tin cơ bản của khách hàng
Bên trên là ý tưởng về chức năng giao diện, phần chính sẽ là API
- tôi sẽ làm theo clean arch
- các domain cơ bản tôi xác định gồm có: Bus, Booking, User như sau
    - Bus
        - id
        - 
    - User
        - ID
        - Name
        - Email
        - PhoneNumber
        - LastActive
    - Booking
        - ID
        - BusID
        - SeatID
        - UserId
        - Status
- technical gồm có golang là ngôn ngữ chính, xử lý rest api gằng gin framework, redis và websocket để xử lý đặt vé,
postgreSQL để store data, bạn có thể gợi ý cho tôi nên sài gRPC hay Kafka cho case này để tách riêng các service,
thứ tôi muốn học ở case này ngoài các technical cần có asynchonus transaction, và consistent data
- Hãy lên kế hoạch lại cho dư án này, phân tích tổ chức thành một file readme để tôi làm mô tả cho dự án, bạn có thể đề xuất các usecase cho dự án này, nếu có chỗ nào chưa hiểu hoặc cần làm rõ, hãy hỏi lại tôi để chúng ta đưa ra giải pháp
