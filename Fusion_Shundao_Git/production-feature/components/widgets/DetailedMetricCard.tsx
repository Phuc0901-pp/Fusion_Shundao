import React from 'react';
import { motion } from 'framer-motion';
import { cn } from '../../utils/cn';
import { Skeleton } from '../ui/Skeleton';
import type { LucideIcon } from 'lucide-react';

interface SiteData {
    id: string;
    name: string;
    value: number;
}

interface DetailedMetricCardProps {
    title: string;
    unit: string;
    icon: LucideIcon;
    color: string;
    totalValue: number;
    sites: SiteData[];
    loading?: boolean;
    delay?: number;
}

const colorMap: Record<string, { icon: string; bg: string; border: string }> = {
    solar: { icon: 'text-yellow-600', bg: 'bg-yellow-50', border: 'border-yellow-100' },
    blue: { icon: 'text-blue-600', bg: 'bg-blue-50', border: 'border-blue-100' },
    green: { icon: 'text-green-600', bg: 'bg-green-50', border: 'border-green-100' },
    slate: { icon: 'text-slate-600', bg: 'bg-slate-50', border: 'border-slate-100' },
    orange: { icon: 'text-orange-600', bg: 'bg-orange-50', border: 'border-orange-100' },
    emerald: { icon: 'text-emerald-600', bg: 'bg-emerald-50', border: 'border-emerald-100' },
};

export const DetailedMetricCard: React.FC<DetailedMetricCardProps> = ({
    title,
    unit,
    icon: Icon,
    color,
    totalValue,
    sites,
    loading = false,
    delay = 0
}) => {
    if (loading) {
        return (
            <div className="h-[200px] bg-white rounded-xl border border-slate-200 p-4 flex flex-col gap-4">
                <div className="flex items-center gap-3">
                    <Skeleton className="h-10 w-10 rounded-lg" />
                    <Skeleton className="h-6 w-32" />
                </div>
                <div className="space-y-2">
                    <Skeleton className="h-4 w-full" />
                    <Skeleton className="h-4 w-full" />
                    <Skeleton className="h-4 w-full" />
                </div>
            </div>
        );
    }

    const theme = colorMap[color] || colorMap.solar;

    return (
        <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: delay * 0.1, duration: 0.4 }}
            className="h-full"
        >
            <div className="bg-white rounded-xl border border-slate-200 shadow-sm hover:shadow-md transition-shadow duration-300 overflow-hidden h-full flex flex-col">
                {/* Header & Total */}
                <div className="p-4 border-b border-slate-100">
                    <div className="flex items-start justify-between mb-2">
                        <div className="flex items-center gap-3">
                            <div className={cn("p-2 rounded-lg", theme.bg, theme.icon)}>
                                <Icon size={20} />
                            </div>
                            <h3 className="font-semibold text-slate-700">{title}</h3>
                        </div>
                    </div>

                    <div className="flex items-baseline gap-2 mt-1 px-1">
                        <span className="text-2xl font-bold text-slate-800 tracking-tight">
                            {totalValue.toLocaleString('en-US', { maximumFractionDigits: 2 })}
                        </span>
                        <span className="text-sm font-medium text-slate-400">{unit}</span>
                    </div>
                </div>

                {/* Sites Breakdown */}
                <div className="p-3 bg-slate-50/50 flex-grow flex flex-col justify-center gap-2">
                    {sites.map((site) => (
                        <div key={site.id} className="flex justify-between items-center bg-white p-2.5 rounded-lg border border-slate-100 shadow-sm">
                            <div className="flex items-center gap-2">
                                <div className={cn("w-1.5 h-1.5 rounded-full", theme.icon.replace('text-', 'bg-'))}></div>
                                <span className="text-sm font-medium text-slate-600">{site.name}</span>
                            </div>
                            <div className="flex items-baseline gap-1">
                                <span className="text-sm font-bold text-slate-700">
                                    {site.value.toLocaleString('en-US', { maximumFractionDigits: 2 })}
                                </span>
                                <span className="text-[10px] text-slate-400">{unit}</span>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </motion.div>
    );
};
