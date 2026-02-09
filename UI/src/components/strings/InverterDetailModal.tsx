import React from 'react';
import { X, Battery } from 'lucide-react';
import type { Inverter } from '../../types';
import { PVString } from './PVString';

interface InverterDetailModalProps {
    isOpen: boolean;
    onClose: () => void;
    inverter: Inverter;
}

export const InverterDetailModal: React.FC<InverterDetailModalProps> = ({ isOpen, onClose, inverter }) => {
    if (!isOpen) return null;

    // Close on backdrop click
    const handleBackdropClick = (e: React.MouseEvent) => {
        if (e.target === e.currentTarget) {
            onClose();
        }
    };

    // Derived Metrics
    const activeStrings = inverter.strings.filter(s => s.current > 0 && s.voltage > 0);
    const totalPower = activeStrings.reduce((sum, s) => sum + (s.current * s.voltage), 0);
    const avgVoltage = activeStrings.length > 0
        ? activeStrings.reduce((sum, s) => sum + s.voltage, 0) / activeStrings.length
        : 0;

    return (
        <div
            className="fixed inset-0 z-40 flex items-center justify-center bg-black/50 backdrop-blur-sm animate-fade-in"
            onClick={handleBackdropClick}
        >
            <div className="bg-white rounded-2xl shadow-2xl w-[95%] max-w-4xl h-[85vh] flex flex-col overflow-hidden animate-scale-in">
                {/* Header */}
                <div className="bg-slate-50 px-6 py-4 border-b border-slate-100 flex justify-between items-center shrink-0">
                    <div className="flex items-center gap-3">
                        <div className={`w-3 h-3 rounded-full ${inverter.deviceStatus === 'Grid connected' ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)]' : 'bg-red-500'}`} />
                        <div>
                            <h3 className="font-bold text-xl text-slate-800">{inverter.name}</h3>
                            <p className="text-sm text-slate-500">Trạng thái: <span className="font-medium">{inverter.deviceStatus || "Không xác định"}</span></p>
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-2 hover:bg-slate-200 rounded-full transition-colors text-slate-500"
                    >
                        <X size={24} />
                    </button>
                </div>

                {/* Dashboard Stats */}
                <div className="grid grid-cols-3 gap-4 p-6 bg-white shrink-0 border-b border-slate-100">
                    <div className="bg-emerald-50 p-4 rounded-xl border border-emerald-100 flex flex-col items-center">
                        <span className="text-slate-500 text-xs uppercase font-semibold mb-1">Điện Áp TB</span>
                        <div className="flex items-baseline gap-1">
                            <span className="text-2xl font-bold text-emerald-700 tabular-nums">{(avgVoltage / 1000).toFixed(2)}</span>
                            <span className="text-xs text-emerald-600 font-medium">kV</span>
                        </div>
                    </div>
                    <div className="bg-blue-50 p-4 rounded-xl border border-blue-100 flex flex-col items-center">
                        <span className="text-slate-500 text-xs uppercase font-semibold mb-1">Tổng Công Suất</span>
                        <div className="flex items-baseline gap-1">
                            <span className="text-2xl font-bold text-blue-700 tabular-nums">{(totalPower / 1000).toFixed(1)}</span>
                            <span className="text-xs text-blue-600 font-medium">kW</span>
                        </div>
                    </div>
                    <div className="bg-slate-50 p-4 rounded-xl border border-slate-100 flex flex-col items-center">
                        <span className="text-slate-500 text-xs uppercase font-semibold mb-1">Số Chuỗi</span>
                        <div className="flex items-baseline gap-1">
                            <span className="text-2xl font-bold text-slate-700 tabular-nums">{inverter.strings.length}</span>
                            <span className="text-xs text-slate-500 font-medium">Str</span>
                        </div>
                    </div>
                </div>

                {/* Strings Grid (Scrollable) */}
                <div className="flex-1 overflow-y-auto p-6 bg-slate-50/50">
                    <h4 className="flex items-center gap-2 font-semibold text-slate-700 mb-4">
                        <Battery size={18} /> Chi Tiết Chuỗi PV
                    </h4>
                    <div className="grid grid-cols-4 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-10 xl:grid-cols-12 gap-3 pb-8">
                        {inverter.strings.map((str) => (
                            <PVString key={str.id} data={str} />
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};
