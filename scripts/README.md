# 🛠️ Hướng dẫn sử dụng các Script bổ trợ dưới Local

Thư mục này chứa các file script bổ trợ cho việc xây dựng (build), phát triển và dọn dẹp hệ thống dưới môi trường local Docker Swarm.

---

## 🚀 1. Script tự động build & deploy nhanh (`deploy_local.ps1`)

*   **Đường dẫn:** `scripts/deploy_local.ps1`
*   **Chức năng:** Tự động biên dịch chéo Go sang môi trường Linux (`amd64`), đóng gói Docker image dưới máy của bạn và đẩy lên local Registry trên VM Swarm Manager, sau đó kích hoạt deploy stack.
*   **Cách sử dụng:**
    Mở Terminal dưới máy host Windows của bạn và chạy:
    ```bash
    make deploy-local-fast
    ```
    *Hoặc chạy thủ công qua PowerShell:*
    ```powershell
    powershell -ExecutionPolicy Bypass -File scripts/deploy_local.ps1
    ```

---

## 🧹 2. Script dọn dẹp đĩa cứng cụm Swarm (`run_prune.sh`)

*   **Đường dẫn:** `scripts/run_prune.sh`
*   **Chức năng:** Khởi chạy một dịch vụ Swarm toàn cục (Global Service) kết nối vào `/var/run/docker.sock` trên từng node nhằm tự động dọn dẹp các container đã dừng, image mồ côi và cache build để tránh lỗi hết đĩa (`No space left on device`).
*   **Cách sử dụng:**
    Chạy trực tiếp từ máy host Windows bằng cách đẩy script qua kết nối SSH:
    ```bash
    make prune-swarm
    ```
    *Lệnh này sẽ tự động nạp nội dung file `scripts/run_prune.sh` từ máy bạn sang shell trên VM Swarm Manager để thực thi.*
