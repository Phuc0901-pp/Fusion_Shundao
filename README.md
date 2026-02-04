# FusionSolar Data Formatter (Fusion-Format)

> **Note**: Tài liệu này mô tả chi tiết kiến trúc và quy trình hoạt động của hệ thống thu thập dữ liệu FusionSolar phiên bản mới nhất.

---

## 1. Tổng quan hệ thống
**Fusion-Format** là một giải pháp Middleware tự động hóa, được xây dựng để kết nối giữa hệ thống đóng (closed system) Huawei FusionSolar và các hệ thống giám sát bên thứ ba. Công cụ này giải quyết bài toán:
- Tự động đăng nhập và duy trì phiên làm việc (Session).
- Vượt qua sự phức tạp của API nội bộ (Internal API) của Huawei.
- Chuẩn hóa dữ liệu từ hàng ngàn mã tín hiệu (ID) khó hiểu sang dạng JSON tường minh.

---

## 2. Quy trình Xử lý Dữ liệu (Data Pipeline)

Hệ thống hoạt động theo quy trình khép kín 3 bước: **Fetch (Lấy) -> Format (Chuẩn hóa) -> Save (Lưu trữ)**.

### Bước 1: Thu thập Dữ liệu (Data Fetching)
Module `src/api` và `src/browser` thực hiện:

1.  **Khởi tạo & Đăng nhập**:
    -   Sử dụng **Headless Chrome** để giả lập người dùng thật truy cập trang login.
    -   Tự động nhập User/Pass (từ `app.json`).
    -   Bắt (Intercept) token xác thực **`Roarand`** từ network request.
    -   **Cơ chế Self-Healing**: Nếu token hết hạn (API trả về 401), hệ thống tự động kích hoạt luồng đăng nhập lại.

2.  **Khám phá Trạm & Thiết bị (Discovery)**:
    -   **Site Tree**: Từ cấu hình `app.json`, hệ thống quét cây tổ chức để tìm tất cả các **SmartLogger**.
    -   **Device Scan**: Với mỗi SmartLogger, hệ thống quét tiếp để lấy danh sách thiết bị con (Inverter, Meter, Sensor).

3.  **Lấy dữ liệu vận hành (Operational Data)**:
    -   **Batch Processing**: Để tăng tốc độ, hệ thống không gọi API lẻ tẻ mà gom thiết bị thành các lô (Batch) - mặc định 15 Inverter/lô.
    -   **Multichannel Fetching**:
        -   *Realtime KPI*: Lấy P, Q, U, I, Nhiệt độ, Hiệu suất.
        -   *String Data*: Lấy điện áp/dòng điện của từng chuỗi pin (PV String).

### Bước 2: Chuẩn hóa Dữ liệu (Formatting)
Module `src/formatter` thực hiện:

1.  **Phát hiện cấu trúc (Structure Detection)**:
    -   API của Huawei trả về dữ liệu lúc thì dạng **Lồng nhau (Nested)** (cho Inverter), lúc thì dạng **Phẳng (Flat)** (cho SmartLogger).
    -   Hàm `extractSignals` tự động phân tích và đưa về dạng Map phẳng nhất quán.

2.  **Ánh xạ Tín hiệu (Signal Mapping)**:
    -   Dữ liệu thô chỉ chứa các ID vô nghĩa (Ví dụ: `100` là công suất, `101` là điện áp).
    -   Hệ thống đọc file `config/signals.json` để dịch các ID này sang tên tiếng Anh chuẩn (Ví dụ: `"active_power"`, `"output_current_a"`).

3.  **Làm sạch & Thêm Metadata**:
    -   Gắn thêm thông tin: Tên trạm, Tên thiết bị, Serial Number, Timestamp.
    -   Loại bỏ các trường dư thừa hoặc không xác định.

### Bước 3: Lưu trữ (Saving)
Module `src/main.go` điều phối việc lưu:
-   Tự động tạo cây thư mục theo tên trạm và tên thiết bị.
-   Ghi file JSON với format "Pretty Print" (dễ đọc).

---

## 3. Các tính năng cốt lõi

*    **High Performance**: Sử dụng kỹ thuật Batch Request giúp giảm thời gian lấy dữ liệu từ vài phút xuống còn vài chục giây cho hàng trăm thiết bị.
*    **Config-Driven Architecture**: Mọi thông số (Danh sách trạm, ID bản đồ, API Endpoint) đều nằm ở file JSON bên ngoài. Không cần sửa code Go khi thêm trạm mới.
*    **Auto-Recovery**: Tự động phát hiện lỗi mạng hoặc lỗi session để thử lại (Retry) hoặc đăng nhập lại.
*    **Structured Logging**: Ghi log chi tiết ra File và Console, hỗ trợ debug và trace lỗi hiệu quả.
*    **Unit Tested**: Các module logic cốt lõi (Config, Formatter) đã được kiểm thử để đảm bảo độ chính xác.

---

## 4. Cấu trúc thư mục dự án

```text
Fusion/
├── config/                 # CHỨA CẤU HÌNH (QUAN TRỌNG)
│   ├── app.json            # Cấu hình hệ thống, tài khoản, danh sách trạm
│   └── signals.json        # Bản đồ ánh xạ ID -> Tên trường (Inverter/Meter/Sensor)
├── logs/                   # CHỨA LOG FILE
│   └── app_YYYY-MM-DD.log  # File log sinh ra mỗi ngày
├── output/                 # CHỨA DỮ LIỆU ĐẦU RA
│   └── {Tên_Trạm}/...      # Xem chi tiết mục output data
├── src/                    # MÃ NGUỒN (SOURCE CODE)
│   ├── api/                # Giao tiếp với FusionSolar API
│   ├── browser/            # Điều khiển Chrome Driver
│   ├── config/             # Logic đọc file JSON config
│   ├── formatter/          # Logic chuẩn hóa dữ liệu
│   ├── login/              # Logic đăng nhập tự động
│   ├── utils/              # Các hàm tiện ích (Logger, UUID)
│   └── main.go             # Hàm Main
├── fusion.exe              # ⚙️ FILE CHẠY CHÍNH
└── go.mod                  # File quản lý thư viện Go
```

---

## 5. Định dạng dữ liệu đầu ra (Output Data)

Dữ liệu được lưu tại thư mục `output/` với cấu trúc phân cấp:

```text
output/
└── {Site_Name}/                # Thư mục Trạm
    ├── Station/
    │   └── overview.json       # Dữ liệu tổng quan trạm (Sản lượng, Doanh thu, CO2)
    └── {SmartLogger_Name}/     # Thư mục SmartLogger
        ├── smartLogger_data.json # Thông tin SmartLogger (IP, SN, Trạng thái)
        └── {Device_Name}/      # Thư mục Thiết bị con (Inverter, Meter...)
            ├── data.json       # Dữ liệu chi tiết thiết bị
```

### Mẫu file `data.json` (Inverter)
```json
{
  "timestamp": 1716182000,
  "device_name": "Inverter-1",
  "device_id": "NE=12345678",
  "data": {
    "active_power": 50.12,          // Công suất tác dụng (kW)
    "reactive_power": 0.5,          // Công suất phản kháng (kVar)
    "power_factor": 0.99,           // Hệ số công suất
    "efficiency": 98.5,             // Hiệu suất
    "internal_temperature": 45.2,   // Nhiệt độ bên trong
    "pv01_voltage": 650.5,          // Điện áp chuỗi 1
    "pv01_current": 10.2,           // Dòng điện chuỗi 1
    ...
  }
}
```

### Mẫu file `overview.json` (Station)
```json
{
  "station_name": "Solar Farm A",
  "kpi": {
    "daily_energy": 1250.5,         // Sản lượng ngày (kWh)
    "total_income": 5000000,        // Doanh thu
    "active_power": 450.2           // Công suất phát hiện tại (kW)
  },
  "env": {
    "co2_reduction": 800.5,         // Giảm thải CO2 (kg)
    "tree_planted": 45              // Số cây trồng tương đương
  }
}
```

---

## 6. Hướng dẫn Biên dịch & Cài đặt

### Yêu cầu môi trường
1.  **Go (Golang)**: Phiên bản 1.20 trở lên.
2.  **Google Chrome**: Cài đặt sẵn trên máy để chạy browser automation.

### Lệnh biên dịch
Mở terminal tại thư mục gốc dự án và chạy:
```powershell
go build -o fusion.exe ./src
```
Màn hình sẽ không báo lỗi và file `fusion.exe` sẽ xuất hiện (hoặc được cập nhật).

### Lệnh chạy test
Để kiểm tra tính đúng đắn của logic trước khi build:
```powershell
go test ./...
```

### Lệnh chạy ứng dụng
```powershell
./fusion.exe
```

---

## 7. Khắc phục sự cố (Troubleshooting)

| Vấn đề | Nguyên nhân | Cách khắc phục |
| :--- | :--- | :--- |
| **Login Failed** | Sai User/Pass hoặc web đổi cấu trúc | Kiểm tra `config/app.json` và log lỗi trong `logs/`. |
| **No Data** | Signal ID bị sai hoặc thiết bị offline | Kiểm tra `config/signals.json` và trạng thái thiết bị trên web. |
| **Chrome Error** | Phiên bản Chrome và Driver không khớp | Cập nhật Google Chrome lên bản mới nhất. |
| **Permission Denied** | Không thể ghi file vào `output/` | Chạy ứng dụng dưới quyền Administrator. |

--- 

> **Maintainer**: Phuc Pham
> **Last Update**: 04/02/2026
