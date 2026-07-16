# Hướng Dẫn Triển Khai Docker Swarm - Môi Trường Production

Tài liệu này hướng dẫn chi tiết các bước thiết lập và khởi chạy hạ tầng (Traefik, Portainer, Registry) trên môi trường Production sử dụng cụm Docker Swarm và chứng chỉ SSL Let's Encrypt tự động.

---

## 📋 Điều kiện tiên quyết (Prerequisites)

Trước khi bắt đầu, hãy đảm bảo các yêu cầu sau đã được cấu hình:
1. **DNS Records**: Đã trỏ các tên miền phụ (A Records) về địa chỉ IP Public của Swarm Manager VPS:
   - `traefik.yourdomain.com` ➔ `IP_PUBLIC_MANAGER`
   - `portainer.yourdomain.com` ➔ `IP_PUBLIC_MANAGER`
   - `registry.yourdomain.com` ➔ `IP_PUBLIC_MANAGER`
   - `api.yourdomain.com` ➔ `IP_PUBLIC_MANAGER`
2. **Security Groups / Firewall**: Đã mở các cổng kết nối trên VPS:
   - `80/TCP` (HTTP) và `443/TCP` (HTTPS) - Cửa ngõ của Traefik.
   - `9001/TCP` - Cổng kết nối Agent của Portainer giữa các node.
   - `2377/TCP`, `7946/TCP/UDP`, `4789/UDP` - Các cổng phục vụ kết nối Docker Swarm.
3. **Swarm Mode**: Đã khởi tạo cụm Swarm (`docker swarm init`) và kết nối các Worker Node (`docker swarm join`).

---

## 🚀 Các bước triển khai (Deployment Steps)

Thực hiện lần lượt các bước dưới đây trên **Swarm Manager Node**:

### Bước 1: Tạo mạng Overlay dùng chung
Mạng overlay này cho phép Traefik giao tiếp bảo mật với Portainer và các microservices khác chạy ở các node khác nhau trong cụm.

```bash
docker network create --driver overlay --attachable traefik-public
```

---

### Bước 2: Khởi chạy Traefik làm API Gateway & SSL Resolver
Traefik sẽ đảm nhận việc đón traffic ở cổng 80/443, tự động cấp phát SSL và chuyển hướng HTTP ➔ HTTPS.

1. Chuyển vào thư mục chứa file cấu hình production:
   ```bash
   cd deployment/production
   ```
2. Cập nhật email quản trị của em trong file `traefik-stack.yml` tại dòng:
   `--certificatesresolvers.productionresolver.acme.email=admin@yourdomain.com`
3. Deploy Traefik Stack:
   ```bash
   docker stack deploy -c traefik-stack.yml traefik
   ```
4. Kiểm tra trạng thái khởi động của Traefik:
   ```bash
   docker stack ps traefik
   ```
   *Đợi khoảng 30s-1 phút để Traefik tải ảnh và tự tạo file lưu trữ chứng chỉ `/letsencrypt/acme.json`.*

---

### Bước 3: Khởi chạy Portainer để quản trị cụm Swarm
Portainer Agent sẽ chạy trên mọi node để thu thập thông tin, còn Portainer UI sẽ chạy trên Manager Node và đi qua Gateway Traefik.

1. Deploy Portainer Stack:
   ```bash
   docker stack deploy -c portainer-agent-stack.yml portainer
   ```
2. Kiểm tra tiến trình khởi chạy:
   ```bash
   docker stack ps portainer
   ```
3. Truy cập vào trang quản trị thông qua giao diện bảo mật HTTPS:
   👉 **`https://portainer.yourdomain.com`**
   *(Lần đầu truy cập, Portainer sẽ yêu cầu em tạo tài khoản Admin và mật khẩu).*

---

### Bước 4: Kiểm tra trạng thái SSL và Logs (Troubleshooting)

Nếu em không thể truy cập qua HTTPS hoặc bị lỗi kết nối, hãy chạy các lệnh sau trên máy ảo Manager để kiểm tra logs:

* **Xem logs của Traefik**:
  ```bash
  docker service logs traefik_traefik -f
  ```
  *Tìm kiếm các dòng có chữ `ACME` hoặc `Let's Encrypt` để xem tiến trình cấp chứng chỉ SSL có thành công hay không.*

* **Xem danh sách dịch vụ đang chạy**:
  ```bash
  docker service ls
  ```

---

## 🔒 Khuyến nghị bảo mật cho Production

1. **Bảo mật Traefik Dashboard**: Mặc định dashboard của Traefik trong file `traefik-stack.yml` đã được tắt chế độ `insecure`. Nếu muốn bật lên để theo dõi qua tên miền, hãy uncomment các dòng Basic Auth ở cuối file `traefik-stack.yml` để tránh rò rỉ thông tin hạ tầng.
2. **Giới hạn quyền truy cập Registry**: Khi cấu hình Private Registry ở production, hãy tích hợp thêm cơ chế xác thực Basic Auth để tránh người lạ tự ý tải hoặc đẩy image lên cluster của em.
