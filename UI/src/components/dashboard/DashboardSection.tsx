import React, { type ReactNode } from 'react';
import { motion } from 'framer-motion';
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
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: "easeOut" }}
        >
            <Card
                variant="gradient"
                className={cn(
                    "h-full flex flex-col hover:shadow-xl transition-all duration-500",
                    className
                )}
                noPadding
            >
                <div className="px-6 py-4 border-b border-slate-100/80 flex justify-between items-center bg-gradient-to-r from-slate-50/80 to-white">
                    <div className="flex items-center gap-3">
                        {Icon && (
                            <motion.div
                                initial={{ scale: 0.8, opacity: 0 }}
                                animate={{ scale: 1, opacity: 1 }}
                                transition={{ delay: 0.1, type: "spring", stiffness: 200 }}
                                className="p-2.5 bg-gradient-to-br from-solar-400 to-orange-500 rounded-xl shadow-md shadow-solar-500/20 text-white"
                            >
                                <Icon size={20} strokeWidth={2} />
                            </motion.div>
                        )}
                        <div>
                            <h3 className="font-semibold text-slate-800 text-lg">{title}</h3>
                        </div>
                    </div>
                    {action}
                </div>
                <div className="p-6 flex-1">
                    {children}
                </div>
            </Card>
        </motion.div>
    );
};
