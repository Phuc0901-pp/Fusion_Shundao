import React from 'react';
import { motion } from 'framer-motion';
import { Zap } from 'lucide-react';
import { cn } from '../../utils/cn';
import type { StringData } from '../../types';

interface PVStringProps {
    data: StringData;
}

export const PVString: React.FC<PVStringProps> = React.memo(({ data }) => {
    const isActive = data.current > 0 && data.voltage > 0;

    return (
        <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            whileHover={{ scale: 1.05 }}
            className={cn(
                "relative p-3 rounded-xl border flex flex-col items-center justify-center gap-2 transition-all duration-300 cursor-pointer",
                isActive
                    ? "bg-gradient-to-br from-white to-slate-50 border-green-500/30 hover:border-green-500 shadow-sm"
                    : "bg-slate-100 border-slate-200 opacity-60 grayscale"
            )}
        >
            <span className="text-[10px] font-mono text-slate-500 uppercase tracking-wider truncate w-full text-center" title={data.id}>
                {data.id.split('-').pop()}
            </span>

            <div className="flex flex-col items-center">
                <span className={cn("text-base font-bold tabular-nums", isActive ? "text-slate-900" : "text-slate-400")}>
                    {data.current.toFixed(1)} <span className="text-[10px] text-slate-500">A</span>
                </span>
                <span className="text-[10px] text-slate-400 tabular-nums">
                    {data.voltage.toFixed(0)} V
                </span>
            </div>

            {isActive && (
                <Zap size={10} className="absolute top-1.5 right-1.5 text-solar-500 animate-pulse" />
            )}

            {/* Simple Progress Bar */}
            <div className="w-full h-1 bg-slate-100 rounded-full mt-2 overflow-hidden">
                <motion.div
                    className="h-full bg-emerald-500"
                    initial={{ width: 0 }}
                    animate={{ width: isActive ? `${Math.min(data.current * 5, 100)}%` : 0 }}
                    transition={{ duration: 1, ease: "easeOut" }}
                />
            </div>
        </motion.div>
    );
});
