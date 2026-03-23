// ─────────────────────────────────────────────────────────────────────────────
// FRONTEND CONFIGURATION (Single Source of Truth)
// Chỉnh sửa tại đây để điều chỉnh toàn bộ hành vi hệ thống giao diện.
// ─────────────────────────────────────────────────────────────────────────────

// ─── Solar / Working Hours ───────────────────────────────────────────
/** Giờ bắt đầu thu/hiển thị dữ liệu năng lượng mặt trời */
export const SOLAR_START_HOUR = 6;

/** Giờ kết thúc thu/hiển thị dữ liệu năng lượng mặt trời */
export const SOLAR_END_HOUR = 18;

// ─── Refresh Intervals ──────────────────────────────────────────────
/** Khoảng thời gian auto-refresh cho biểu đồ công suất (ms) */
export const CHART_REFRESH_INTERVAL = 5 * 60 * 1000; // 5 phút

// (Removed DASHBOARD_REFRESH_INTERVAL because SSE is used instead)

// ─── API ────────────────────────────────────────────────────────────
/** API request timeout (ms) */
export const API_TIMEOUT = 10_000; // 10 giây

// ─── Chart Config ───────────────────────────────────────────────────
/** Số điểm dữ liệu tối thiểu khi zoom */
export const CHART_MIN_VISIBLE_POINTS = 10;

/** Số mốc thời gian hiển thị trên trục X */
export const CHART_X_AXIS_TICK_COUNT = 12;

/** Thời lượng animation khi dữ liệu thay đổi (ms) */
export const CHART_ANIMATION_DURATION = 800;

// ─── Smart Alert Algorithm ──────────────────────────────────────────
/**
 * Cấu hình thuật toán cảnh báo thông minh (Stateful Debounce).
 * Thay đổi các giá trị này nếu muốn điều chỉnh độ nhạy hoặc độ trễ cảnh báo.
 */
export const ALERT_CONFIG = {
    /** Tỷ lệ phần trăm ngưỡng: String < X% so với average → Cảnh báo */
    thresholdPercent: 0.8,

    /** Dòng điện tối thiểu tuyệt đối (A) - Dưới mức này coi là nhiễu rác */
    minCurrentThresholdA: 0.5,

    /**
     * Thời gian giữ nguyên lỗi liên tục trước khi phát Alarm (ms).
     * Faults that self-heal within this window are silently discarded.
     */
    debounceMs: 15 * 60 * 1000, // 15 phút
} as const;

// ─── Audio Assets ───────────────────────────────────────────────────
/**
 * Đường dẫn tới các file âm thanh (đặt trong /public/assets/audio/).
 * Thay đổi tên file hoặc thêm file nhạc mới vào thư mục đó mà không cần sửa code.
 */
export const AUDIO_ASSETS = {
    alarm: '/assets/audio/alarm.mp3',
    gentleChime: '/assets/audio/gentle_chime.mp3',
    happyChime: '/assets/audio/happy_chime.mp3',
} as const;

// ─── Query Client Config ────────────────────────────────────────────
/** Cấu hình React Query Client mặc định */
export const QUERY_CONFIG = {
    /** Thời gian cache data trước khi coi là "stale" (ms) */
    staleTime: CHART_REFRESH_INTERVAL, // 5 phút
    /** Số lần retry khi API thất bại */
    retryCount: 1,
} as const;

// ─── Colors ─────────────────────────────────────────────────────────
export const COLORS = {
    /** Công suất thuần (DC Power) */
    dcPower: {
        stroke: '#f59e0b',
        gradientStart: 'rgba(245, 158, 11, 0.25)',
        gradientEnd: 'rgba(245, 158, 11, 0.02)',
        text: 'text-amber-600',
        bg: 'bg-amber-50',
        border: 'border-amber-200',
        dot: 'bg-amber-500',
    },
    /** Tổng công suất đầu vào (AC Power / p_out) */
    acPower: {
        stroke: '#3b82f6',
        gradientStart: 'rgba(59, 130, 246, 0.25)',
        gradientEnd: 'rgba(59, 130, 246, 0.02)',
        text: 'text-blue-600',
        bg: 'bg-blue-50',
        border: 'border-blue-200',
        dot: 'bg-blue-500',
    },
    /** Status indicators */
    status: {
        online: 'bg-green-500',
        offline: 'bg-red-500',
        warning: 'bg-yellow-500',
    },
} as const;
