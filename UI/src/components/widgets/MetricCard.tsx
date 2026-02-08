import React from 'react';
import { motion } from 'framer-motion';
import { cn } from '../../utils/cn';
import { Card } from '../ui/Card';
import { Skeleton } from '../ui/Skeleton';

interface MetricCardProps {
    title: string;
    value: string | number;
    unit?: string;
    icon: React.ElementType;
    trend?: number;
    color?: string;
    loading?: boolean;
}

const colorMap: Record<string, string> = {
    solar: 'text-solar-500 bg-solar-500/10',
    blue: 'text-blue-500 bg-blue-500/10',
    green: 'text-green-500 bg-green-500/10',
    emerald: 'text-emerald-500 bg-emerald-500/10',
    red: 'text-red-500 bg-red-500/10',
};

const bgMap: Record<string, string> = {
    solar: 'bg-solar-500',
    blue: 'bg-blue-500',
    green: 'bg-green-500',
    emerald: 'bg-emerald-500',
    red: 'bg-red-500',
};

export const MetricCard: React.FC<MetricCardProps> = ({
    title,
    value,
    unit,
    icon: Icon,
    trend,
    color = 'solar',
    loading = false
}) => {
    if (loading) {
        return (
            <Card className="h-[140px] relative overflow-hidden">
                <div className="flex justify-between items-start mb-4">
                    <div className="space-y-2">
                        <Skeleton className="h-4 w-24" />
                        <Skeleton className="h-8 w-32" />
                    </div>
                    <Skeleton className="h-12 w-12 rounded-xl" />
                </div>
                <Skeleton className="h-4 w-40 mt-4" />
            </Card>
        )
    }

    const iconColorClass = colorMap[color] || colorMap['solar'];
    const decorationColorClass = bgMap[color] || bgMap['solar'];

    return (
        <motion.div whileHover={{ scale: 1.02 }} className="h-full">
            <Card className="h-full relative overflow-hidden group">
                <div className="flex justify-between items-start">
                    <div>
                        <p className="text-slate-500 text-sm font-medium mb-1">{title}</p>
                        <div className="flex items-baseline gap-1">
                            <h3 className="text-2xl font-bold text-slate-900 tracking-tight">{value}</h3>
                            {unit && <span className="text-slate-400 text-sm">{unit}</span>}
                        </div>
                    </div>
                    <div className={cn("p-3 rounded-xl transition-colors", iconColorClass)}>
                        <Icon size={24} />
                    </div>
                </div>

                {/* Decoration */}
                <div className={cn(
                    "absolute -right-6 -bottom-6 w-24 h-24 rounded-full opacity-0 group-hover:opacity-20 transition-opacity blur-2xl",
                    decorationColorClass
                )} />

                {/* Trend Indicator */}
                {trend !== undefined && (
                    <div className={cn("mt-4 text-xs font-medium flex items-center gap-1", trend >= 0 ? "text-green-500" : "text-red-500")}>
                        <span>{trend > 0 ? '+' : ''}{trend}%</span>
                        <span className="text-slate-500">vs yesterday</span>
                    </div>
                )}
            </Card>
        </motion.div>
    );
};
