import React, { useEffect, useState } from 'react';
import { WeatherWidget } from '../widgets/WeatherWidget';
import { MapPin, Clock, User } from 'lucide-react';
import { useWeather } from '../../hooks/useWeather';
import { cn } from '../../utils/cn';
import logo from '../../assets/LOGO.png';

export const Header: React.FC = () => {
    const [time, setTime] = useState(new Date());
    const { data: weather } = useWeather();
    const [scrolled, setScrolled] = useState(false);

    useEffect(() => {
        const timer = setInterval(() => setTime(new Date()), 1000);
        const handleScroll = () => setScrolled(window.scrollY > 0);
        window.addEventListener('scroll', handleScroll);
        return () => {
            clearInterval(timer);
            window.removeEventListener('scroll', handleScroll);
        };
    }, []);

    return (
        <header className={cn(
            "fixed top-0 left-0 right-0 z-40 transition-all duration-200 h-16 flex items-center justify-between px-6 lg:px-8",
            scrolled ? "bg-white shadow-sm border-b border-slate-100" : "bg-white border-b border-slate-100"
        )}>
            {/* Left: Brand */}
            <div className="flex items-center gap-3">
                <img src={logo} alt="Shundao Solar" className="h-8 w-auto" />
                <div className="hidden md:block h-8 w-px bg-slate-200 mx-2"></div>
                <div className="hidden md:block">
                    <h1 className="text-sm font-bold text-slate-800 tracking-wide uppercase">
                        Shundao Solar
                    </h1>
                    <p className="text-[10px] text-slate-500 font-medium">
                        Raitek version 1.0
                    </p>
                </div>
            </div>

            {/* Right: Info Group */}
            <div className="flex items-center gap-6">

                {/* Weather */}
                <div className="hidden lg:block">
                    <WeatherWidget className="bg-transparent border-none shadow-none p-0 gap-2" />
                </div>

                {/* Divider */}
                <div className="hidden lg:block h-5 w-px bg-slate-200"></div>

                {/* Location */}
                <div className="hidden md:flex items-center gap-2 max-w-[200px] xl:max-w-[300px]" title={weather?.locationName}>
                    <MapPin size={16} className="text-slate-400 shrink-0" />
                    <div className="flex flex-col leading-tight overflow-hidden">
                        <span className="text-xs font-semibold text-slate-700 truncate">
                            {weather?.locationName || "Đang định vị..."}
                        </span>
                        <span className="text-[10px] text-slate-500 truncate">
                            Vị trí hiện tại
                        </span>
                    </div>
                </div>

                {/* Divider */}
                <div className="hidden md:block h-5 w-px bg-slate-200"></div>

                {/* Time */}
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

                {/* Divider */}
                <div className="h-5 w-px bg-slate-200"></div>

                {/* User & Actions */}
                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-3 cursor-pointer group">
                        <div className="w-8 h-8 rounded-full bg-slate-100 border border-slate-200 flex items-center justify-center text-slate-600 group-hover:bg-slate-200 transition-colors">
                            <User size={16} />
                        </div>
                        <div className="hidden xl:block text-left leading-none">
                            <p className="text-xs font-bold text-slate-700 group-hover:text-slate-900">Admin</p>
                        </div>
                    </div>
                </div>
            </div>
        </header>
    );
};
