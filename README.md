# Shundao Solar (Fusion-Shundao) Version 2.5

> **Tài liệu chính thức mô tả chi tiết kiến trúc, quy trình và cách vận hành dự án Shundao Solar SCADA. Cập nhật mới nhất: Version 2.5.**

---

## 1. Tổng quan hệ thống
**Shundao Solar** là một giải pháp giám sát, điều khiển và thu thập dữ liệu (SCADA/IoT) toàn diện. Do hệ thống gốc Huawei FusionSolar không có public API, hệ thống này đóng vai trò là một cầu nối tự động trích xuất, chuẩn hóa, và hiển thị lại dữ liệu năng lượng mặt trời theo định hướng Real-time tối đa hóa lợi ích cho người vận hành.

---

## 2. Kiến trúc & Ngăn xếp Công nghệ (Tech Stack)

Hệ thống được chia làm hai phần chính với kiến trúc **Clean Architecture** mạnh mẽ:

*   **Backend (Golang):** Hiệu năng cao vượt trội, xử lý Concurrency (Goroutines) cho quá trình Crawl dữ liệu và đẩy real-time (`gorilla/websocket` thông qua Server-Sent Events - SSE). Tích hợp cấu hình đo lường rò rỉ bộ nhớ (Memory leak) với `pprof` trên port `6060`.
*   **Database (Time-Series + RDBMS):** 
    *   **VictoriaMetrics**: Lưu trữ hàng triệu điểm dữ liệu (metrics) một cách mượt mà và nén siêu việt.
    *   **PostgreSQL**: Lưu trữ trạng thái tài khoản, thiết đặt, cảnh báo và logs hệ thống.
*   **Frontend (React + Vite):** Xây dựng với phong cách glassmorphism, được tối ưu tốc độ bằng kỹ thuật Lazy Loading/Code-Splitting, `lucide-react`, TailwindCSS.
*   **Triển khai (Deployment):** Đóng gói 100% bằng **Docker** + **NGINX**.

---

## 3. Quy trình Trích xuất & Hệ thống Cảnh báo (Data Pipeline)

Hệ thống hoạt động liên tục 24/7 theo chu trình sau:

### 3.1. Crawler (Trích xuất ngầm)
- Mô-đun `internal/browser` dùng **Chromedp (Headless Browser)** ngầm mở giao diện của FusionSolar.
- Tự động điền mật khẩu, quét toàn bộ cây thiết bị (SmartLogger -> Inverter -> Chuỗi PV).
- Phân tách HTML DOM lấy số liệu V, A, kW trong vòng vài giây mà không bị phát hiện.

### 3.2. Real-time Streaming (SSE)
- Sau khi Crawler lưu xong dữ liệu thô vào VictoriaMetrics, Backend sẽ đóng gói payload (JSON) và gửi Push Notification (1 chiều) về trình duyệt web thông qua cổng `/api/stream/dashboard`.
- Độ trễ từ khi lấy dữ liệu cục bộ đến khi Frontend nhảy số chỉ tốn **dưới 1 giây**. Trình duyệt hoàn toàn dập tắt các Request Polling lãng phí.

### 3.3. Thuật toán Smart Alerts (Khử nhiễu Cảnh báo giả)
- Thay vì phụ thuộc cảnh báo thô từ Huawei (thường xuyên báo láo do tắt chuỗi cắm), Frontend tự động phân tích:
    - Bỏ qua toàn bộ các String được cấu hình "Loại trừ" (Ví dụ: `1,2,7,8`).
    - **Debounce 15 phút**: Mất điện hoặc tụt dòng tạm thời (VD: Đám mây bay qua) sẽ không làm chớp đỏ. Chỉ khi vượt quá 15 phút, lỗi mới bị "Confirmed" và kích hoạt âm thanh cảnh báo.
    - Cảnh báo truy vết tận gốc: `Tên Trạm Logger - Tên Inverter` hiển thị rõ ràng.

---

## 4. Hướng dẫn Biên dịch & Chạy (Build & Run)

### 4.1. Môi trường kiểm thử Local (Development)
Chạy tự động bằng kịch bản Powershell đã cung cấp sẵn (yêu cầu Windows + cài sẵn Go và Nodejs):
```powershell
.\run_all.ps1
```
*Lệnh này sẽ dọn dẹp port tĩnh, bật Backend API (port 5040), bật Backend Legacy (port 5039) và Frontend Vite tự động mở trình duyệt.*

### 4.2. Khởi tạo bản Build Production (Docker Tarball)
Để xuất bản ra Server thực tế (Linux Ubuntu) mà không cần cài mã nguồn hay cài Node/Go trên Server, chạy lệnh sau tại máy Dev:
```powershell
.\build_export_docker.ps1
```
*Quá trình này sẽ:*
1. Build `fusion-backend:latest` (Alpine OS, siêu nhẹ).
2. Build `fusion-frontend:latest` (NGINX tĩnh, trỏ reverse proxy qua cổng 5040).
3. Đóng gói 2 file Image thành `shundao_production.tar` nén bên trong thư mục `shundao_deploy_package/` cùng với file định nghĩa `docker-compose.prod.yml`.

### 4.3. Đưa lên Server (Linux Environment)
Copy toàn bộ thư mục `shundao_deploy_package` lên máy chủ Ubuntu thông qua SFTP và SSH vào chạy:
```bash
# 1. Giải nén các file image vào Docker cục bộ
docker load -i shundao_production.tar

# 2. Sinh cấu hình môi trường gốc
cp .env.example .env

# 3. Kích hoạt toàn bộ hệ sinh thái (Frontend + Backend) ở chế độ mờ (Detached)
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

---

## 5. Mã lỗi & Ghi log Bảo mật (Error Codes & Security)

Dự án có cơ chế bảo vệ phân tầng chuyên sâu. Bạn sẽ thấy các mã lỗi HTTP sau:

| Mã HTTP | Hiện tượng | Nguyên nhân xử lý |
| :--- | :--- | :--- |
| **401 Unauthorized** | Token hết hạn / Lỗi Xác Thực | Tự động Kick về Login. Crawler đằng sau nếu bị (từ Huawei) sẽ tự kích hoạt luồng đăng nhập làm mới (`Self-healing`). |
| **403 Forbidden** | Khóa theo khung giờ | Người dùng cố tình đăng nhập ngoài giờ hành chính (SOLAR_START_HOUR - SOLAR_END_HOUR) chưa được phân quyền khẩn cấp. |
| **418 I'm a teapot** | Chuyển hướng "Lockdown" đỏ loét | (Đặc quyền Counter-Hack) Hệ thống phát hiện Hacker cố tình quét API, Request rác, SQL Injection -> Bẫy Honey-pot lập tức khóa chặn IP / Session. Dọa bằng hiệu ứng Matrix Hacker trên trình duyệt. |
| **429 Too Many Requests** | Khóa IP Brute-Force | Kẻ tấn công cố tình rải Password vào form Đăng nhập. Hệ thống dùng Exponential Backoff -> 10s khóa -> 20s khóa -> 2h khóa tài khoản tạm thời. |

*Mọi log vận hành (Truy xuất DB, Cảnh báo Lỗi, Crash) được xử lý bằng `Lumberjack` xoay log định kỳ tại thư mục `backend/logs/` nhằm tránh tình trạng đầy ổ cứng máy chủ sau 5 năm.*

---

## 6. Sức khỏe Nội tại Hệ thống (Profiling)
Kiểm tra rò rỉ RAM (OOM) / Thống kê cấu trúc Heap bằng cách SSH vào giao diện ảo của Backend (Mặc định Port ẩn `6060`):
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```
Màn hình console của bộ pprof sẽ giúp kĩ sư soi kĩ lượng Bytes phân bổ của từng hàm khi chạy trên Server thật.

--- 

> **Phát triển bởi**: Phạm Phúc 
> **Phiên bản kiến trúc hiện tại**: 2.5 (Năm 2026)
