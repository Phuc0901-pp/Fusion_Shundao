# Hướng dẫn Triển khai trên Linux

Tài liệu này hướng dẫn cách triển khai ứng dụng Fusion Shundao lên máy chủ Linux sử dụng Docker và Ngrok.

## 1. Chuẩn bị

Đảm bảo máy chủ Linux của bạn đã cài đặt:
- **Docker**: [Hướng dẫn cài đặt](https://docs.docker.com/engine/install/)
- **Docker Compose**: [Hướng dẫn cài đặt](https://docs.docker.com/compose/install/)
- **Git** (tùy chọn, để clone code)

## 2. Sao chép mã nguồn

Bạn cần sao chép toàn bộ thư mục dự án lên máy chủ Linux. Có 2 cách:

### Cách 1: Sử dụng Git (Khuyên dùng)
Nếu bạn đã đẩy code lên GitHub, hãy clone về máy chủ:
```bash
git clone https://github.com/hoaibao0993-dev/Fusion-Shundao.git
cd Fusion-Shundao
git checkout backup-2026-02-09  # Hoặc branch mới nhất của bạn
```

### Cách 2: Triển khai Offline (Tối ưu dung lượng)
Đây là cách **nhanh nhất và nhẹ nhất**, không cần copy source code, không cần internet mạnh trên server.

#### Bước 1: Tại máy Windows của bạn
1.  Chạy script đóng gói:
    ```powershell
    ./scripts/pack_deploy.ps1
    ```
    Script này sẽ tạo ra file `fusion_images.tar` chứa tất cả những gì cần thiết.

#### Bước 2: Copy lên Server
Bạn chỉ cần copy 2 file này lên Linux (chung 1 thư mục):
1.  `fusion_images.tar`
2.  `deployments/docker-compose.prod.yml`
3.  Thư mục `configs/` (Chứa `app.json` và `signals.json`)

#### Bước 3: Tại Server Linux
Chạy các lệnh sau để nạp image và khởi động:

```bash
# Nạp image từ file
docker load -i fusion_images.tar

# Sửa file docker-compose.yml một chút để dùng image có sẵn
# (Xem phần Cấu hình bên dưới)

# Khởi chạy
docker-compose up -d
```

### Cách 3: Copy file thủ công (Toàn bộ source)
Copy toàn bộ thư mục `Fusion-Shundao` lên máy chủ.

## 3. Cấu hình

Nếu dùng **Cách 2 (Offline)**, bạn hãy sử dụng file `deployments/docker-compose.prod.yml` mà tôi đã tạo sẵn (đã cấu hình dùng image thay vì build).

Khi copy lên server, bạn đổi tên nó thành `docker-compose.yml` cho tiện:
```bash
mv docker-compose.prod.yml docker-compose.yml
```
Hoặc chạy trực tiếp với flag `-f`:
```bash
docker-compose -f docker-compose.prod.yml up -d
```

Kiểm tra file `deployments/docker-compose.yml` để đảm bảo token Ngrok đã chính xác (đã được cấu hình sẵn trong code).

## 4. Khởi chạy Ứng dụng

Di chuyển vào thư mục `deployments` và chạy lệnh sau:

```bash
cd deployments

# Dừng các container cũ nếu đang chạy
docker-compose down

# Build và chạy container ở chế độ nền (detached mode)
docker-compose up -d --build
```

## 5. Kiểm tra và Truy cập

### Xem trạng thái container
```bash
docker-compose ps
```
Bạn sẽ thấy:
- `shundao-frontend` chạy trên port `0.0.0.0:9005->80/tcp`
- `shundao-backend` (không public port, chỉ giao tiếp nội bộ)
- `shundao-ngrok`

### Truy cập Local
Truy cập qua trình duyệt tại máy chủ: `http://localhost:9005`

### Lấy link Public từ Ngrok
Để lấy đường dẫn truy cập công khai, hãy xem log của container ngrok:

```bash
docker-compose logs ngrok
```
Tìm dòng có dạng `url=https://xxxx-xxxx.app`. Đó chính là địa chỉ truy cập web của bạn.

### Kiểm tra Logs
Nếu cần debug, bạn có thể xem log của từng dịch vụ:
```bash
docker-compose logs -f backend
docker-compose logs -f frontend
```
