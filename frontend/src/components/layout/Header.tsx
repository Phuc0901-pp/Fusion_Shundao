import React, { useEffect, useRef, useState } from 'react';
import { WeatherWidget } from '../widgets/WeatherWidget';
import { MapPin, Clock, User, LogOut, KeyRound, Info, ChevronDown, Loader2 } from 'lucide-react';
import { useWeather } from '../../hooks/useWeather';
import { cn } from '../../utils/cn';
import logo from '../../assets/LOGO.png';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

const CHANGE_PWD_URL = '/api/auth/change-password';

// ── ChangePasswordModal ─────────────────────────────────────────────────────
interface ChangePasswordModalProps { onClose: () => void; }
const ChangePasswordModal: React.FC<ChangePasswordModalProps> = ({ onClose }) => {
    const [oldPwd, setOldPwd] = useState('');
    const [newPwd, setNewPwd] = useState('');
    const [confirmPwd, setConfirmPwd] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        if (newPwd !== confirmPwd) { setError('Mật khẩu mới không khớp.'); return; }
        if (newPwd.length < 6) { setError('Mật khẩu mới phải có ít nhất 6 ký tự.'); return; }
        setLoading(true);
        try {
            const token = sessionStorage.getItem('shundao_token');
            await axios.post(CHANGE_PWD_URL, { old_password: oldPwd, new_password: newPwd }, {
                headers: { Authorization: `Bearer ${token}` },
            });
            setSuccess(true);
        } catch (err) {
            setError(axios.isAxiosError(err) && err.response ? err.response.data?.error || 'Thất bại.' : 'Không thể kết nối máy chủ.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm" onClick={onClose}>
            <div className="bg-white rounded-2xl shadow-2xl p-8 w-full max-w-sm" onClick={e => e.stopPropagation()}>
                <h2 className="text-xl font-bold text-slate-800 mb-6 flex items-center gap-2">
                    <KeyRound size={20} className="text-blue-500" /> Đổi mật khẩu
                </h2>
                {success ? (
                    <div className="text-center py-4">
                        <p className="text-green-600 font-bold text-lg">✅ Đổi mật khẩu thành công!</p>
                        <button onClick={onClose} className="mt-4 px-6 py-2 bg-blue-500 text-white rounded-lg text-sm font-medium">Đóng</button>
                    </div>
                ) : (
                    <form onSubmit={handleSubmit} className="space-y-4">
                        {['Mật khẩu cũ', 'Mật khẩu mới', 'Xác nhận mật khẩu mới'].map((label, i) => (
                            <div key={i}>
                                <label className="text-xs font-medium text-slate-600">{label}</label>
                                <input type="password"
                                    value={[oldPwd, newPwd, confirmPwd][i]}
                                    onChange={e => [setOldPwd, setNewPwd, setConfirmPwd][i](e.target.value)}
                                    required
                                    className="w-full mt-1 border border-slate-200 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
                                />
                            </div>
                        ))}
                        {error && <p className="text-red-500 text-xs bg-red-50 p-2 rounded-lg">{error}</p>}
                        <div className="flex gap-3 mt-2">
                            <button type="button" onClick={onClose} className="flex-1 py-2.5 border border-slate-200 rounded-lg text-sm text-slate-600 hover:bg-slate-50">Hủy</button>
                            <button type="submit" disabled={loading} className="flex-1 py-2.5 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-sm font-medium disabled:opacity-60 flex items-center justify-center gap-2">
                                {loading && <Loader2 size={14} className="animate-spin" />} Lưu
                            </button>
                        </div>
                    </form>
                )}
            </div>
        </div>
    );
};

// ── Helper: read user from sessionStorage ────────────────────────────────────
const getUserInfo = () => {
    try {
        const raw = sessionStorage.getItem('shundao_user');
        if (raw) return JSON.parse(raw) as { username: string; role: string; full_name: string };
    } catch (_) { /* ignore */ }
    return { username: 'Admin', role: 'viewer', full_name: '' };
};

// ── Header ────────────────────────────────────────────────────────────────────
export const Header: React.FC = () => {
    const [time, setTime] = useState(new Date());
    const { data: weather } = useWeather();
    const [scrolled, setScrolled] = useState(false);
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const [showChangePassword, setShowChangePassword] = useState(false);
    const [showUserInfo, setShowUserInfo] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);
    const navigate = useNavigate();
    const user = getUserInfo();

    useEffect(() => {
        const timer = setInterval(() => setTime(new Date()), 1000);
        const handleScroll = () => setScrolled(window.scrollY > 0);
        window.addEventListener('scroll', handleScroll);
        const handleClickOutside = (e: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
                setDropdownOpen(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => {
            clearInterval(timer);
            window.removeEventListener('scroll', handleScroll);
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, []);

    const handleLogout = () => {
        sessionStorage.removeItem('shundao_auth');
        sessionStorage.removeItem('shundao_token');
        sessionStorage.removeItem('shundao_user');
        navigate('/login', { replace: true });
    };

    return (
        <>
            <header className={cn(
                "fixed top-0 left-0 right-0 z-40 transition-all duration-200 h-16 flex items-center justify-between px-6 lg:px-8",
                scrolled ? "bg-white shadow-sm border-b border-slate-100" : "bg-white border-b border-slate-100"
            )}>
                {/* Brand */}
                <div className="flex items-center gap-3">
                    <img src={logo} alt="Shundao Solar" className="h-8 w-auto" />
                    <div className="hidden md:block h-8 w-px bg-slate-200 mx-2" />
                    <div className="hidden md:block">
                        <h1 className="text-sm font-bold text-slate-800 tracking-wide uppercase">Shundao Solar</h1>
                        <p className="text-[10px] text-slate-500 font-medium">Version 2.5</p>
                    </div>
                </div>

                {/* Right */}
                <div className="flex items-center gap-6">
                    <div className="hidden lg:block">
                        <WeatherWidget className="bg-transparent border-none shadow-none p-0 gap-2" />
                    </div>
                    <div className="hidden lg:block h-5 w-px bg-slate-200" />
                    <div className="hidden md:flex items-center gap-2 max-w-[200px] xl:max-w-[300px]" title={weather?.locationName}>
                        <MapPin size={16} className="text-slate-400 shrink-0" />
                        <div className="flex flex-col leading-tight overflow-hidden">
                            <span className="text-xs font-semibold text-slate-700 truncate">{weather?.locationName || "Đang định vị..."}</span>
                            <span className="text-[10px] text-slate-500">Vị trí hiện tại</span>
                        </div>
                    </div>
                    <div className="hidden md:block h-5 w-px bg-slate-200" />
                    <div className="hidden md:flex items-center gap-3">
                        <Clock size={16} className="text-slate-400" />
                        <div className="flex flex-col items-end leading-none">
                            <span className="text-sm font-bold text-slate-800 tabular-nums">
                                {time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                            </span>
                            <span className="text-[10px] text-slate-500">
                                {time.toLocaleDateString('vi-VN', { weekday: 'short', day: '2-digit', month: '2-digit' })}
                            </span>
                        </div>
                    </div>
                    <div className="h-5 w-px bg-slate-200" />

                    {/* User Dropdown */}
                    <div className="relative" ref={dropdownRef}>
                        <button
                            onClick={() => setDropdownOpen(o => !o)}
                            className="flex items-center gap-2 cursor-pointer group rounded-xl px-2 py-1.5 hover:bg-slate-50 transition-colors"
                        >
                            <div className="w-8 h-8 rounded-full bg-blue-50 border border-blue-200 flex items-center justify-center text-blue-600 group-hover:bg-blue-100 transition-colors">
                                <User size={16} />
                            </div>
                            <div className="hidden xl:block text-left leading-none">
                                <p className="text-xs font-bold text-slate-700 group-hover:text-slate-900">{user.full_name || user.username}</p>
                                <p className="text-[10px] text-slate-400 capitalize">{user.role}</p>
                            </div>
                            <ChevronDown size={14} className={cn("text-slate-400 transition-transform duration-200", dropdownOpen && "rotate-180")} />
                        </button>

                        {dropdownOpen && (
                            <div className="absolute right-0 top-full mt-2 w-56 bg-white border border-slate-100 rounded-2xl shadow-xl overflow-hidden z-50 animate-in fade-in slide-in-from-top-2 duration-200">
                                <div className="px-4 py-3 bg-gradient-to-r from-blue-50 to-slate-50 border-b border-slate-100">
                                    <p className="text-sm font-bold text-slate-800">{user.full_name || user.username}</p>
                                    <p className="text-xs text-slate-500 capitalize">{user.role}</p>
                                </div>
                                <div className="p-1.5 space-y-0.5">
                                    <button onClick={() => { setShowUserInfo(true); setDropdownOpen(false); }}
                                        className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-slate-600 hover:bg-blue-50 hover:text-blue-600 transition-colors text-sm">
                                        <Info size={16} /> Thông tin tài khoản
                                    </button>
                                    <button onClick={() => { setShowChangePassword(true); setDropdownOpen(false); }}
                                        className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-slate-600 hover:bg-amber-50 hover:text-amber-600 transition-colors text-sm">
                                        <KeyRound size={16} /> Đổi mật khẩu
                                    </button>
                                    <div className="border-t border-slate-100 pt-1 mt-1">
                                        <button onClick={handleLogout}
                                            className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-red-500 hover:bg-red-50 transition-colors text-sm font-medium">
                                            <LogOut size={16} /> Đăng xuất
                                        </button>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </header>

            {/* Modals */}
            {showChangePassword && <ChangePasswordModal onClose={() => setShowChangePassword(false)} />}
            {showUserInfo && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm" onClick={() => setShowUserInfo(false)}>
                    <div className="bg-white rounded-2xl shadow-2xl p-8 w-full max-w-sm" onClick={e => e.stopPropagation()}>
                        <h2 className="text-xl font-bold text-slate-800 mb-6 flex items-center gap-2">
                            <Info size={20} className="text-blue-500" /> Thông tin tài khoản
                        </h2>
                        <div className="space-y-4">
                            <div className="flex items-center gap-4 p-4 bg-blue-50 rounded-xl">
                                <div className="w-14 h-14 rounded-full bg-blue-100 border-2 border-blue-200 flex items-center justify-center text-blue-600">
                                    <User size={28} />
                                </div>
                                <div>
                                    <p className="font-bold text-slate-800">{user.full_name || user.username}</p>
                                    <p className="text-sm text-slate-500">@{user.username}</p>
                                </div>
                            </div>
                            <div className="grid grid-cols-2 gap-3 text-sm">
                                <div className="bg-slate-50 rounded-xl p-3">
                                    <p className="text-[10px] text-slate-400 uppercase font-medium mb-1">Vai trò</p>
                                    <p className="font-bold text-slate-700 capitalize">{user.role}</p>
                                </div>
                                <div className="bg-slate-50 rounded-xl p-3">
                                    <p className="text-[10px] text-slate-400 uppercase font-medium mb-1">Hệ thống</p>
                                    <p className="font-bold text-green-600">Hoạt động</p>
                                </div>
                            </div>
                        </div>
                        <button onClick={() => setShowUserInfo(false)} className="mt-6 w-full py-2.5 bg-slate-100 hover:bg-slate-200 rounded-xl text-sm text-slate-600 font-medium transition-colors">
                            Đóng
                        </button>
                    </div>
                </div>
            )}
        </>
    );
};
