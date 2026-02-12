import axios from 'axios';
import { toast } from 'sonner';
import { API_TIMEOUT } from '../config/constants';

// Create Axios instance with default config
const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL || '/api',
    headers: {
        'Content-Type': 'application/json',
    },
    timeout: API_TIMEOUT,
});

// Request interceptor
api.interceptors.request.use(
    (config) => {
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// Response interceptor
api.interceptors.response.use(
    (response) => response.data,
    (error) => {
        // Global error handling with Toast notifications
        if (error.code === 'ECONNABORTED') {
            toast.error('Hết thời gian kết nối', {
                description: 'Máy chủ phản hồi quá lâu. Vui lòng thử lại.',
            });
        } else if (!error.response) {
            // Network error (no response from server)
            toast.error('Lỗi kết nối mạng', {
                description: 'Không thể kết nối đến máy chủ. Kiểm tra kết nối internet.',
            });
        } else {
            const status = error.response.status;
            if (status === 401) {
                toast.error('Phiên đăng nhập hết hạn', {
                    description: 'Vui lòng đăng nhập lại.',
                });
            } else if (status === 500) {
                toast.error('Lỗi máy chủ', {
                    description: 'Đã xảy ra lỗi phía máy chủ. Vui lòng thử lại sau.',
                });
            } else if (status === 404) {
                // Silent - let individual components handle 404
            } else if (status >= 400) {
                toast.error(`Lỗi yêu cầu (${status})`, {
                    description: error.response.data?.message || 'Đã xảy ra lỗi không xác định.',
                });
            }
        }
        return Promise.reject(error);
    }
);

export default api;
