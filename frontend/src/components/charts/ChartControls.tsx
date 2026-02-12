import React from 'react';
import { Calendar, Clock, Filter, Check } from 'lucide-react';
import { cn } from '../../utils/cn';

export type ViewMode = 'day' | 'month';

interface ChartControlsProps {
    viewMode: ViewMode;
    onViewModeChange: (mode: ViewMode) => void;
    visibleSites: { site1: boolean; site2: boolean };
    onToggleSite: (site: 'site1' | 'site2') => void;
}

export const ChartControls: React.FC<ChartControlsProps> = ({
    viewMode,
    onViewModeChange,
    visibleSites,
    onToggleSite
}) => {
    return (
        <div className="flex flex-wrap items-center justify-between gap-4 mb-4">
            {/* View Mode Toggle */}
            <div className="inline-flex bg-slate-100 rounded-lg p-1">
                <button
                    onClick={() => onViewModeChange('day')}
                    className={cn(
                        "px-3 py-1.5 text-sm font-medium rounded-md flex items-center gap-1.5 transition-all duration-200",
                        viewMode === 'day'
                            ? "bg-white text-slate-800 shadow-sm ring-1 ring-slate-200"
                            : "text-slate-500 hover:text-slate-700 hover:bg-slate-200/50"
                    )}
                >
                    <Clock size={14} />
                    Hôm nay
                </button>
                <button
                    onClick={() => onViewModeChange('month')}
                    className={cn(
                        "px-3 py-1.5 text-sm font-medium rounded-md flex items-center gap-1.5 transition-all duration-200",
                        viewMode === 'month'
                            ? "bg-white text-slate-800 shadow-sm ring-1 ring-slate-200"
                            : "text-slate-500 hover:text-slate-700 hover:bg-slate-200/50"
                    )}
                >
                    <Calendar size={14} />
                    Theo tháng
                </button>
            </div>

            {/* Site Filters */}
            <div className="flex items-center gap-2">
                <div className="flex items-center gap-2 mr-2 text-sm text-slate-500">
                    <Filter size={14} />
                    <span className="hidden sm:inline">Hiển thị:</span>
                </div>

                <button
                    onClick={() => onToggleSite('site1')}
                    className={cn(
                        "px-3 py-1.5 text-sm font-medium rounded-full flex items-center gap-1.5 transition-all border",
                        visibleSites.site1
                            ? "bg-green-50 text-green-700 border-green-200"
                            : "bg-slate-50 text-slate-400 border-slate-200 opacity-60 hover:opacity-100"
                    )}
                >
                    <div className={cn(
                        "w-2 h-2 rounded-full transition-colors",
                        visibleSites.site1 ? "bg-green-500" : "bg-slate-400"
                    )} />
                    Shundao 1
                    {visibleSites.site1 && <Check size={12} className="ml-1" />}
                </button>

                <button
                    onClick={() => onToggleSite('site2')}
                    className={cn(
                        "px-3 py-1.5 text-sm font-medium rounded-full flex items-center gap-1.5 transition-all border",
                        visibleSites.site2
                            ? "bg-blue-50 text-blue-700 border-blue-200"
                            : "bg-slate-50 text-slate-400 border-slate-200 opacity-60 hover:opacity-100"
                    )}
                >
                    <div className={cn(
                        "w-2 h-2 rounded-full transition-colors",
                        visibleSites.site2 ? "bg-blue-500" : "bg-slate-400"
                    )} />
                    Shundao 2
                    {visibleSites.site2 && <Check size={12} className="ml-1" />}
                </button>
            </div>
        </div>
    );
};
