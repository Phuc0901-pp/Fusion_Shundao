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
    variant?: 'default' | 'flat';
    onClick?: () => void;
}

const colorMap: Record<string, string> = {
    solar: 'text-solar-500 bg-solar-500/10',
    blue: 'text-blue-500 bg-blue-500/10',
    green: 'text-green-500 bg-green-500/10',
    emerald: 'text-emerald-500 bg-emerald-500/10',
    red: 'text-red-500 bg-red-500/10',
    slate: 'text-slate-500 bg-slate-500/10',
    orange: 'text-orange-500 bg-orange-500/10',
};

const bgMap: Record<string, string> = {
    solar: 'bg-solar-500',
    blue: 'bg-blue-500',
    green: 'bg-green-500',
    emerald: 'bg-emerald-500',
    red: 'bg-red-500',
    slate: 'bg-slate-500',
    orange: 'bg-orange-500',
};

export const MetricCard: React.FC<MetricCardProps> = ({
    title,
    value,
    unit,
    icon: Icon,
    trend,
    color = 'solar',
    loading = false,
    variant = 'default',
    onClick
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

    const cardContent = (
        <div className="h-full relative overflow-hidden group">
            <div className="flex justify-between items-start z-10 relative">
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
                "absolute -right-6 -bottom-6 w-24 h-24 rounded-full opacity-0 group-hover:opacity-10 transition-opacity blur-2xl",
                decorationColorClass
            )} />

            {/* Trend Indicator */}
            {trend !== undefined && (
                <div className={cn("mt-4 text-xs font-medium flex items-center gap-1", trend >= 0 ? "text-green-500" : "text-red-500")}>
                    <span>{trend > 0 ? '+' : ''}{trend}%</span>
                    <span className="text-slate-500">vs yesterday</span>
                </div>
            )}
        </div>
    );

    if (variant === 'flat') {
        return (
            <motion.div whileHover={onClick ? { scale: 1.02 } : {}} className="h-full" onClick={onClick}>
                <div className={cn(
                    "h-full p-5 rounded-2xl border border-slate-100 bg-slate-50/50 hover:bg-white hover:border-slate-200 transition-all",
                    onClick ? "cursor-pointer active:scale-95 transition-transform" : "cursor-default"
                )}>
                    {cardContent}
                </div>
            </motion.div>
        );
    }

    return (
        <motion.div whileHover={onClick ? { scale: 1.02 } : {}} className="h-full" onClick={onClick}>
            <Card className={cn(
                "h-full relative overflow-hidden group p-6",
                onClick && "cursor-pointer"
            )}>
                {cardContent}
            </Card>
        </motion.div>
    );
};
