import React from 'react';
import { createPortal } from 'react-dom';
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

    return createPortal(
        <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm animate-fade-in"
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

                {/* Inverter Details List */}
                <div className="p-6 bg-white shrink-0 border-b border-slate-100 overflow-y-auto max-h-[50vh]">
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-x-8 gap-y-4 text-sm">

                        {/* Column 1: Status & Power */}
                        <div className="space-y-3">
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Trạng thái bộ biến tần</span>
                                <span className={`font-medium ${inverter.deviceStatus === 'Grid connected' ? 'text-green-600' : 'text-slate-700'}`}>
                                    {inverter.deviceStatus || "Không xác định"}
                                </span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Công suất thuần</span>
                                <span className="font-medium text-slate-700">{inverter.pOutKw?.toFixed(3) || "0"} kW</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Hệ số công suất</span>
                                <span className="font-medium text-slate-700">{inverter.powerFactor?.toFixed(3) || "1.000"}</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Dòng điện lưới pha A</span>
                                <span className="font-medium text-slate-700">{inverter.gridIaA?.toFixed(3) || "0"} A</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Điện áp pha A</span>
                                <span className="font-medium text-slate-700">{inverter.gridVaV?.toFixed(1) || "0"} V</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Thời gian khởi động</span>
                                <span className="font-medium text-slate-700">{inverter.startupTime || "--"}</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Điện trở cách điện</span>
                                <span className="font-medium text-slate-700">{inverter.insulationResistanceMO?.toFixed(3) || "0"} MΩ</span>
                            </div>
                        </div>

                        {/* Column 2: Energy & Output */}
                        <div className="space-y-3">
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Năng lượng hàng ngày</span>
                                <span className="font-medium text-slate-700">{inverter.eDailyKwh?.toFixed(2) || "0"} kWh</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Công suất vô công</span>
                                <span className="font-medium text-slate-700">{inverter.qOutKvar?.toFixed(3) || "0"} kvar</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Tần số lưới điện</span>
                                <span className="font-medium text-slate-700">{inverter.gridFreqHz?.toFixed(2) || "50.00"} Hz</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Dòng điện lưới pha B</span>
                                <span className="font-medium text-slate-700">{inverter.gridIbA?.toFixed(3) || "0"} A</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Điện áp pha B</span>
                                <span className="font-medium text-slate-700">{inverter.gridVbV?.toFixed(1) || "0"} V</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Thời gian tắt</span>
                                <span className="font-medium text-slate-700">{inverter.shutdownTime || "--"}</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Hiệu suất DC/AC</span>
                                <span className="font-medium text-slate-700">{inverter.pOutKw && inverter.dcPowerKw ? ((inverter.pOutKw / inverter.dcPowerKw) * 100).toFixed(2) : 0}%</span>
                            </div>
                        </div>

                        {/* Column 3: Total & Info */}
                        <div className="space-y-3">
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Tổng sản lượng</span>
                                <span className="font-medium text-slate-700">{(inverter.eTotalKwh || 0).toLocaleString()} kWh</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Công suất định mức</span>
                                <span className="font-medium text-slate-700">{inverter.ratedPowerKw?.toFixed(3) || "0"} kW</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Chế độ đầu ra</span>
                                <span className="font-medium text-slate-700">{inverter.outputMode || "3 pha 4 dây"}</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Dòng điện lưới pha C</span>
                                <span className="font-medium text-slate-700">{inverter.gridIcA?.toFixed(3) || "0"} A</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Điện áp pha C</span>
                                <span className="font-medium text-slate-700">{inverter.gridVcV?.toFixed(1) || "0"} V</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Nhiệt độ bên trong</span>
                                <span className="font-medium text-slate-700">{inverter.internalTempDegC?.toFixed(1) || "0"} °C</span>
                            </div>
                            <div className="flex justify-between border-b border-slate-100 pb-1">
                                <span className="text-slate-500">Số chuỗi PV</span>
                                <span className="font-medium text-slate-700">{inverter.strings.length}</span>
                            </div>
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
        </div>,
        document.body
    );
};
