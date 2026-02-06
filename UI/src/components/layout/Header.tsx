import React, { useEffect, useState } from 'react';
import { Bell, Search, User } from 'lucide-react';
import { cn } from '../../utils/cn';
import { WeatherWidget } from '../widgets/WeatherWidget';

export const Header: React.FC = () => {
    const [time, setTime] = useState(new Date());

    useEffect(() => {
        const timer = setInterval(() => setTime(new Date()), 1000);
        return () => clearInterval(timer);
    }, []);

    return (
        <header className="h-16 bg-white/80 backdrop-blur-md border-b border-slate-200 flex items-center justify-between px-6 sticky top-0 z-30">

            {/* Left: Breadcrumbs or Active Page Title (Future) */}
            <div>
                <h1 className="text-xl font-bold text-slate-900">
                    Tổng Quan Hệ Thống
                </h1>
            </div>

            {/* Right: Actions */}
            <div className="flex items-center gap-4">
                {/* Weather Widget */}
                <div className="hidden md:block">
                    <WeatherWidget />
                </div>

                {/* Clock */}
                <div className="hidden md:block text-right mr-4 border-l border-slate-200 pl-4 h-8 flex flex-col justify-center">
                    <p className="text-sm font-semibold text-slate-900 leading-none mb-0.5">{time.toLocaleTimeString()}</p>
                    <p className="text-xs text-slate-500 leading-none">{time.toLocaleDateString('vi-VN')}</p>
                </div>

                {/* Search Bar - Visual Only */}
                <div className="hidden md:flex items-center bg-slate-100 rounded-full px-4 py-1.5 border border-slate-200 focus-within:border-solar-500 transition-colors">
                    <Search size={16} className="text-slate-400" />
                    <input
                        type="text"
                        placeholder="Tìm kiếm thiết bị..."
                        className="bg-transparent border-none outline-none text-sm ml-2 text-slate-900 w-48 placeholder-slate-400"
                    />
                </div>

                {/* Notifications */}
                <button className="relative p-2 text-slate-400 hover:text-slate-900 transition-colors rounded-full hover:bg-slate-100">
                    <Bell size={20} />
                    <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full border border-white" />
                </button>

                {/* User Profile */}
                <div className="flex items-center gap-3 pl-4 border-l border-slate-200">
                    <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-solar-500 to-amber-300 flex items-center justify-center text-white font-bold text-xs shadow-sm">
                        AD
                    </div>
                    <div className="hidden lg:block">
                        <p className="text-sm font-medium text-slate-900">Admin</p>
                        <p className="text-xs text-slate-500">Quản Trị Viên</p>
                    </div>
                </div>
            </div>
        </header>
    );
};
