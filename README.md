# Shundao Solar — Fusion-Shundao v2.5

> **Tài liệu vận hành chính thức** — Mô tả đầy đủ kiến trúc, quy trình khởi chạy, hướng dẫn sử dụng, mã lỗi và phương pháp xử lý sự cố của hệ thống giám sát năng lượng mặt trời Shundao Solar SCADA.
> Cập nhật: **Tháng 4 / 2026 — Version 2.5**

---

## Mục lục

1. [Tổng quan hệ thống](#1-tổng-quan-hệ-thống)
2. [Kiến trúc & Tech Stack](#2-kiến-trúc--tech-stack)
3. [Cấu trúc thư mục dự án](#3-cấu-trúc-thư-mục-dự-án)
4. [Cấu hình môi trường](#4-cấu-hình-môi-trường)
5. [Quy trình khởi chạy (Local Dev)](#5-quy-trình-khởi-chạy-local-dev)
6. [Quy trình Build & Triển khai Production (Docker)](#6-quy-trình-build--triển-khai-production-docker)
7. [Hướng dẫn sử dụng giao diện](#7-hướng-dẫn-sử-dụng-giao-diện)
8. [Luồng dữ liệu & Data Pipeline](#8-luồng-dữ-liệu--data-pipeline)
9. [Mã lỗi & Phương pháp xử lý](#9-mã-lỗi--phương-pháp-xử-lý)
10. [Giám sát sức khỏe hệ thống](#10-giám-sát-sức-khỏe-hệ-thống)
11. [Câu hỏi thường gặp (FAQ)](#11-câu-hỏi-thường-gặp-faq)

---

## 1. Tổng quan hệ thống

**Shundao Solar** là hệ thống SCADA/IoT giám sát toàn diện hai nhà máy năng lượng mặt trời:
- **SHUNDAO 1** — `NE=50143101`
- **SHUNDAO 2** — `NE=50127734`

Do nền tảng gốc **Huawei FusionSolar** không cung cấp Public API, hệ thống đóng vai trò là **cầu nối thông minh**: tự động trích xuất, chuẩn hóa và hiển thị dữ liệu theo thời gian thực thông qua giao diện nội bộ riêng.

**Các chỉ số theo dõi chính:**
- Công suất phát điện thực tế (Active Power — kW)
- Sản lượng trong ngày / tháng / tổng tích lũy (Yield — kWh)
- Cường độ bức xạ mặt trời (Irradiance — W/m²)
- Điện áp / Dòng điện từng chuỗi pin (String V/A)
- Trạng thái inverter (Online / Offline / Fault)

---

## 2. Kiến trúc & Tech Stack

Hệ thống áp dụng **Clean Architecture** với 3 lớp tách biệt hoàn toàn:

```
┌─────────────────────────────────────────────────────────┐
│  FRONTEND (React 19 + Vite)   ←─── Port 2608 / 9000     │
│  [Glassmorphism UI, TailwindCSS, Recharts, Zustand]      │
└─────────────────────┬───────────────────────────────────┘
                      │  REST API + SSE (Server-Sent Events)
┌─────────────────────▼───────────────────────────────────┐
│  BACKEND API (Go / Golang)                               │
│  Port 5040 — Clean Architecture Routes                   │
│  Port 5039 — Legacy Routes (Headless Crawl & SSE)        │
│  Port 6060 — pprof (Memory Profiling, ẩn)                │
│  [net/http, chromedp, gorilla/websocket, JWT, bcrypt]    │
└──────────────────┬──────────────────┬───────────────────┘
                   │                  │
    ┌──────────────▼──┐    ┌──────────▼──────────────────┐
    │  PostgreSQL      │    │  VictoriaMetrics             │
    │  Port 2602       │    │  Port 8428                   │
    │  Tài khoản,      │    │  Time-series metrics:        │
    │  Cấu hình,       │    │  Inverter, String, KPI       │
    │  Cảnh báo logs   │    │  (triệu điểm / ngày)         │
    └──────────────────┘    └─────────────────────────────┘
```

| Lớp | Công nghệ | Mục đích |
|---|---|---|
| Frontend | React 19, Vite, TailwindCSS, Recharts, Zustand | Giao diện realtime, biểu đồ |
| Backend | Go (Golang), net/http, chromedp | API server, Crawler Headless |
| Time-series DB | VictoriaMetrics | Lưu trữ metrics theo thời gian |
| Relational DB | PostgreSQL (qua GORM) | Tài khoản, config, log |
| Auth | JWT (HS256), bcrypt | Bảo mật & xác thực |
| Log xoay vòng | Lumberjack | Tránh đầy disk dài hạn |
| Container | Docker + NGINX | Triển khai, reverse proxy |

---

## 3. Cấu trúc thư mục dự án

```
Shundao/
├── backend/                    # Go source code
│   ├── cmd/server/main.go      # Server entrypoint
│   ├── internal/
│   │   ├── api/                # Crawl API: fetcher.go, kpi_fetcher.go, ...
│   │   ├── browser/            # Chromedp Headless browser automation
│   │   ├── core/               # Usecase & business logic
│   │   ├── database/           # Models PostgreSQL (Account, ...)
│   │   ├── delivery/http/      # HTTP handlers & router (Clean Arch)
│   │   ├── domain/             # Domain entities
│   │   ├── errors/             # Định nghĩa AppError, mã lỗi
│   │   ├── middleware/         # Rate limiter, CORS
│   │   ├── notify/             # Lark notification
│   │   ├── platform/           # Logger, utils
│   │   ├── repository/         # VictoriaMetrics & Postgres repo
│   │   ├── usecase/            # Dashboard usecase
│   │   └── victoriametrics/    # VM client
│   ├── Dockerfile
│   └── go.mod
│
├── frontend/                   # React/Vite source code
│   └── src/
│
├── configs/
│   ├── app.json                # Cấu hình crawler, DB, sites, Lark
│   └── signals.json            # Danh sách tín hiệu / metric name
│
├── deployments/
│   ├── docker-compose.prod.yml # Production (2 containers: backend + frontend)
│   └── docker-compose.yml      # Dev/test compose
│
├── logs/                       # Log xoay vòng (Lumberjack)
├── output/                     # Output từ crawler
│
├── .env.example                # Template biến môi trường
├── build_export_docker.ps1     # Script đóng gói Docker (Windows)
└── run_all.ps1                 # Script chạy local dev (Windows)
```

---

## 4. Cấu hình môi trường

Sao chép file mẫu và điền thông tin thực tế:

```powershell
cp .env.example .env
```

Nội dung file `.env`:

| Biến | Ví dụ | Mô tả |
|---|---|---|
| `DB_DSN` | `host=192.168.31.254 user=xxx password=xxx dbname=SHUNDAO_DASHBOARD port=2602 sslmode=disable TimeZone=Asia/Ho_Chi_Minh` | Connection string PostgreSQL |
| `VICTORIA_URL` | `http://100.81.226.125:8428` | Địa chỉ VictoriaMetrics |
| `JWT_SECRET` | `shundao-solar-secret-2026` | Khóa ký JWT (đổi trước khi deploy thật) |
| `GEMINI_API_KEY` | `AIzaSy...` | Google Gemini Flash API key |

> **⚠ Lưu ý bảo mật:** Không commit file `.env` vào Git. File `.gitignore` đã loại trừ file này.

### Cấu hình trong `configs/app.json`

File `configs/app.json` chứa các thông số quan trọng của Crawler và hệ thống:

| Trường | Giá trị mặc định | Mô tả |
|---|---|---|
| `system.fetch_interval_minutes` | `5` | Chu kỳ crawl dữ liệu (phút) |
| `system.browser_ttl_hours` | `24` | Thời gian sống phiên Headless Browser |
| `system.max_login_retries` | `3` | Số lần thử đăng nhập lại FusionSolar |
| `system.batch_size_inverter` | `15` | Số inverter xử lý mỗi batch |
| `credentials.username` | `om@raitek.vn` | Tài khoản FusionSolar |
| `sites` | SHUNDAO 1, SHUNDAO 2 | Danh sách trạm được monitor |

---

## 5. Quy trình khởi chạy (Local Dev)

> **Yêu cầu:** Windows, Golang ≥ 1.21, Node.js ≥ 18, npm

### Bước 1 — Chuẩn bị môi trường

```powershell
# Kiểm tra Go đã cài
go version

# Kiểm tra Node.js đã cài
node --version
npm --version
```

### Bước 2 — Cài phụ thuộc Frontend

```powershell
cd frontend
npm install
cd ..
```

### Bước 3 — Tạo file .env

```powershell
cp .env.example .env
# Mở .env và điền DB_DSN, VICTORIA_URL, JWT_SECRET thực tế
```

### Bước 4 — Khởi chạy toàn bộ hệ thống

```powershell
.\run_all.ps1
```

Script `run_all.ps1` thực hiện tuần tự:

```
[STEP 1] Dừng process "server.exe" cũ nếu đang chạy
[STEP 2] Build Backend Go → tạo ra server.exe
[STEP 3] Mở cửa sổ PowerShell mới → chạy ./server.exe
[STEP 4] Delay 2 giây để Backend khởi động xong
[STEP 5] Mở cửa sổ PowerShell thứ hai → chạy npm run dev trong /frontend
[STEP 6] In thông báo địa chỉ truy cập
```

**Kết quả — Địa chỉ truy cập:**

| Dịch vụ | URL |
|---|---|
| Giao diện Web | `http://localhost:9000` |
| Backend API (Legacy SSE) | `http://localhost:5039` |
| Backend API (Clean Arch) | `http://localhost:5040` |
| Health Check | `http://localhost:5040/healthz` |
| pprof Profiler | `http://localhost:6060/debug/pprof/` |

---

## 6. Quy trình Build & Triển khai Production (Docker)

### 6.1. Đóng gói Docker tại máy Windows (Dev Machine)

> **Yêu cầu:** Docker Desktop đang chạy trên Windows

```powershell
.\build_export_docker.ps1
```

Script thực hiện **4 bước tuần tự**:

```
[1/4] Build Backend Image  → fusion-backend:latest   (Alpine OS, ~30MB)
[2/4] Build Frontend Image → fusion-frontend:latest  (NGINX static, ~25MB)
[3/4] docker save → xuất ra file shundao_production.tar
[4/4] Tạo thư mục shundao_deploy_package/ gồm:
       ├── shundao_production.tar   ← Docker images
       ├── docker-compose.prod.yml  ← Định nghĩa container
       ├── .env.example             ← Template cấu hình
       └── configs/                 ← app.json, signals.json
```

### 6.2. Copy gói lên Server Linux

```bash
# Tại Windows — dùng SCP hoặc WinSCP
scp -r .\shundao_deploy_package\ user@server-ip:/home/user/shundao/
```

### 6.3. Khởi động trên Server Linux

SSH vào server và chạy theo thứ tự:

```bash
# ─── BƯỚC 1: Di chuyển vào thư mục deploy ────────────────────────
cd /home/user/shundao/shundao_deploy_package

# ─── BƯỚC 2: Nạp Docker images từ file tar ───────────────────────
docker load -i shundao_production.tar
# Kết quả mong đợi:
# Loaded image: fusion-backend:latest
# Loaded image: fusion-frontend:latest

# ─── BƯỚC 3: Tạo file cấu hình môi trường ────────────────────────
cp .env.example .env
nano .env   # Điền DB_DSN, VICTORIA_URL, JWT_SECRET thực tế

# ─── BƯỚC 4: Khởi động hệ thống ──────────────────────────────────
docker compose -f docker-compose.prod.yml --env-file .env up -d

# ─── BƯỚC 5: Kiểm tra trạng thái container ───────────────────────
docker compose -f docker-compose.prod.yml ps
```

**Kết quả mong đợi sau khi khởi động:**

| Container | Port | Trạng thái |
|---|---|---|
| `shundao-backend` | `5039`, `5040` | `healthy` (sau ~15 giây) |
| `shundao-frontend` | `2608` | `healthy` (sau backend healthy) |

**Truy cập web:** `http://<IP_SERVER>:2608`

### 6.4. Quản lý Container

```bash
# Xem log realtime
docker compose -f docker-compose.prod.yml logs -f backend
docker compose -f docker-compose.prod.yml logs -f frontend

# Dừng hệ thống
docker compose -f docker-compose.prod.yml down

# Khởi động lại backend
docker compose -f docker-compose.prod.yml restart backend

# Xem chi tiết sức khỏe
docker inspect shundao-backend | grep -A 10 "Health"
```

---

## 7. Hướng dẫn sử dụng giao diện

### 7.1. Đăng nhập

1. Mở trình duyệt, truy cập `http://localhost:9000` (dev) hoặc `http://<IP_SERVER>:2608` (prod).
2. Nhập **Tên đăng nhập** và **Mật khẩu** do quản trị viên cấp.
3. Nhấn **Đăng nhập**.

> **Giới hạn an toàn:** Sau **5 lần nhập sai liên tiếp**, địa chỉ IP sẽ bị **khóa tạm thời** theo cơ chế Exponential Backoff (5 phút → 10 phút → 30 phút → 90 phút...). Màn hình sẽ hiện countdown đếm ngược thời gian mở khóa.

### 7.2. Dashboard chính

Sau khi đăng nhập thành công, giao diện hiển thị:

| Vùng | Thông tin |
|---|---|
| **Header** | Tên trạm, thời gian cập nhật cuối, trạng thái kết nối SSE |
| **Metric Cards** | Active Power (kW), Daily Yield (kWh), Monthly Yield, Total Yield |
| **Biểu đồ đường** | Sản lượng theo giờ trong ngày (Daily Line Chart) |
| **Danh sách Inverter** | Trạng thái Online/Offline/Fault từng inverter |
| **Bảng cảnh báo** | Danh sách inverter/string bị lỗi đang ở trạng thái Confirmed |

### 7.3. Theo dõi cảnh báo thông minh

- **Màu xanh lá**: Inverter hoạt động bình thường.
- **Màu vàng (Pending)**: Phát hiện bất thường, đang trong thời gian debounce **15 phút** — chưa xác nhận lỗi.
- **Màu đỏ (Confirmed)**: Lỗi đã kéo dài > 15 phút → kích hoạt âm thanh cảnh báo.

**Cấu hình String bị loại trừ** (ví dụ: string đang bảo trì):
- Vào phần cài đặt, nhập danh sách số thứ tự chuỗi cần bỏ qua, ví dụ: `1,2,7,8`.
- Các string trong danh sách sẽ không bao giờ kích hoạt cảnh báo.

### 7.4. Xem dữ liệu lịch sử

- Chọn **khoảng thời gian** từ picker (ngày / tuần / tháng).
- Biểu đồ Recharts hỗ trợ **Zoom**: giữ `Ctrl` + cuộn chuột để phóng to.
- Nhấn vào legend để **bật/tắt** từng layer dữ liệu trên biểu đồ.

### 7.5. Đổi mật khẩu

1. Nhấn vào avatar góc trên phải → **"Đổi mật khẩu"**.
2. Nhập mật khẩu cũ và mật khẩu mới (tối thiểu 8 ký tự).
3. Xác nhận. Hệ thống sẽ hash lại bằng `bcrypt` và lưu vào PostgreSQL.

---

## 8. Luồng dữ liệu & Data Pipeline

Hệ thống vận hành tự động 24/7 theo quy trình tuần tự sau:

```
┌─────────────────────────────────────────────────────────────┐
│  VÒNG LẶP CRAWL (Lặp lại mỗi 5 phút — cấu hình app.json)  │
└─────────────────────────────────────────────────────────────┘
        │
        ▼
[1] Chromedp mở phiên Headless Browser (chạy ngầm, không hiện UI)
        │  → Truy cập FusionSolar intl.fusionsolar.huawei.com
        │  → Điền tài khoản / mật khẩu tự động (configs/credentials)
        │  → Xác thực SSO thành công → lấy Cookie phiên
        │
        ▼
[2] API Fetcher gọi nội bộ FusionSolar (dùng Cookie vừa lấy)
        │  → /rest/pvms/web/station/v1/overview/station-kpi-data  (KPI tổng)
        │  → /rest/pvms/web/device/v1/device-real-kpi-history      (Inverter)
        │  → /rest/pvms/web/device/v1/inverter-string-data         (String PV)
        │  → Xử lý theo batch: inverter (15/batch), meter (20/batch)
        │
        ▼
[3] Chuẩn hóa & Lưu vào VictoriaMetrics
        │  → Normalize metric name (dựa theo configs/signals.json)
        │  → Ghi hàng loạt (bulk write) theo định dạng Prometheus
        │  → Label: site="SHUNDAO 1", device="Inverter-03", ...
        │
        ▼
[4] Backend đóng gói payload JSON và PUSH về Frontend qua SSE
        │  → Endpoint: GET /api/stream/dashboard  (persistent connection)
        │  → Độ trễ từ lấy dữ liệu → Frontend nhảy số: < 1 giây
        │
        ▼
[5] Smart Alert Engine phân tích (chạy song song)
        │  → So sánh giá trị kW hiện tại với ngưỡng (80% công suất danh định)
        │  → Bỏ qua String trong danh sách "Loại trừ"
        │  → Debounce 15 phút: chỉ Confirm sau khi lỗi kéo dài liên tục
        │  → Nếu Confirmed → ghi log vào Lark Bitable + kích hoạt âm báo UI
        │
        ▼
[6] Lumberjack xoay log định kỳ
        │  → Ghi toàn bộ log vận hành vào backend/logs/
        │  → Tự xoay file khi đạt dung lượng giới hạn → không đầy disk
        └──────────────────────────────(quay lại Bước 1)
```

### Session tự phục hồi (Self-Healing)

Nếu phiên FusionSolar bị hết hạn (Cookie expire sau ~24h), hệ thống **tự động**:
1. Phát hiện lỗi xác thực (401 từ FusionSolar).
2. Kích hoạt luồng đăng nhập lại (`max_login_retries: 3`).
3. Làm mới Cookie và tiếp tục crawl — **không cần can thiệp thủ công**.

---

## 9. Mã lỗi & Phương pháp xử lý

### 9.1. Mã lỗi HTTP từ Backend API

Dưới đây là toàn bộ mã lỗi HTTP mà hệ thống có thể trả về, kèm nguyên nhân và hướng xử lý cụ thể:

---

#### 400 Bad Request — `BAD_REQUEST`

| Trường | Thông tin |
|---|---|
| **Mã nội bộ** | `BAD_REQUEST` |
| **Khi nào xảy ra** | Body request không đúng JSON / thiếu trường bắt buộc (`username`, `password`, `old_password`, `new_password`) |
| **Response mẫu** | `{"error": "Tên đăng nhập và mật khẩu không được để trống"}` |

**Cách xử lý:**
1. Kiểm tra lại request body gửi lên có đúng định dạng JSON không.
2. Đảm bảo `Content-Type: application/json` được set trong header.
3. Không để trống các trường bắt buộc.

---

#### 401 Unauthorized — `UNAUTHORIZED` / `SESSION_EXPIRED`

| Trường | Thông tin |
|---|---|
| **Mã nội bộ** | `UNAUTHORIZED`, `SESSION_EXPIRED` |
| **Khi nào xảy ra** | — Nhập sai tên đăng nhập hoặc mật khẩu<br>— JWT Token hết hạn (mặc định: 15 phút)<br>— Không có header `Authorization: Bearer <token>`<br>— Token bị giả mạo / sai chữ ký |
| **Response mẫu** | `{"error": "Tên đăng nhập hoặc mật khẩu không đúng", "remaining": 3}` |

**Cách xử lý — người dùng cuối:**
1. Kiểm tra lại tên đăng nhập và mật khẩu (phân biệt chữ hoa/thường).
2. Nếu Token hết hạn → Frontend tự động gọi `/api/auth/refresh` để làm mới.
3. Nếu không thể refresh → tự động chuyển hướng về trang Login.

**Cách xử lý — kỹ thuật viên:**
- Trường `remaining` trong response cho biết còn bao nhiêu lần thử trước khi bị khóa IP.
- Kiểm tra log: `backend/logs/` → tìm dòng `[AUTH] Login failed for user=...`.

---

#### 403 Forbidden

| Trường | Thông tin |
|---|---|
| **Mã nội bộ** | *(HTTP standard)* |
| **Khi nào xảy ra** | — Cố tình đăng nhập ngoài giờ hoạt động (nếu Time-Lock được bật)<br>— Gọi `/api/auth/refresh` mà thiếu header `X-Shundao-Bot-Check` (chống bot) |
| **Response mẫu** | `{"error": "Hệ thống chỉ hoạt động từ 04:00 đến 18:30. Vui lòng quay lại trong giờ làm việc."}` |

**Cách xử lý:**
1. **Time-Lock:** Đợi đến giờ hoạt động (04:00 – 18:30) hoặc liên hệ quản trị viên để được cấp quyền khẩn cấp.
2. **Bot-Check:** Đảm bảo request từ Frontend hợp lệ luôn đính kèm header `X-Shundao-Bot-Check`.

---

#### 418 I'm a Teapot — Lockdown (Chống Brute-Force)

| Trường | Thông tin |
|---|---|
| **Mã nội bộ** | `hacked: true` |
| **Khi nào xảy ra** | IP nhập sai mật khẩu **≥ 5 lần liên tiếp** |
| **Response mẫu** | `{"error": "Truy cập tạm thời bị khóa", "hacked": true, "unlock_at": 1744123456}` |

**Lịch trình khóa (Exponential Backoff):**

| Lần khóa | Thời gian khóa |
|---|---|
| Lần 1 | 5 phút |
| Lần 2 | 10 phút |
| Lần 3 | 30 phút |
| Lần 4 | 90 phút |
| Lần 5+ | Tăng x3 mỗi lần (270 phút → 810 phút → ...) |

**Cách xử lý — người dùng cuối:**
1. Giao diện sẽ hiện màn hình **Lockdown** với đồng hồ đếm ngược (`unlock_at`).
2. Đợi hết thời gian khóa, sau đó thử lại.
3. Liên hệ quản trị viên nếu không nhớ mật khẩu.

**Cách xử lý — kỹ thuật viên (mở khóa khẩn cấp):**
```bash
# SSH vào server, restart backend để xóa in-memory lockout table
docker compose -f docker-compose.prod.yml restart backend
# ⚠ Lưu ý: Reset toàn bộ IP records. Chỉ dùng khi khẩn cấp.
```

---

#### 404 Not Found — `NOT_FOUND`

| Trường | Thông tin |
|---|---|
| **Mã nội bộ** | `NOT_FOUND` |
| **Khi nào xảy ra** | Tài khoản không tồn tại trong database khi đổi mật khẩu |
| **Response mẫu** | `{"error": "Tài khoản không tồn tại"}` |

**Cách xử lý:**
1. Kiểm tra lại tên đăng nhập.
2. Liên hệ quản trị viên để tạo lại tài khoản trong PostgreSQL.

---

#### 429 Too Many Requests — `RATE_LIMITED`

| Trường | Thông tin |
|---|---|
| **Mã nội bộ** | `RATE_LIMITED` |
| **Khi nào xảy ra** | Vượt ngưỡng **30 request/giây** hoặc **burst 50 request** tức thời (áp dụng toàn hệ thống — Token Bucket Algorithm) |
| **Response mẫu** | `{"error": {"code": "RATE_LIMITED", "message": "Too many requests. Please try again later."}}` |
| **Header bổ sung** | `Retry-After: 1` |

**Cách xử lý:**
1. Frontend nên implement **retry với delay** khi nhận 429 (đọc header `Retry-After`).
2. **Không bao giờ** gửi request vòng lặp tight (polling liên tục) — hãy dùng SSE đã có sẵn.
3. Nếu hệ thống monitoring nội bộ bị ảnh hưởng, kiểm tra có agent/bot nào đang quét không.

---

#### 500 Internal Server Error — `INTERNAL_ERROR`

| Trường | Thông tin |
|---|---|
| **Mã nội bộ** | `INTERNAL_ERROR` |
| **Khi nào xảy ra** | — Lỗi kết nối PostgreSQL (DB down, sai DSN)<br>— Lỗi kết nối VictoriaMetrics<br>— Lỗi trong quá trình hash bcrypt<br>— Lỗi generate JWT token |

**Cách xử lý — kỹ thuật viên:**

```bash
# 1. Xem log backend ngay lập tức
docker compose -f docker-compose.prod.yml logs --tail=100 backend

# 2. Kiểm tra kết nối PostgreSQL
docker exec -it shundao-backend wget -qO- http://localhost:5040/healthz

# 3. Kiểm tra biến môi trường đã đúng chưa
docker exec shundao-backend env | grep DB_DSN
docker exec shundao-backend env | grep VICTORIA_URL

# 4. Restart nếu cần
docker compose -f docker-compose.prod.yml restart backend
```

---

### 9.2. Mã lỗi nội bộ Backend (AppError codes)

Các mã này xuất hiện trong JSON response dưới trường `error.code`:

| Code | HTTP Status | Mô tả |
|---|---|---|
| `UNAUTHORIZED` | 401 | Chưa xác thực / Token không hợp lệ |
| `SESSION_EXPIRED` | 401 | Phiên đã hết hạn, cần đăng nhập lại |
| `NOT_FOUND` | 404 | Tài nguyên không tìm thấy |
| `BAD_REQUEST` | 400 | Dữ liệu đầu vào không hợp lệ |
| `INTERNAL_ERROR` | 500 | Lỗi nội bộ server |
| `RATE_LIMITED` | 429 | Vượt giới hạn tốc độ request |
| `FETCH_FAILED` | 500 | Không lấy được dữ liệu từ FusionSolar |

---

### 9.3. Quy trình xử lý sự cố tuần tự (Troubleshooting Flowchart)

```
Phát hiện sự cố?
      │
      ▼
[1] Kiểm tra Health Endpoint
    GET http://localhost:5040/healthz
      │
      ├─ 200 OK ──────────────────────────────────────────────────────┐
      │                                                               │
      └─ Không phản hồi / lỗi                                        │
            │                                                         │
            ▼                                                         │
      [2] Xem log Docker                                              │
          docker logs shundao-backend --tail=50                       │
            │                                                         │
            ├─ "DB connection failed" ─→ Kiểm tra DB_DSN trong .env  │
            ├─ "Port already in use"  ─→ Tắt process đang chiếm port │
            └─ Crash / panic          ─→ Báo cáo kèm stack trace     │
                                                                      │
                                            Backend OK ───────────────┘
                                                │
                                                ▼
                                      [3] Kiểm tra Frontend
                                          curl http://localhost:2608
                                            │
                                            ├─ 200 OK ─→ Thử xóa cache trình duyệt
                                            └─ Lỗi    ─→ docker logs shundao-frontend
                                                │
                                                ▼
                                          [4] Kiểm tra SSE stream
                                              curl -N http://localhost:5039/api/stream/dashboard
                                            │
                                            ├─ Có data ─→ Lỗi ở UI, kiểm tra console trình duyệt
                                            └─ Không có ─→ Crawler bị tắc, xem log chi tiết
                                                │
                                                ▼
                                          [5] Kiểm tra pprof (nếu nghi ngờ memory leak)
                                              go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## 10. Giám sát sức khỏe hệ thống

### Health Check tự động

Docker tự động kiểm tra sức khỏe container mỗi 30 giây:

```yaml
healthcheck:
  test: wget --spider http://localhost:5040/healthz
  interval: 30s
  timeout: 10s
  retries: 3
```

Nếu backend không phản hồi sau 3 lần thử → Docker đánh dấu `unhealthy` → Frontend container sẽ không khởi động.

### Profiling bộ nhớ (Go pprof)

```bash
# Xem heap allocation (ai đang chiếm RAM nhiều nhất)
go tool pprof http://localhost:6060/debug/pprof/heap

# Xem goroutine đang chạy (phát hiện goroutine leak)
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Xem CPU profile trong 30 giây
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

> Port `6060` chỉ expose nội bộ (không mở ra ngoài internet). SSH tunnel nếu cần truy cập từ xa:
> ```bash
> ssh -L 6060:localhost:6060 user@server-ip
> ```

### Log hệ thống

```bash
# Xem log realtime
tail -f backend/logs/app.log

# Tìm lỗi trong log
grep "ERROR" backend/logs/app.log | tail -20

# Log Docker (production)
docker compose -f docker-compose.prod.yml logs -f --tail=100
```

---

## 11. Câu hỏi thường gặp (FAQ)

**Q: Tại sao giao diện không cập nhật số liệu?**
> A: Kiểm tra kết nối SSE — nhìn vào header status bar của Dashboard. Nếu hiện "Disconnected", thử reload trang. SSE sẽ tự động reconnect sau 3-5 giây.

**Q: Crawler không lấy được dữ liệu mới?**
> A: Kiểm tra log backend. Thường do Cookie FusionSolar hết hạn → hệ thống sẽ tự đăng nhập lại (Self-Healing) trong vòng 1 chu kỳ (5 phút). Nếu vẫn lỗi sau 15 phút, kiểm tra credentials trong `configs/app.json`.

**Q: Tôi bị khóa IP 418 do nhập sai mật khẩu. Phải làm gì?**
> A: Đợi thời gian lockout hết (giao diện có đồng hồ đếm ngược). Nếu khẩn cấp, yêu cầu kỹ thuật viên restart container backend.

**Q: Làm sao thêm trạm mới (site mới)?**
> A: Thêm entry vào mảng `sites` trong `configs/app.json` với `id` (NE code từ FusionSolar) và `name` tương ứng, sau đó restart backend.

**Q: Dữ liệu VictoriaMetrics lưu bao lâu?**
> A: Mặc định VictoriaMetrics giữ data 1 năm (`-retentionPeriod=12`). Có thể điều chỉnh khi khởi động VictoriaMetrics.

---

> **Phát triển bởi:** Phạm Hoàng Phúc
> **Phiên bản kiến trúc:** 2.5 (Năm 2026)
