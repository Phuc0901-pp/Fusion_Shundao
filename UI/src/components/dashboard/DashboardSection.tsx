import React, { type ReactNode } from 'react';
import type { LucideIcon } from 'lucide-react';
import { Card } from '../ui/Card';
import { cn } from '../../utils/cn';

interface DashboardSectionProps {
    title: string;
    icon?: LucideIcon;
    children: ReactNode;
    className?: string;
    action?: ReactNode;
}

export const DashboardSection: React.FC<DashboardSectionProps> = ({
    title,
    icon: Icon,
    children,
    className,
    action
}) => {
    return (
        <Card className={cn("h-full flex flex-col hover:shadow-md transition-shadow duration-300", className)} noPadding>
            <div className="px-6 py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
                <div className="flex items-center gap-3">
                    {Icon && (
                        <div className="p-2 bg-white rounded-lg border border-slate-200 shadow-sm text-solar-500">
                            <Icon size={20} />
                        </div>
                    )}
                    <h3 className="font-semibold text-slate-800 text-lg">{title}</h3>
                </div>
                {action}
            </div>
            <div className="p-6 flex-1">
                {children}
            </div>
        </Card>
    );
};
