// ─── Refresh Intervals ──────────────────────────────────────────────
/** Khoảng thời gian auto-refresh cho biểu đồ công suất (ms) */
export const CHART_REFRESH_INTERVAL = 5 * 60 * 1000; // 5 phút

/** Khoảng thời gian auto-refresh cho dashboard tổng quan (ms) */
export const DASHBOARD_REFRESH_INTERVAL = 60 * 1000; // 1 phút

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

// ─── Solar Hours ────────────────────────────────────────────────────
/** Giờ bắt đầu hiển thị dữ liệu (Backend) */
export const SOLAR_START_HOUR = 6;

/** Giờ kết thúc hiển thị dữ liệu (Backend) */
export const SOLAR_END_HOUR = 18;
