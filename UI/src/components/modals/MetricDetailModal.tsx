import React, { useRef, useEffect } from 'react';
import { X } from 'lucide-react';
import { createPortal } from 'react-dom';
import type { LucideIcon } from 'lucide-react';
import { Card } from '../ui/Card';

interface SiteKPI {
    id: string;
    name: string;
    value: number;
}

interface MetricDetailModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: string;
    unit: string;
    icon: LucideIcon;
    color: string;
    totalValue: number;
    sites: SiteKPI[];
}

export const MetricDetailModal: React.FC<MetricDetailModalProps> = ({
    isOpen,
    onClose,
    title,
    unit,
    icon: Icon,
    color,
    totalValue,
    sites,
}) => {
    const modalRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleEscape = (e: KeyboardEvent) => {
            if (e.key === 'Escape') onClose();
        };

        if (isOpen) {
            document.addEventListener('keydown', handleEscape);
            document.body.style.overflow = 'hidden';
        }

        return () => {
            document.removeEventListener('keydown', handleEscape);
            document.body.style.overflow = 'unset';
        };
    }, [isOpen, onClose]);

    if (!isOpen) return null;

    // Color classes map
    const colorMap: Record<string, string> = {
        solar: 'text-yellow-500 bg-yellow-50',
        blue: 'text-blue-500 bg-blue-50',
        green: 'text-green-500 bg-green-50',
        slate: 'text-slate-500 bg-slate-50',
        orange: 'text-orange-500 bg-orange-50',
        emerald: 'text-emerald-500 bg-emerald-50',
    };

    const colorClass = colorMap[color] || colorMap.solar;

    return createPortal(
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm animate-fade-in">
            <div ref={modalRef} className="w-full max-w-md animate-scale-in" onClick={(e) => e.stopPropagation()}>
                <Card className="overflow-hidden shadow-xl" noPadding={true}>
                    {/* Header */}
                    <div className="p-5 border-b border-slate-100 flex justify-between items-center bg-white">
                        <div className="flex items-center gap-3">
                            <div className={`p-2 rounded-lg ${colorClass}`}>
                                <Icon size={20} />
                            </div>
                            <h3 className="font-semibold text-slate-800 text-lg">{title}</h3>
                        </div>
                        <button
                            onClick={onClose}
                            className="p-1 rounded-full hover:bg-slate-100 text-slate-400 hover:text-slate-600 transition-colors"
                        >
                            <X size={20} />
                        </button>
                    </div>

                    {/* Content */}
                    <div className="p-5 bg-slate-50/50">
                        {/* Total Highlight */}
                        <div className="bg-white p-4 rounded-xl border border-slate-100 shadow-sm mb-4 flex justify-between items-center">
                            <span className="text-slate-500 font-medium">Tổng cộng</span>
                            <div className="text-right">
                                <div className="text-xl font-bold text-slate-900">
                                    {totalValue.toLocaleString('en-US', { maximumFractionDigits: 3 })}
                                </div>
                                <div className="text-xs text-slate-400 font-medium uppercase">{unit}</div>
                            </div>
                        </div>

                        {/* Sites List */}
                        <div className="space-y-3">
                            {sites.length > 0 ? (
                                sites.map((site) => (
                                    <div key={site.id} className="bg-white p-3 rounded-lg border border-slate-100 flex justify-between items-center hover:border-slate-200 transition-colors">
                                        <div className="flex items-center gap-3">
                                            <div className="w-2 h-2 rounded-full bg-slate-300"></div>
                                            <span className="text-slate-700 font-medium">{site.name}</span>
                                        </div>
                                        <div className="flex items-center gap-2">
                                            <span className="font-semibold text-slate-800">
                                                {site.value.toLocaleString('en-US', { maximumFractionDigits: 3 })}
                                            </span>
                                            <span className="text-xs text-slate-400 w-8 text-right">{unit}</span>
                                        </div>
                                    </div>
                                ))
                            ) : (
                                <div className="text-center py-6 text-slate-400">
                                    Không có dữ liệu chi tiết
                                </div>
                            )}
                        </div>
                    </div>
                </Card>
            </div>
        </div>,
        document.body
    );
};
