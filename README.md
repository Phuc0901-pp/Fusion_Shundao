# FusionSolar Data Formatter (Fusion-Format)

Công cụ tự động hóa việc thu thập và chuẩn hóa dữ liệu từ hệ thống Huawei FusionSolar sang định dạng JSON đồng nhất, hỗ trợ quản lý tập trung và tích hợp hệ thống bên thứ ba.

##  Giới thiệu
**Fusion-Format** được thiết kế để giải quyết vấn đề cấu trúc dữ liệu phức tạp và không đồng nhất của FusionSolar. Công cụ sử dụng cơ chế Browser Automation để vượt qua rào cản đăng nhập và gọi trực tiếp các API nội bộ để lấy dữ liệu thời gian thực của Inverter, Power Meter, Weather Station và đặc biệt là thông tin chi tiết của SmartLogger.

##  Cấu trúc dự án
Dự án được viết bằng ngôn ngữ Go (Golang), chia thành các module logic:

```text
Fusion/
├── src/
│   ├── api/          # Xử lý các yêu cầu REST API tới FusionSolar
│   ├── browser/      # Điều khiển trình duyệt (chromedp) để login & lấy token
│   ├── formatter/    # Module cốt lõi: Chuyển đổi dữ liệu sang chuẩn Native
│   │   ├── mapper.go    # Định nghĩa bản đồ ánh xạ (Signal Map) các field
│   │   ├── converter.go # Logic xử lý cấu trúc Nested/Flat của API
│   │   ├── helper.go    # Các hàm hỗ trợ trích xuất Key-Value
│   │   └── types.go     # Định nghĩa cấu trúc dữ liệu chung
│   ├── login/        # Quản lý luồng đăng nhập
│   └── main.go       # Điểm khởi đầu, điều phối luồng chạy chính
├── config/           # Cấu hình hệ thống (Site, Output directory)
├── fusion.exe        # File thực thi đã build
└── go.mod            # Quản lý dependencies
```

##  Cơ chế hoạt động & Định dạng dữ liệu

### 1. Luồng xử lý logic
1. **Login & Scrape**: Khởi động trình duyệt ngầm (headless), lấy token `Roarand`.
2. **Device Discovery**: Quét toàn bộ cây tổ chức để tìm trạm (Site) và danh sách thiết bị dưới mỗi SmartLogger.
3. **SmartLogger Detail**: Sử dụng API của NetEco để lấy SN, IP, Model và Software Version.
4. **Data Normalization**: Ánh xạ các Signal ID khó hiểu sang tên tiếng Anh (snake_case) chuẩn.

### 2. Cấu trúc thư mục Output
Dữ liệu được lưu trữ phân cấp theo Site và SmartLogger:
```text
output/
└── {Site_Name}/
    ├── Station/
    │   └── overview.json       # KPI tổng quan của trạm
    └── {SmartLogger_Name}/
        ├── smartLogger_data.json # Thông tin chi tiết SmartLogger (SN, IP,...)
        └── {Device_Name}/
            ├── data.json         # Dữ liệu vận hành (P, Q, U, I,...)
            └── string_data.json  # Dữ liệu chi tiết từng chuỗi PV (nếu là Inverter)
```

### 3. Định dạng JSON mẫu
Tất cả dữ liệu được làm phẳng (flatten) để dễ dàng xử lý:
```json
{
  "timestamp": 1700000000,
  "device_name": "Inverter_01",
  "device_id": "NE=12345678",
  "data": {
    "active_power": 50.5,
    "input_voltage_pv1": 650.2,
    "internal_temperature": 45.0
  }
}
```

##  Hướng dẫn thực thi

### Yêu cầu hệ thống
- Hệ điều hành: Windows/Linux
- Trình duyệt: Google Chrome (để chạy automation)

### Lệnh thực thi
Để biên dịch lại dự án:
```powershell
go build -o fusion.exe ./src
```

Để chạy ứng dụng:
```powershell
./fusion.exe
```

##  Tính năng nổi bật
- **SmartLogger Grouping**: Tự động nhóm thiết bị theo SmartLogger quản lý.
- **Auto-Detection**: Tự động nhận diện Weather Station, Power Meter và Inverter để áp dụng bộ map dữ liệu riêng.
- **Robust Parsing**: Xử lý được cả cấu trúc dữ liệu lồng nhau (Nested) và phẳng (Flat) từ API Huawei.
- **Clean Output**: Chỉ giữ lại những dữ liệu quan trọng, loại bỏ các trường thừa từ API gốc.
