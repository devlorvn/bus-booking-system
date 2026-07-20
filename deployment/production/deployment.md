# Hướng Dẫn Triển Khai Docker Swarm - Môi Trường Production & Monitoring

Tài liệu này hướng dẫn chi tiết toàn bộ các bước thiết lập, đóng gói, phân phối và khởi chạy hạ tầng ứng dụng **Bus Booking System** trên môi trường Production sử dụng cụm **Docker Swarm**. Hạ tầng bao gồm hệ thống định tuyến (Traefik), quản trị (Portainer), giám sát logs/metrics (Grafana, Prometheus, Loki, Promtail) và các microservices của ứng dụng.

---

## 🏗️ Tổng Quan Kiến Trúc Hạ Tầng (Architecture Overview)

Kiến trúc hạ tầng trên cụm Docker Swarm được thiết lập như sau:
*   **API Gateway & SSL**: [Traefik](file:///c:/Users/luan.nguyen/Documents/devlor/bus-booking-system/deployment/production/traefik-stack.yml) hứng toàn bộ traffic (HTTP/HTTPS), tự động giải quyết chứng chỉ Let's Encrypt SSL và cân bằng tải tới các service phía sau.
*   **Quản trị cụm**: [Portainer](file:///c:/Users/luan.nguyen/Documents/devlor/bus-booking-system/deployment/production/portainer-agent-stack.yml) chạy Agent trên toàn bộ các node, quản trị tập trung qua giao diện Web của Node Manager.
*   **Hệ thống Monitoring & Logging**:
    *   **Prometheus**: Thu thập metrics (CPU, RAM, HTTP Requests, DB Connection Pool...) từ các microservices qua các cổng metrics riêng biệt (9091-9095).
    *   **Loki**: Cơ sở dữ liệu tập trung lưu trữ logs.
    *   **Promtail**: Chạy dạng `global` trên mọi node để bóc tách logs từ thư mục `/var/lib/docker/containers/` và gửi về Loki.
    *   **Grafana**: Giao diện trực quan hóa dữ liệu logs và metrics.
*   **Ứng dụng (App Stack)**: Chạy 10 services bao gồm 7 microservices viết bằng Go, 1 Postgres database, 1 Redis cache, và 1 Kafka message broker.

---

## 📋 Điều kiện tiên quyết (Prerequisites)

Trước khi bắt đầu, hãy đảm bảo các yêu cầu sau đã được cấu hình:
1.  **Cấu hình DNS (A Records)**: Đã trỏ các tên miền phụ về địa chỉ IP Public của Swarm Manager VPS:
    *   `traefik.yourdomain.com` ➔ `IP_PUBLIC_MANAGER`
    *   `portainer.yourdomain.com` ➔ `IP_PUBLIC_MANAGER`
    *   `registry.yourdomain.com` ➔ `IP_PUBLIC_MANAGER` (Nếu dùng Private Registry tự dựng)
    *   `api.yourdomain.com` ➔ `IP_PUBLIC_MANAGER` (Dành cho API Gateway ứng dụng)
    *   `grafana.yourdomain.com` ➔ `IP_PUBLIC_MANAGER` (Dành cho giám sát hệ thống)
2.  **Mở cổng Firewall / Security Groups**:
    *   `80/TCP` (HTTP) và `443/TCP` (HTTPS) - Cổng vào của Traefik.
    *   `9001/TCP` - Cổng kết nối agent của Portainer giữa các node.
    *   `2377/TCP`, `7946/TCP/UDP`, `4789/UDP` - Các cổng giao tiếp nội bộ của cụm Docker Swarm.
3.  **Khởi tạo Swarm Cluster**:
    *   Trên node Manager: `docker swarm init --advertise-addr <IP_PUBLIC_MANAGER>`
    *   Trên các node Worker: Chạy lệnh `docker swarm join` được cấp bởi node Manager.

---

## 🚀 Các Bước Triển Khai Từng Bước (Step-by-Step Deployment)

Toàn bộ các lệnh dưới đây được thực hiện trực tiếp trên máy ảo **Swarm Manager Node** (sau khi chuyển vào thư mục `deployment/production`):

```bash
cd deployment/production
```

### Bước 1: Tạo các mạng Overlay dùng chung
Các mạng này giúp định tuyến lưu lượng giữa hệ thống cổng vào (Traefik), hệ thống giám sát (Monitoring) và các microservices chạy trên các node khác nhau.

```bash
# Mạng dành cho Traefik giao tiếp với các services công cộng (Portainer, API, Web)
docker network create --driver overlay --attachable traefik-public

# Mạng nội bộ dành cho các microservices kết nối chéo và kết nối database/broker
docker network create --driver overlay --attachable app-network
```

---

### Bước 2: Triển khai Traefik (API Gateway & SSL Resolver)
Traefik tự động quét các cấu hình nhãn (labels) trên các container và tự sinh chứng chỉ Let's Encrypt SSL.

1.  Mở file [traefik-stack.yml](file:///c:/Users/luan.nguyen/Documents/devlor/bus-booking-system/deployment/production/traefik-stack.yml) và cấu hình lại email quản trị (dòng 21) để Let's Encrypt gửi thông báo khi có sự cố.
2.  Deploy Traefik Stack:
    ```bash
    docker stack deploy -c traefik-stack.yml traefik
    ```
3.  Kiểm tra trạng thái khởi động:
    ```bash
    docker stack ps traefik
    ```

---

### Bước 3: Triển khai Portainer (Quản trị cụm Swarm trực quan)
Portainer Agent chạy dưới dạng `global` (mỗi node 1 instance) để lắng nghe sự kiện của Docker daemon, giao tiếp trực tiếp với Portainer Server chính chạy tại Manager Node.

1.  Deploy Portainer Stack:
    ```bash
    docker stack deploy -c portainer-agent-stack.yml portainer
    ```
2.  Kiểm tra dịch vụ:
    ```bash
    docker stack ps portainer
    ```
3.  Truy cập qua HTTPS: **`https://portainer.yourdomain.com`** và tiến hành khởi tạo tài khoản Admin.

---

### Bước 4: Triển khai Private Docker Registry (Tùy chọn)
Để phân phối các image ứng dụng tự build sang các node Worker khác trong cụm Swarm:

1.  Có thể tận dụng registry local chạy trên cổng `5000` của node Manager bằng cách chạy stack registry hoặc dùng dịch vụ Docker Hub / Github Container Registry (GHCR).
2.  Nếu dùng registry nội bộ của host:
    ```bash
    docker service create --name registry --publish published=5000,target=5000 --repository registry:2
    ```

---

### Bước 5: Triển khai Monitoring Stack (Grafana, Prometheus, Loki, Promtail)
Đây là hệ thống thu thập metrics và logs tập trung. **Lưu ý**: Do stack này liên kết trực tiếp với các tệp cấu hình trên đĩa thông qua cơ chế mount, bạn bắt buộc phải deploy từ đúng thư mục `deployment/production/`.

1.  Kiểm tra và cấu hình các tệp cấu hình:
    *   [prometheus.yml](file:///c:/Users/luan.nguyen/Documents/devlor/bus-booking-system/deployment/production/prometheus.yml): Cấu hình các target thu thập dữ liệu (cổng 9091 đến 9095 tương ứng với từng service).
    *   [promtail-config.yml](file:///c:/Users/luan.nguyen/Documents/devlor/bus-booking-system/deployment/production/promtail-config.yml): Định nghĩa đường dẫn logs container cần thu thập và đẩy về Loki.
2.  Deploy Monitoring Stack:
    ```bash
    docker stack deploy -c monitoring-stack.yml monitoring
    ```
3.  Kiểm tra trạng thái các service:
    ```bash
    docker stack ps monitoring
    ```
4.  **Truy cập Grafana**:
    *   Mở trình duyệt: `http://grafana.local` (hoặc thông qua tên miền trỏ về VPS với nhãn cấu hình Traefik).
    *   Tài khoản mặc định: `admin` / Mật khẩu: `admin` (Hãy đổi mật khẩu ngay lần đầu đăng nhập).
    *   **Thêm Data Sources**:
        *   **Prometheus**: Điền URL `http://prometheus:9090`
        *   **Loki**: Điền URL `http://loki:3100`

---

### Bước 6: Đóng gói và Triển khai Application Stack (Microservices)

1.  **Build và Push Image lên Registry**:
    Trước khi deploy ứng dụng, các image ứng dụng cần được build và đẩy lên Docker Registry mà cụm Swarm có thể truy cập được. Bạn có thể sử dụng CI/CD hoặc chạy script build thủ công:
    ```bash
    # Ví dụ build và push local registry
    export DOCKER_REGISTRY=192.168.220.128:5000
    export TAG=v1.1.0

    docker build --build-arg SERVICE_NAME=user-service -t $DOCKER_REGISTRY/bus-booking-user-service:$TAG .
    docker push $DOCKER_REGISTRY/bus-booking-user-service:$TAG
    # Lặp lại tương tự cho các service khác: bus-service, booking-service, payment-service, notification-service, ws-service, api.
    ```

2.  **Triển khai App Stack**:
    Cung cấp các biến môi trường cần thiết và thực hiện deploy:
    ```bash
    export DOCKER_REGISTRY=192.168.220.128:5000
    export TAG=v1.1.0
    export API_DOMAIN=api.yourdomain.com

    docker stack deploy -c app-stack.yml bus-booking
    ```

3.  **Xác nhận dịch vụ hội tụ (Service Convergence)**:
    Do các service sử dụng cơ chế kết nối cơ sở dữ liệu có cơ chế tự động thử lại (Retry Connection Logic) trong Go, chúng sẽ tự động chờ Postgres/Kafka khởi chạy hoàn tất rồi tự động migration schema database. Kiểm tra trạng thái:
    ```bash
    docker service ls
    docker stack ps bus-booking
    ```

---

## 🛠️ Khắc Phục Sự Cố & Bảo Trì Đĩa Hệ Thống (Maintenance & Disk Cleanup)

Trong môi trường Docker Swarm, tình trạng đầy đĩa hệ thống (100% disk space) trên các Worker Node là cực kỳ phổ biến do lượng logs container dồn ứ và cache build. Khi đĩa đầy, dịch vụ cơ sở dữ liệu (Postgres) sẽ bị crash với lỗi `No space left on device` (không thể ghi file giao dịch ghi trước `pg_wal`).

### 1. Cơ chế tự động dọn dẹp hàng tuần (Cronjob Prune)
Nên thiết lập dọn dẹp định kỳ trên tất cả các Node (cả Manager và Worker).

Tạo một tệp script dọn dẹp `/usr/local/bin/docker-cleanup.sh` trên mỗi node:
```bash
#!/bin/bash
# Dọn dẹp container đã dừng, volume mồ côi và cache build dư thừa
docker system prune -a --volumes -f
docker builder prune -a -f
```
Cấp quyền thực thi và tạo cấu hình cronjob chạy vào 2:00 sáng chủ nhật hàng tuần:
```bash
chmod +x /usr/local/bin/docker-cleanup.sh
echo "0 2 * * 0 /usr/local/bin/docker-cleanup.sh > /dev/null 2>&1" | crontab -
```

### 2. Xử lý khẩn cấp khi đĩa đầy 100% (Phá khóa Deadlock)
Khi đĩa đầy 100%, bạn không thể chạy các container dọn dẹp thông thường (như cài `apk add docker-cli` để prune) vì hệ thống không thể ghi file tạm.

Để xử lý, bạn có thể triển khai một script Perl chạy trên container base có sẵn (ví dụ image `postgres:16` đã kéo sẵn trên node) để giao tiếp trực tiếp với cổng socket của Docker thông qua giao thức raw HTTP:

1.  Tạo tệp cấu hình script Perl [prune.pl](file:///c:/Users/luan.nguyen/Documents/devlor/bus-booking-system/deployment/production/prune.pl) trên Manager:
    ```perl
    use IO::Socket::UNIX;
    my $socket = IO::Socket::UNIX->new(Type => SOCK_STREAM, Peer => '/var/run/docker.sock') or die $!;
    # Gọi trực tiếp API Prune của Docker Engine qua socket
    print $socket "POST /v1.41/containers/prune?force=true HTTP/1.0\r\n\r\n";
    print $socket "POST /v1.41/images/prune?all=true HTTP/1.0\r\n\r\n";
    print $socket "POST /v1.41/volumes/prune HTTP/1.0\r\n\r\n";
    while (<$socket>) { print $_; }
    ```
2.  Tạo Docker Config và deploy một service khẩn cấp chạy trên node bị đầy đĩa (ví dụ `worker2`):
    ```bash
    # Khởi tạo config
    docker config create prune_pl prune.pl
    
    # Chạy service một lần để dọn dẹp
    docker service create --name prune-worker --restart-condition none \
      --constraint node.hostname==worker2 \
      --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
      --config source=prune_pl,target=/prune.pl \
      postgres:16 perl /prune.pl
    ```
3.  Sau khi dọn dẹp xong, xóa service và config để giải phóng tài nguyên:
    ```bash
    docker service rm prune-worker
    docker config rm prune_pl
    ```
