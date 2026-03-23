import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

const REFRESH_URL = '/api/auth/refresh';
const REFRESH_INTERVAL_MS = 10 * 60 * 1000; // every 10 minutes

/**
 * useTokenRefresh - silently refreshes JWT before it expires.
 * Runs automatically every 10 minutes while the user is logged in.
 * On failure (expired / outside operating hours), logs out the user.
 */
export const useTokenRefresh = () => {
    const navigate = useNavigate();

    useEffect(() => {
        const refresh = async () => {
            const token = sessionStorage.getItem('shundao_token');
            if (!token) return;

            // Bot-check: a simple timestamp-based challenge
            const botCheck = btoa(`shundao-${Date.now()}-${Math.random().toString(36).slice(2)}`);

            try {
                const response = await axios.post(REFRESH_URL, {}, {
                    headers: {
                        Authorization: `Bearer ${token}`,
                        'X-Shundao-Bot-Check': botCheck,
                    },
                });
                const { token: newToken } = response.data;
                if (newToken) {
                    sessionStorage.setItem('shundao_token', newToken);
                }
            } catch (err) {
                if (axios.isAxiosError(err)) {
                    if (err.response?.status === 418) {
                        // Counter-hack – redirect to Lockdown page
                        navigate('/lockdown', { replace: true });
                        return;
                    }
                    // Token expired or outside hours – logout
                    sessionStorage.removeItem('shundao_auth');
                    sessionStorage.removeItem('shundao_token');
                    sessionStorage.removeItem('shundao_user');
                    navigate('/login', { replace: true });
                }
            }
        };

        const interval = setInterval(refresh, REFRESH_INTERVAL_MS);
        return () => clearInterval(interval);
    }, [navigate]);
};
