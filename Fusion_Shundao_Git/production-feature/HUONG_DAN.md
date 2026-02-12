# Hướng Dẫn Tích Hợp — Production Feature (Solar Dashboard)

## 📋 Tổng Quan

Feature này bao gồm **2 phần giao diện chính** của Dashboard giám sát năng lượng mặt trời:

1. **Production Section** — 3 thẻ metric tổng quan (Sản lượng hôm nay, Tổng sản lượng, Công suất định mức)
2. **Production Chart** — Biểu đồ tổng hợp Công suất + Bức xạ (Line chart hàng ngày / Bar chart hàng tháng)

---

## 📁 Cấu Trúc Thư Mục

```
production-feature/
├── HUONG_DAN.md              ← File này
├── components/
│   ├── charts/
│   │   └── ProductionChart.tsx    ← Biểu đồ chính (Line + Bar, zoom, dual Y-axis)
│   ├── dashboard/
│   │   └── ProductionSection.tsx  ← 3 thẻ metric tổng quan
│   ├── widgets/
│   │   └── DetailedMetricCard.tsx ← Card hiển thị metric + breakdown theo site
│   └── ui/
│       ├── Card.tsx               ← UI primitive: Card container
│       └── Skeleton.tsx           ← UI primitive: Loading skeleton
├── services/
│   └── api.ts                     ← Axios instance (gọi API backend)
├── types/
│   └── index.ts                   ← TypeScript types (KPI, Site, Inverter, ...)
└── utils/
    └── cn.ts                      ← Utility gộp class CSS (clsx + tailwind-merge)
```

---

## 🧩 Mô Tả Từng File

### `ProductionSection.tsx`
- Nhận props: `kpi` (tổng KPI), `sites` (danh sách site), `isLoading`
- Render 3 thẻ `DetailedMetricCard` cho: **Sản lượng hôm nay** (MWh), **Tổng sản lượng** (GWh), **Công suất định mức** (MW)
- Mỗi thẻ hiển thị tổng + breakdown theo từng site

### `ProductionChart.tsx`
- **2 chế độ xem**: "Hôm nay" (LineChart) và "Theo tháng" (BarChart)
- **Dual Y-axis**: trục trái = Công suất (kW), trục phải = Bức xạ (W/m²)
- **Zoom**: Ctrl + Scroll để zoom vào khoảng thời gian cụ thể
- **Tooltip tùy chỉnh**: hiển thị dữ liệu chi tiết cả 2 site khi hover
- **Lọc giờ**: chỉ hiển thị dữ liệu 06:00–18:00 (giờ có nắng)
- Tổng MWh mỗi site hiển thị ở header biểu đồ

### `DetailedMetricCard.tsx`
- Card có animation (framer-motion) fade-in khi load
- Hiển thị: icon + title → tổng giá trị → breakdown từng site
- Hỗ trợ 6 theme màu: `solar`, `blue`, `green`, `slate`, `orange`, `emerald`
- Có skeleton loading state

### `Card.tsx` / `Skeleton.tsx`
- UI primitives dùng chung, hỗ trợ 3 variant (`default`, `glass`, `gradient`)

### `api.ts`
- Axios instance với baseURL từ `VITE_API_URL`
- Response interceptor trả `response.data` trực tiếp
- Timeout 10s

### `types/index.ts`
- Định nghĩa các interface: `KPI`, `Site`, `SmartLogger`, `Inverter`, `Sensor`, `Meter`, ...

### `cn.ts`
- Utility kết hợp `clsx` + `tailwind-merge` để gộp class CSS

---

## 📦 Dependencies Cần Cài

```bash
npm install recharts @tanstack/react-query framer-motion lucide-react axios clsx tailwind-merge
```

| Package | Mục đích |
|---|---|
| `recharts` | LineChart, BarChart, XAxis, YAxis, Tooltip... |
| `@tanstack/react-query` | Fetch monthly data với cache |
| `framer-motion` | Animation cho MetricCard |
| `lucide-react` | Icon (Zap, Activity, LineChart, BarChart2...) |
| `axios` | HTTP client |
| `clsx` + `tailwind-merge` | Utility gộp CSS class |

---

## 🔌 API Endpoints Cần Có

| Endpoint | Method | Mô tả | Response |
|---|---|---|---|
| `/api/production-monthly` | GET | Lấy dữ liệu sản lượng theo tháng | `MonthlyDataPoint[]` |
| (Daily data truyền qua props) | — | Dữ liệu hàng ngày được truyền từ component cha | `ProductionDataPoint[]` |

### Cấu trúc dữ liệu Daily (truyền qua props):
```ts
interface ProductionDataPoint {
    date: string;           // "06:00", "06:05", ...
    site1Power: number;     // kW
    site1Irradiance: number; // W/m²
    site2Power: number;
    site2Irradiance: number;
}
```

### Cấu trúc dữ liệu Monthly (từ API):
```ts
interface MonthlyDataPoint {
    date: string;              // "01", "02", ... (ngày trong tháng)
    site1MaxPower: number | null;
    site1MaxIrrad: number | null;
    site2MaxPower: number | null;
    site2MaxIrrad: number | null;
}
```

---

## 🚀 Cách Sử Dụng

### 1. ProductionSection
```tsx
import { ProductionSection } from './components/dashboard/ProductionSection';

<div className="grid grid-cols-3 gap-4">
    <ProductionSection
        kpi={kpiData}        // KPI tổng hệ thống
        sites={sitesData}    // Mảng Site[]
        isLoading={loading}
    />
</div>
```

### 2. ProductionChart
```tsx
import { ProductionChart } from './components/charts/ProductionChart';

<ProductionChart
    data={dailyProductionData}  // ProductionDataPoint[]
    loading={isLoading}
/>
```

---

## ⚠️ Lưu Ý Khi Bứng Sang Dự Án Khác

1. **TailwindCSS** — Toàn bộ styling dùng Tailwind. Dự án đích phải có Tailwind đã cấu hình
2. **Import paths** — Cần chỉnh lại đường dẫn import (`../../utils/cn`, `../../services/api`...) cho phù hợp project mới
3. **Hardcode tên site** — `"Shundao 1"`, `"Shundao 2"` đang hardcode trong `ProductionChart.tsx`. Cần thay đổi nếu dùng cho site khác
4. **Biến môi trường** — Cần set `VITE_API_URL` trong `.env` của project mới
5. **React Query Provider** — Project mới cần wrap `<QueryClientProvider>` ở root component
6. **Dữ liệu daily** — interval mặc định là 5 phút (`5/60` trong tính toán MWh). Chỉnh nếu interval khác
