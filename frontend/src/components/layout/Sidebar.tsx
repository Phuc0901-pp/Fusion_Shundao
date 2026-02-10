import React from 'react';
import { LayoutDashboard, Zap, Menu } from 'lucide-react';
import { cn } from '../../utils/cn';

interface SidebarProps {
    isOpen: boolean;
    setIsOpen: (isOpen: boolean) => void;
}

const MENU_ITEMS = [
    { icon: LayoutDashboard, label: 'Tổng Quan', id: 'overview' }
];

import { useNavigate, useLocation } from 'react-router-dom';

export const Sidebar: React.FC<SidebarProps> = ({ isOpen, setIsOpen }) => {
    const navigate = useNavigate();
    const location = useLocation();
    const currentPath = location.pathname.substring(1) || 'overview';

    const handleNavigation = (id: string) => {
        navigate(`/${id}`);
    };

    return (
        <aside
            className={cn(
                "fixed left-0 top-0 z-40 h-screen bg-white border-r border-slate-200 transition-all duration-300 ease-in-out shadow-lg",
                isOpen ? "w-64" : "w-20"
            )}
        >
            {/* Logo Area */}
            <div className="flex items-center justify-center h-16 border-b border-slate-200 relative">
                <div className="flex items-center gap-2 text-solar-500 font-bold text-xl overflow-hidden whitespace-nowrap cursor-pointer" onClick={() => navigate('/')}>
                    <Zap className="w-8 h-8 fill-solar-500" />
                    <span className={cn("transition-opacity duration-300", isOpen ? "opacity-100" : "opacity-0 hidden")}>
                        FusionSolar
                    </span>
                </div>

                {/* Toggle Button */}
                <button
                    onClick={() => setIsOpen(!isOpen)}
                    className="absolute -right-3 top-6 bg-white p-1 rounded-full text-slate-400 hover:text-slate-900 border border-slate-200 hover:border-solar-500 transition-colors shadow-sm"
                >
                    <Menu size={14} />
                </button>
            </div>

            {/* Menu Items */}
            <nav className="mt-8 px-4 flex flex-col gap-2">
                {MENU_ITEMS.map((item) => {
                    const isActive = currentPath.includes(item.id);
                    return (
                        <button
                            key={item.id}
                            onClick={() => handleNavigation(item.id)}
                            className={cn(
                                "flex items-center gap-3 p-3 rounded-lg transition-all duration-200 group relative overflow-hidden",
                                isActive
                                    ? "bg-solar-50/50 text-solar-600 border-l-2 border-solar-500"
                                    : "text-slate-500 hover:bg-slate-50 hover:text-slate-900"
                            )}
                        >
                            <item.icon size={22} className={cn("min-w-[22px]", isActive && "fill-solar-500/20")} />

                            <span className={cn(
                                "font-medium transition-all duration-300 whitespace-nowrap",
                                isOpen ? "opacity-100 translate-x-0" : "opacity-0 -translate-x-4 absolute left-12"
                            )}>
                                {item.label}
                            </span>

                            {/* Hover Glow Effect */}
                            <div className="absolute inset-0 bg-solar-500/5 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
                        </button>
                    );
                })}
            </nav>

            {/* Footer Status */}
            <div className="absolute bottom-4 left-0 w-full px-4">
                <div className={cn(
                    "p-3 rounded-xl bg-slate-50 border border-slate-100 flex items-center gap-3",
                    !isOpen && "justify-center p-2"
                )}>
                    <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                    {isOpen && (
                        <div>
                            <p className="text-xs text-slate-400">Trạng thái hệ thống</p>
                            <p className="text-xs font-bold text-green-400">HOẠT ĐỘNG TỐT</p>
                        </div>
                    )}
                </div>
            </div>
        </aside>
    );
};
