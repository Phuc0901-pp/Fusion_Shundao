import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Lock, User, ArrowRight, Loader2, ShieldAlert } from 'lucide-react';
import { playHappyChime } from '../../utils/audio';
import logoImg from '../../assets/logov1.png';
import axios from 'axios';

const AUTH_API_URL = '/api/auth/login';

// Format seconds to MM:SS
const formatCountdown = (secs: number): string => {
    if (secs <= 0) return '00:00';
    const h = Math.floor(secs / 3600);
    const m = Math.floor((secs % 3600) / 60);
    const s = secs % 60;
    if (h > 0) return `${h}h ${String(m).padStart(2, '0')}m`;
    return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
};

export const LoginForm: React.FC = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');
    const [remaining, setRemaining] = useState<number | null>(null); // attempts left
    const [lockedUntil, setLockedUntil] = useState<number | null>(null); // Unix timestamp
    const [countdown, setCountdown] = useState(0); // seconds left
    const navigate = useNavigate();

    // Countdown timer when locked
    useEffect(() => {
        if (!lockedUntil) return;
        const tick = () => {
            const secsLeft = Math.max(0, lockedUntil - Math.floor(Date.now() / 1000));
            setCountdown(secsLeft);
            if (secsLeft === 0) {
                setLockedUntil(null);
                setError('');
                navigate('/lockdown', { replace: false });

            }
        };
        tick();
        const id = setInterval(tick, 1000);
        return () => clearInterval(id);
    }, [lockedUntil, navigate]);

    const isLocked = lockedUntil !== null && countdown > 0;

    const handleLogin = useCallback(async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setRemaining(null);

        if (!username || !password) {
            setError('Vui lòng nhập đầy đủ tên đăng nhập và mật khẩu.');
            return;
        }

        setIsLoading(true);

        try {
            const response = await axios.post(AUTH_API_URL, { username, password });
            const { token, username: uname, role, full_name } = response.data;

            // Unlock AudioContext and play welcome sound after user interaction
            playHappyChime();

            // Store JWT token and user info in sessionStorage
            sessionStorage.setItem('shundao_auth', 'true');
            sessionStorage.setItem('shundao_token', token);
            sessionStorage.setItem('shundao_user', JSON.stringify({ username: uname, role, full_name }));

            // Navigate to dashboard
            navigate('/overview', { replace: true });
        } catch (err: unknown) {
            if (axios.isAxiosError(err) && err.response) {
                const data = err.response.data;
                if (err.response.status === 418 && data?.unlock_at) {
                    // Exponential backoff lockout
                    setLockedUntil(data.unlock_at as number);
                    navigate('/lockdown', { replace: false });
                } else {
                    setError(data?.error || 'Đăng nhập thất bại. Vui lòng thử lại.');
                    if (data?.remaining !== undefined) {
                        setRemaining(data.remaining as number);
                    }
                }
            } else {
                setError('Không thể kết nối đến máy chủ. Vui lòng kiểm tra lại.');
            }
        } finally {
            setIsLoading(false);
        }
    }, [username, password, navigate]);

    return (
        <div className="relative z-10 w-full max-w-md mx-4 animate-in fade-in slide-in-from-bottom-8 duration-700">
            {/* Glassmorphism Card */}
            <div className="bg-white/10 backdrop-blur-xl border border-white/20 p-8 sm:p-10 rounded-3xl shadow-[0_8px_32px_0_rgba(0,0,0,0.36)] relative overflow-hidden">

                {/* Decorative top glow */}
                <div className="absolute top-0 left-1/2 -translate-x-1/2 w-3/4 h-1 bg-gradient-to-r from-transparent via-cyan-400 to-transparent opacity-50" />

                <div className="text-center mb-6">
                    <div className="inline-flex items-center justify-center w-40 h-40 mb-2 transform hover:scale-105 transition-transform duration-300">
                        <img
                            src={logoImg}
                            alt="Raitek Logo"
                            className="w-full h-full object-contain filter hover:rotate-6 transition-transform duration-300 drop-shadow-[0_0_16px_rgba(255,255,255,0.8)]"
                        />
                    </div>

                    <h1 className="text-2xl font-bold text-white tracking-tight mb-2">
                        Raitek Join Stock Company
                    </h1>

                    <p className="text-blue-200/80 text-sm font-medium">
                        Hệ Thống Giám Sát Năng Lượng Của Dự Án Shundao
                    </p>
                </div>

                <form onSubmit={handleLogin} className="space-y-5">
                    {/* Username Input */}
                    <div className="space-y-1.5">
                        <label className="text-sm font-medium text-blue-100 flex items-center gap-2">
                            <User size={14} className="opacity-70" /> Tên đăng nhập
                        </label>
                        <div className="relative group">
                            <input
                                type="text"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                className="w-full bg-slate-900/50 border border-slate-700/50 text-white rounded-xl px-4 py-3 pl-11 focus:outline-none focus:ring-2 focus:ring-cyan-500/50 focus:border-cyan-400/50 transition-all placeholder:text-slate-500"
                                placeholder="Nhập tên đăng nhập"
                                required
                            />
                            <div className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-cyan-400 transition-colors">
                                <User size={18} />
                            </div>
                        </div>
                    </div>

                    {/* Password Input */}
                    <div className="space-y-1.5">
                        <label className="text-sm font-medium text-blue-100 flex items-center gap-2">
                            <Lock size={14} className="opacity-70" /> Mật khẩu
                        </label>
                        <div className="relative group">
                            <input
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                className="w-full bg-slate-900/50 border border-slate-700/50 text-white rounded-xl px-4 py-3 pl-11 focus:outline-none focus:ring-2 focus:ring-cyan-500/50 focus:border-cyan-400/50 transition-all placeholder:text-slate-500"
                                placeholder="Nhập mật khẩu"
                                required
                            />
                            <div className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-cyan-400 transition-colors">
                                <Lock size={18} />
                            </div>
                        </div>
                    </div>

                    {/* Remaining attempts warning */}
                    {remaining !== null && remaining > 0 && remaining <= 3 && (
                        <div className="flex items-center gap-2 bg-amber-500/10 border border-amber-500/30 text-amber-200 text-xs p-2.5 rounded-lg animate-in fade-in zoom-in duration-200">
                            <ShieldAlert size={14} className="shrink-0" />
                            <span>Cảnh báo: Còn <strong>{remaining}</strong> lần thử trước khi bị tạm khóa.</span>
                        </div>
                    )}

                    {/* Error Message */}
                    {error && (
                        <div className="bg-red-500/10 border border-red-500/30 text-red-200 text-sm p-3 rounded-lg text-center animate-in fade-in zoom-in duration-200">
                            {error}
                        </div>
                    )}

                    {/* Submit Button */}
                    <button
                        type="submit"
                        disabled={isLoading || isLocked}
                        className="w-full mt-2 bg-gradient-to-r from-blue-500 to-cyan-500 hover:from-blue-400 hover:to-cyan-400 text-white font-bold py-3.5 px-4 rounded-xl shadow-lg shadow-cyan-500/25 flex items-center justify-center gap-2 transition-all transform hover:-translate-y-0.5 active:translate-y-0 disabled:opacity-70 disabled:pointer-events-none group"
                    >
                        {isLoading ? (
                            <>
                                <Loader2 size={20} className="animate-spin" />
                                <span>Đang xác thực...</span>
                            </>
                        ) : isLocked ? (
                            <>
                                <ShieldAlert size={20} className="animate-pulse" />
                                <span>Mở khóa sau: <strong className="tabular-nums">{formatCountdown(countdown)}</strong></span>
                            </>
                        ) : (
                            <>
                                <span>Đăng nhập hệ thống</span>
                                <ArrowRight size={20} className="group-hover:translate-x-1 transition-transform" />
                            </>
                        )}
                    </button>
                </form>

                <div className="mt-8 text-center border-t border-white/10 pt-6">
                    <p className="text-xs text-blue-200/50">
                        © 2026 Monitoring Shundao System by Raitek
                    </p>
                </div>
            </div>
        </div>
    );
};
