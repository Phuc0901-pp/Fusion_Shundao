import React, { useEffect, useState } from 'react';
// import { cn } from '../../utils/cn';
import { WeatherWidget } from '../widgets/WeatherWidget';

import logo from '../../assets/LOGO.png';

export const Header: React.FC = () => {
    const [time, setTime] = useState(new Date());

    useEffect(() => {
        const timer = setInterval(() => setTime(new Date()), 1000);
        return () => clearInterval(timer);
    }, []);

    return (
        <header className="h-16 bg-white/80 backdrop-blur-md border-b border-slate-200 flex items-center justify-between px-6 sticky top-0 z-30">

            {/* Left: Branding */}
            <div className="flex items-center gap-3">
                <img src={logo} alt="Shundao Solar" className="h-10 w-auto" />
                <h1 className="text-xl font-bold text-slate-900">
                    SHUNDAO SOLAR
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
