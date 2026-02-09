import React, { useEffect, useState } from 'react';
import { WeatherWidget } from '../widgets/WeatherWidget';
import { Sun } from 'lucide-react';

import logo from '../../assets/LOGO.png';

export const Header: React.FC = () => {
    const [time, setTime] = useState(new Date());

    useEffect(() => {
        const timer = setInterval(() => setTime(new Date()), 1000);
        return () => clearInterval(timer);
    }, []);

    return (
        <header className="h-16 glass-strong border-b border-white/30 flex items-center justify-between px-6 sticky top-0 z-30 shadow-sm">
            {/* Left: Branding */}
            <div className="flex items-center gap-3">
                <div className="relative">
                    <img src={logo} alt="Shundao Solar" className="h-10 w-auto relative z-10" />
                    <div className="absolute inset-0 bg-solar-400 blur-lg opacity-30 animate-pulse" />
                </div>
                <div>
                    <h1 className="text-xl font-bold text-gradient">
                        SHUNDAO SOLAR
                    </h1>
                    <p className="text-xs text-slate-500 -mt-0.5">Solar Monitoring System</p>
                </div>
            </div>

            {/* Right: Actions */}
            <div className="flex items-center gap-4">
                {/* Weather Widget */}
                <div className="hidden md:block">
                    <WeatherWidget />
                </div>

                {/* Clock */}
                <div className="hidden md:flex flex-col items-end border-l border-slate-200/50 pl-4 h-8 justify-center">
                    <p className="text-sm font-semibold text-slate-800 leading-none mb-0.5 tabular-nums">
                        {time.toLocaleTimeString()}
                    </p>
                    <p className="text-xs text-slate-500 leading-none">
                        {time.toLocaleDateString('vi-VN', { weekday: 'short', day: '2-digit', month: '2-digit' })}
                    </p>
                </div>

                {/* User Profile */}
                <div className="flex items-center gap-3 pl-4 border-l border-slate-200/50">
                    <div className="relative">
                        <div className="w-9 h-9 rounded-full bg-gradient-to-tr from-solar-500 via-amber-400 to-orange-400 flex items-center justify-center text-white font-bold text-xs shadow-md">
                            <Sun size={18} />
                        </div>
                        <div className="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-green-500 rounded-full border-2 border-white animate-pulse-glow" />
                    </div>
                    <div className="hidden lg:block">
                        <p className="text-sm font-medium text-slate-800">Admin</p>
                        <p className="text-xs text-slate-500">Đang hoạt động</p>
                    </div>
                </div>
            </div>
        </header>
    );
};
