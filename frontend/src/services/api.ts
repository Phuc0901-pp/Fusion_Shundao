import axios from 'axios';

// Create Axios instance with default config
const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL || '/api', // Use relative path for proxy support
    headers: {
        'Content-Type': 'application/json',
    },
    timeout: 10000, // 10 seconds timeout
});

// Request interceptor
api.interceptors.request.use(
    (config) => {
        // You can add auth tokens here if needed
        // const token = localStorage.getItem('token');
        // if (token) {
        //   config.headers.Authorization = `Bearer ${token}`;
        // }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// Response interceptor
api.interceptors.response.use(
    (response) => response.data, // Return data directly for easier consumption
    (error) => {
        // Global error handling
        if (error.response?.status === 401) {
            // Handle unauthorized (redirect to login)
            console.warn('Unauthorized access');
        }
        return Promise.reject(error);
    }
);

export default api;
