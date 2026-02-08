import React from 'react';
import { motion } from 'framer-motion';
import { Card } from '../ui/Card';
import { Skeleton } from '../ui/Skeleton';
import { ChevronRight } from 'lucide-react';

// Import FusionSolar Assets
import flowGridImg from '../../assets/flow-grid.png';
import flowLoadImg from '../../assets/flow-load.png';
import flowPvImg from '../../assets/flow-pv.png';

interface EnergyFlowProps {
    gridPower?: number;       // kW
    loadPower?: number;       // kW
    pvPower?: number;         // kW
    loading?: boolean;
}

// --- Animated Arrow on Path ---
const AnimatedArrow: React.FC<{ pathD: string; color: string; active: boolean }> = ({ pathD, color, active }) => {
    if (!active) return null;
    return (
        <motion.polygon
            points="0,-5 8,0 0,5"
            fill={color}
            initial={{ offsetDistance: "0%" }}
            animate={{ offsetDistance: "100%" }}
            transition={{ repeat: Infinity, duration: 2, ease: "linear" }}
            style={{ offsetPath: `path("${pathD}")`, offsetRotate: "auto" } as React.CSSProperties}
        />
    );
};

export const EnergyFlow: React.FC<EnergyFlowProps> = ({
    gridPower = 0,
    loadPower = 0,
    pvPower = 0,
    loading = false
}) => {
    if (loading) {
        return (
            <Card>
                <div className="relative w-full h-[280px] flex items-center justify-center">
                    <Skeleton className="w-full h-full rounded-xl" />
                </div>
            </Card>
        );
    }

    const gridImporting = gridPower > 0;
    const pvExporting = pvPower > 0;

    const formatPower = (kw: number) => {
        const mw = kw / 1000;
        return mw.toLocaleString('en-US', { minimumFractionDigits: 3, maximumFractionDigits: 3 });
    };

    // Path definitions matching the triangular layout
    const gridToLoadPath = "M 120 195 L 240 115";
    const pvToLoadPath = "M 480 195 L 360 115";

    return (
        <Card className="overflow-visible bg-white">
            <div className="relative w-full" style={{ height: '300px' }}>

                {/* SVG Flow Lines Layer */}
                <svg
                    className="absolute inset-0 w-full h-full pointer-events-none"
                    viewBox="0 0 600 300"
                    preserveAspectRatio="xMidYMid meet"
                >
                    {/* Grid -> Load */}
                    {gridImporting && (
                        <>
                            <path d={gridToLoadPath} fill="none" stroke="#e5e7eb" strokeWidth="1" strokeDasharray="6 4" />
                            <AnimatedArrow pathD={gridToLoadPath} color="#3B82F6" active={gridImporting} />
                        </>
                    )}

                    {/* PV -> Load */}
                    {pvExporting && (
                        <>
                            <path d={pvToLoadPath} fill="none" stroke="#e5e7eb" strokeWidth="1" strokeDasharray="6 4" />
                            <AnimatedArrow pathD={pvToLoadPath} color="#22C55E" active={pvExporting} />
                        </>
                    )}
                </svg>

                {/* ========== LOAD NODE - Top Center ========== */}
                <div className="absolute" style={{ top: '8px', left: '50%', transform: 'translateX(-50%)' }}>
                    <div className="flex flex-col items-center">
                        {/* Power Value */}
                        <div className="flex items-baseline gap-1.5">
                            <span className="text-2xl font-bold text-slate-800 tabular-nums">{formatPower(loadPower)}</span>
                            <span className="text-sm text-slate-500">MW</span>
                            <span className="text-xs text-slate-400 ml-2">Công suất tiêu thụ</span>
                        </div>
                        {/* Vertical Line */}
                        <div className="w-0.5 h-8 bg-slate-200 my-1" />
                        {/* Image */}
                        <div className="w-32 h-24 flex items-center justify-center">
                            <img src={flowLoadImg} alt="Load" className="max-w-full max-h-full object-contain" />
                        </div>
                        {/* Label */}
                        <span className="text-sm font-medium text-slate-600">Tải</span>
                    </div>
                </div>

                {/* ========== GRID NODE - Bottom Left ========== */}
                <div className="absolute" style={{ bottom: '16px', left: '8%' }}>
                    <div className="flex flex-col items-start">
                        {/* Power Value */}
                        <div className="flex items-baseline gap-1.5 mb-0.5">
                            <span className="text-2xl font-bold text-slate-800 tabular-nums">{formatPower(gridPower)}</span>
                            <span className="text-sm text-slate-500">MW</span>
                        </div>
                        <span className="text-xs text-slate-400 mb-1">Công suất hiện tại</span>
                        {/* Image */}
                        <div className="w-28 h-24 flex items-center justify-center">
                            <img src={flowGridImg} alt="Grid" className="max-w-full max-h-full object-contain" />
                        </div>
                        {/* Label */}
                        <span className="text-sm font-medium text-slate-600">Lưới Điện</span>
                    </div>
                </div>

                {/* ========== PV NODE - Bottom Right ========== */}
                <div className="absolute" style={{ bottom: '16px', right: '8%' }}>
                    <div className="flex flex-col items-end">
                        {/* Power Value & Sub Label */}
                        <div className="flex items-baseline gap-1.5 mb-0.5">
                            <span className="text-2xl font-bold text-slate-800 tabular-nums">{formatPower(pvPower)}</span>
                            <span className="text-sm text-slate-500">MW</span>
                            <span className="text-xs text-slate-400 ml-2">Công suất đầu ra</span>
                        </div>
                        {/* PV Link Button */}
                        <button className="flex items-center gap-0 px-1.5 py-0.5 bg-emerald-50 text-emerald-600 text-xs font-bold rounded hover:bg-emerald-100 transition-colors mb-1">
                            PV <ChevronRight size={14} strokeWidth={3} />
                        </button>
                        {/* Image */}
                        <div className="w-32 h-20 flex items-center justify-center">
                            <img src={flowPvImg} alt="PV" className="max-w-full max-h-full object-contain" />
                        </div>
                        {/* Label */}
                        <span className="text-sm font-medium text-slate-600">PV</span>
                    </div>
                </div>
            </div>
        </Card>
    );
};
