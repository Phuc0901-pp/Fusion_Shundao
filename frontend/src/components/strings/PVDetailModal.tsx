import React from 'react';
import { X, Zap, Battery } from 'lucide-react';
import type { StringData } from '../../types';

interface PVDetailModalProps {
    isOpen: boolean;
    onClose: () => void;
    data: StringData | null;
}

export const PVDetailModal: React.FC<PVDetailModalProps> = ({ isOpen, onClose, data }) => {
    if (!isOpen || !data) return null;

    // Close on backdrop click
    const handleBackdropClick = (e: React.MouseEvent) => {
        if (e.target === e.currentTarget) {
            onClose();
        }
    };

    const power = data.current * data.voltage; // Watts
    const isActive = data.current > 0 && data.voltage > 0;

    return (
        <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm animate-fade-in"
            onClick={handleBackdropClick}
        >
            <div className="bg-white rounded-2xl shadow-2xl w-[90%] max-w-sm overflow-hidden animate-scale-in">
                {/* Header */}
                <div className="bg-slate-50 px-6 py-4 border-b border-slate-100 flex justify-between items-center">
                    <div className="flex items-center gap-2">
                        <Battery className={isActive ? "text-green-500" : "text-slate-400"} size={20} />
                        <h3 className="font-bold text-lg text-slate-800">{data.id}</h3>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-1 hover:bg-slate-200 rounded-full transition-colors text-slate-500"
                    >
                        <X size={20} />
                    </button>
                </div>

                {/* Body */}
                <div className="p-6 space-y-6">
                    {/* Status Badge */}
                    <div className="flex justify-center">
                        <span className={`px-4 py-1.5 rounded-full text-sm font-medium border ${isActive
                                ? "bg-green-50 text-green-700 border-green-200"
                                : data.voltage > 0
                                    ? "bg-red-50 text-red-700 border-red-200"
                                    : "bg-slate-100 text-slate-500 border-slate-200"
                            }`}>
                            {isActive ? "Đang hoạt động" : data.voltage > 0 ? "Không có dòng điện" : "Mất kết nối"}
                        </span>
                    </div>

                    {/* Metrics Grid */}
                    <div className="grid grid-cols-2 gap-4">
                        <div className="bg-blue-50 p-4 rounded-xl border border-blue-100 flex flex-col items-center">
                            <span className="text-slate-500 text-xs uppercase font-semibold mb-1">Điện Áp (V)</span>
                            <span className="text-2xl font-bold text-blue-700 tabular-nums">{data.voltage.toFixed(0)}</span>
                        </div>
                        <div className="bg-amber-50 p-4 rounded-xl border border-amber-100 flex flex-col items-center">
                            <span className="text-slate-500 text-xs uppercase font-semibold mb-1">Dòng Điện (A)</span>
                            <span className="text-2xl font-bold text-amber-700 tabular-nums">{data.current.toFixed(2)}</span>
                        </div>
                    </div>

                    {/* Power */}
                    <div className="bg-slate-900 text-white p-5 rounded-xl flex items-center justify-between">
                        <div className="flex items-center gap-2">
                            <Zap className="text-yellow-400" size={20} />
                            <span className="font-medium">Công Suất</span>
                        </div>
                        <span className="text-2xl font-bold tabular-nums">
                            {power.toFixed(0)} <span className="text-sm text-slate-400 font-normal">W</span>
                        </span>
                    </div>
                </div>
            </div>
        </div>
    );
};
