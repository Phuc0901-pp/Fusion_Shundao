import React, { useEffect, useRef } from 'react';
import { AlertTriangle, CheckCircle, Info, XCircle } from 'lucide-react';
import type { AlertMessage } from '../../types';
import { cn } from '../../utils/cn';
import { Card, CardHeader, CardTitle } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Skeleton } from '../ui/Skeleton';

interface AlertBoxProps {
    alerts: AlertMessage[];
    loading?: boolean;
}

const LEVEL_ICONS = {
    info: Info,
    warning: AlertTriangle,
    error: XCircle,
    success: CheckCircle,
};

const LEVEL_COLORS = {
    info: "text-blue-600 bg-blue-50 border-blue-200",
    warning: "text-amber-600 bg-amber-50 border-amber-200",
    error: "text-red-600 bg-red-50 border-red-200",
    success: "text-green-600 bg-green-50 border-green-200",
};

export const AlertBox: React.FC<AlertBoxProps> = ({ alerts, loading = false }) => {
    const scrollRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [alerts]);

    if (loading) {
        return (
            <Card className="h-[400px] flex flex-col p-0" noPadding>
                <div className="p-4 border-b border-slate-700">
                    <Skeleton className="h-6 w-40" />
                </div>
                <div className="p-4 space-y-3">
                    <Skeleton className="h-16 w-full" />
                    <Skeleton className="h-16 w-full" />
                    <Skeleton className="h-16 w-full" />
                </div>
            </Card>
        )
    }

    return (
        <Card className="h-[400px] flex flex-col p-0" noPadding>
            <div className="p-4 border-b border-slate-200 flex justify-between items-center bg-slate-50/50">
                <CardTitle>
                    <AlertTriangle size={18} className="text-amber-500" />
                    Nhật Ký & Cảnh Báo
                </CardTitle>
                <Badge variant="default">{alerts.length} sự kiện</Badge>
            </div>

            <div
                ref={scrollRef}
                className="flex-1 overflow-y-auto p-4 space-y-3 scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent"
            >
                {alerts.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full text-slate-500">
                        <CheckCircle size={32} className="mb-2 opacity-50" />
                        <p>Hệ thống hoạt động bình thường</p>
                    </div>
                ) : (
                    alerts.map((alert) => {
                        const Icon = LEVEL_ICONS[alert.level];
                        const date = new Date(alert.timestamp).toLocaleTimeString();
                        return (
                            <div key={alert.id} className="flex gap-3 items-start animate-fade-in group">
                                <span className="text-xs text-slate-500 mt-1 min-w-[60px] font-mono">{date}</span>
                                <div className={cn(
                                    "p-3 rounded-lg border flex-1 text-sm transition-all duration-200",
                                    "group-hover:translate-x-1",
                                    LEVEL_COLORS[alert.level]
                                )}>
                                    <div className="flex items-center gap-2 font-semibold mb-1">
                                        <Icon size={14} />
                                        <span className="capitalize">{alert.source}</span>
                                    </div>
                                    {alert.message}
                                </div>
                            </div>
                        )
                    })
                )}
            </div>
        </Card>
    );
};
