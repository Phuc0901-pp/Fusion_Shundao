import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { cn } from '../../utils/cn';
import { Skeleton } from '../ui/Skeleton';
import { TrendingUp, TrendingDown } from 'lucide-react';

interface MetricCardProps {
    title: string;
    value: string | number;
    unit?: string;
    icon: React.ElementType;
    trend?: number;
    color?: string;
    loading?: boolean;
    variant?: 'default' | 'flat' | 'glass';
    onClick?: () => void;
    delay?: number;
}

const colorMap: Record<string, { icon: string; glow: string }> = {
    solar: { icon: 'text-solar-500 bg-solar-500/10', glow: 'group-hover:shadow-solar-500/20' },
    blue: { icon: 'text-blue-500 bg-blue-500/10', glow: 'group-hover:shadow-blue-500/20' },
    green: { icon: 'text-green-500 bg-green-500/10', glow: 'group-hover:shadow-green-500/20' },
    emerald: { icon: 'text-emerald-500 bg-emerald-500/10', glow: 'group-hover:shadow-emerald-500/20' },
    red: { icon: 'text-red-500 bg-red-500/10', glow: 'group-hover:shadow-red-500/20' },
    slate: { icon: 'text-slate-500 bg-slate-500/10', glow: 'group-hover:shadow-slate-500/20' },
    orange: { icon: 'text-orange-500 bg-orange-500/10', glow: 'group-hover:shadow-orange-500/20' },
};

const useAnimatedCounter = (end: number, duration: number = 800) => {
    const [count, setCount] = useState(0);
    useEffect(() => {
        let startTime: number;
        let animationFrame: number;
        const animate = (timestamp: number) => {
            if (!startTime) startTime = timestamp;
            const progress = Math.min((timestamp - startTime) / duration, 1);
            const easeOutQuart = 1 - Math.pow(1 - progress, 4);
            setCount(Math.floor(easeOutQuart * end));
            if (progress < 1) animationFrame = requestAnimationFrame(animate);
        };
        animationFrame = requestAnimationFrame(animate);
        return () => cancelAnimationFrame(animationFrame);
    }, [end, duration]);
    return count;
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
    onClick,
    delay = 0
}) => {
    const numericValue = typeof value === 'number' ? value : parseFloat(String(value)) || 0;
    const animatedValue = useAnimatedCounter(Math.floor(numericValue), 800);

    if (loading) {
        return (
            <div className="h-[110px] bg-white rounded-xl border border-slate-200 p-3 flex flex-col">
                <div className="flex justify-between items-start mb-auto">
                    <Skeleton className="h-8 w-8 rounded-lg" />
                </div>
                <div className="space-y-1.5">
                    <Skeleton className="h-3 w-16" />
                    <Skeleton className="h-6 w-20" />
                </div>
            </div>
        )
    }

    const colorStyles = colorMap[color] || colorMap['solar'];
    const displayValue = typeof value === 'number'
        ? (Number.isInteger(value) ? animatedValue.toLocaleString() : value.toLocaleString())
        : value;

    return (
        <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: delay * 0.05, duration: 0.3 }}
            whileHover={onClick ? { scale: 1.03, y: -2 } : { y: -2 }}
            whileTap={onClick ? { scale: 0.98 } : {}}
            onClick={onClick}
            className={cn(
                "h-[110px] group transition-all duration-300",
                colorStyles.glow,
                onClick && "cursor-pointer"
            )}
        >
            <div className={cn(
                "h-full p-3 rounded-xl border flex flex-col transition-all duration-300",
                variant === 'flat'
                    ? "border-slate-100 bg-gradient-to-br from-white to-slate-50/80 hover:border-slate-200 hover:shadow-md"
                    : variant === 'glass'
                        ? "glass hover:shadow-lg"
                        : "bg-white border-slate-200 hover:shadow-md"
            )}>
                {/* Row 1: Icon - Fixed height */}
                <div className="flex justify-end mb-1">
                    <div className={cn("p-2 rounded-lg shrink-0", colorStyles.icon)}>
                        <Icon size={16} strokeWidth={2.5} />
                    </div>
                </div>

                {/* Row 2: Title - Fixed height with line clamp */}
                <p className="text-slate-500 text-[11px] font-medium leading-tight h-[26px] line-clamp-2 mb-auto">
                    {title}
                </p>

                {/* Row 3: Value + Unit - Fixed at bottom */}
                <div className="flex items-baseline gap-1 mt-1">
                    <span
                        className="text-xl font-bold text-slate-800 tracking-tight tabular-nums leading-none"
                        title={typeof value === 'number' ? value.toLocaleString('en-US') : String(value)}
                    >
                        {displayValue}
                    </span>
                    {unit && <span className="text-slate-400 text-xs font-medium">{unit}</span>}
                </div>

                {/* Trend (if provided) */}
                {trend !== undefined && (
                    <div className={cn(
                        "text-[9px] font-medium flex items-center gap-0.5 mt-1",
                        trend >= 0 ? "text-green-600" : "text-red-600"
                    )}>
                        {trend >= 0 ? <TrendingUp size={9} /> : <TrendingDown size={9} />}
                        <span>{trend > 0 ? '+' : ''}{trend}%</span>
                    </div>
                )}
            </div>
        </motion.div>
    );
};
