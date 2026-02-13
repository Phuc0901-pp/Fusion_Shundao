import React, { useCallback, useMemo, lazy, Suspense } from 'react';
import { createPortal } from 'react-dom';
import { X, Battery, AlertTriangle, ShieldCheck } from 'lucide-react';
import type { Inverter, StringData } from '../../types';
import { PVString } from './PVString';
import type { PVStringStatus } from './PVString';

// Lazy load the heavy chart component — only loads when modal opens
const InverterPowerChart = lazy(() =>
    import('./InverterPowerChart').then(m => ({ default: m.InverterPowerChart }))
);

interface InverterDetailModalProps {
    isOpen: boolean;
    onClose: () => void;
    inverter: Inverter;
}

// ─── Smart Alert Logic ─────────────────────────────────────────────
const WORKING_HOUR_START = 6;  // 06:00
const WORKING_HOUR_END = 18;   // 18:00
const THRESHOLD_PERCENT = 0.8; // 80%

interface StringAnalysis {
    status: PVStringStatus;
    data: StringData;
}

function analyzeStrings(strings: StringData[]): StringAnalysis[] {
    const currentHour = new Date().getHours();
    const isWorkingHours = currentHour >= WORKING_HOUR_START && currentHour < WORKING_HOUR_END;

    // Step 1: Calculate averages from active strings (V > 10 and I > 0)
    const activeStrings = strings.filter(s => s.voltage > 10 && s.current > 0);
    const avgVoltage = activeStrings.length > 0
        ? activeStrings.reduce((sum, s) => sum + s.voltage, 0) / activeStrings.length
        : 0;
    const avgCurrent = activeStrings.length > 0
        ? activeStrings.reduce((sum, s) => sum + s.current, 0) / activeStrings.length
        : 0;

    const thresholdVoltage = avgVoltage * THRESHOLD_PERCENT;
    const thresholdCurrent = avgCurrent * THRESHOLD_PERCENT;

    // Step 2: Analyze each string
    return strings.map(s => {
        const hasVoltage = s.voltage > 10;
        const hasCurrent = s.current > 0;

        // Case 1: Outside working hours
        if (!isWorkingHours) {
            if (!hasVoltage && !hasCurrent) {
                return {
                    data: s,
                    status: {
                        state: 'inactive' as const,
                        message: 'Inverter không hoạt động (ngoài giờ làm việc)',
                    },
                };
            }
            // If there IS data outside working hours, it's still "normal"
            return {
                data: s,
                status: {
                    state: 'normal' as const,
                    message: 'Hoạt động bình thường',
                },
            };
        }

        // Case 2: Working hours — Zero output = Critical fault
        if (!hasVoltage && !hasCurrent) {
            return {
                data: s,
                status: {
                    state: 'error' as const,
                    message: 'Inverter có sự cố (mất dòng & mất điện áp)',
                },
            };
        }
        if (!hasCurrent && hasVoltage) {
            return {
                data: s,
                status: {
                    state: 'error' as const,
                    message: 'Inverter có sự cố (mất dòng)',
                },
            };
        }
        if (!hasVoltage && hasCurrent) {
            return {
                data: s,
                status: {
                    state: 'error' as const,
                    message: 'Inverter có sự cố (mất điện áp)',
                },
            };
        }

        // Case 3: Working hours — Below threshold = Warning
        const warnings: string[] = [];
        if (avgCurrent > 0 && s.current < thresholdCurrent) {
            warnings.push(`${s.current.toFixed(2)}A < ${avgCurrent.toFixed(2)}A`);
        }
        if (avgVoltage > 0 && s.voltage < thresholdVoltage) {
            warnings.push(`${s.voltage.toFixed(1)}V < ${avgVoltage.toFixed(1)}V`);
        }

        if (warnings.length > 0) {
            return {
                data: s,
                status: {
                    state: 'warning' as const,
                    message: 'Inverter gặp sự cố (vượt ngưỡng)',
                    detail: warnings.join(' | '),
                },
            };
        }

        // Case 4: Normal
        return {
            data: s,
            status: {
                state: 'normal' as const,
                message: 'Hoạt động bình thường',
            },
        };
    });
}

// Skeleton loader for the chart area
const ChartSkeleton = () => (
    <div className="bg-white border border-slate-200 rounded-xl overflow-hidden animate-pulse">
        <div className="px-4 py-3 border-b border-slate-100">
            <div className="h-4 bg-slate-200 rounded w-48" />
        </div>
        <div className="p-4" style={{ height: 240 }}>
            <div className="h-full bg-slate-100 rounded-lg" />
        </div>
    </div>
);

// Memoized detail row to avoid re-creating on every render
const DetailRow = React.memo(({ label, value }: { label: string; value: string }) => (
    <div className="flex justify-between border-b border-slate-100 pb-1 items-center">
        <span className="text-slate-500">{label}</span>
        <span className="font-medium text-slate-700">{value}</span>
    </div>
));

export const InverterDetailModal: React.FC<InverterDetailModalProps> = React.memo(({ isOpen, onClose, inverter }) => {
    if (!isOpen) return null;

    // Memoize backdrop click handler
    const handleBackdropClick = useCallback((e: React.MouseEvent) => {
        if (e.target === e.currentTarget) onClose();
    }, [onClose]);

    // ─── Smart Alert Analysis ──────────────────────────────────────
    const stringAnalysis = useMemo(() => analyzeStrings(inverter.strings), [inverter.strings]);

    const faultCount = stringAnalysis.filter(a => a.status.state === 'error').length;
    const warningCount = stringAnalysis.filter(a => a.status.state === 'warning').length;
    const inactiveCount = stringAnalysis.filter(a => a.status.state === 'inactive').length;
    const hasIssues = faultCount > 0 || warningCount > 0;

    // Memoize the details rows data to avoid recomputing
    const detailRows = useMemo(() => [
        // Row 1
        { label: 'Trạng thái', value: inverter.deviceStatus || 'Không xác định', isStatus: true },
        { label: 'Năng lượng ngày', value: `${inverter.eDailyKwh?.toFixed(2) || '0'} kWh` },
        { label: 'Tổng sản lượng', value: `${(inverter.eTotalKwh || 0).toLocaleString()} kWh` },
        // Row 2
        { label: 'Công suất thuần', value: `${inverter.pOutKw?.toFixed(3) || '0'} kW` },
        { label: 'Công suất vô công', value: `${inverter.qOutKvar?.toFixed(3) || '0'} kvar` },
        { label: 'Công suất định mức', value: `${inverter.ratedPowerKw?.toFixed(3) || '0'} kW` },
        // Row 3
        { label: 'Hệ số công suất', value: inverter.powerFactor?.toFixed(3) || '0' },
        { label: 'Tần số lưới điện', value: `${inverter.gridFreqHz?.toFixed(2) || '0'} Hz` },
        { label: 'Nhiệt độ bên trong', value: `${inverter.internalTempDegC?.toFixed(1) || '0'} °C` },
        // Row 4
        { label: 'Dòng điện pha A', value: `${inverter.gridIaA?.toFixed(3) || '0'} A` },
        { label: 'Dòng điện pha B', value: `${inverter.gridIbA?.toFixed(3) || '0'} A` },
        { label: 'Dòng điện pha C', value: `${inverter.gridIcA?.toFixed(3) || '0'} A` },
        // Row 5
        { label: 'Điện áp pha A', value: `${inverter.gridVaV?.toFixed(1) || '0'} V` },
        { label: 'Điện áp pha B', value: `${inverter.gridVbV?.toFixed(1) || '0'} V` },
        { label: 'Điện áp pha C', value: `${inverter.gridVcV?.toFixed(1) || '0'} V` },
        // Row 6
        { label: 'TG khởi động', value: inverter.startupTime || '--' },
        { label: 'TG tắt', value: inverter.shutdownTime || '--' },
        { label: 'Điện trở cách điện', value: `${inverter.insulationResistanceMO?.toFixed(3) || '0'} MΩ` },
    ], [inverter]);

    return createPortal(
        <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 animate-fade-in"
            onClick={handleBackdropClick}
        >
            <div className="bg-white rounded-2xl shadow-2xl w-[90%] 2xl:w-[90%] max-w-[95vw] h-[95vh] flex flex-col overflow-hidden animate-scale-in will-change-transform">
                {/* Header */}
                <div className="bg-slate-50 px-6 py-4 border-b border-slate-100 flex justify-between items-center shrink-0">
                    <div className="flex items-center gap-3">
                        <div className={`w-3 h-3 rounded-full ${inverter.deviceStatus === 'Grid connected' ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)]' : 'bg-red-500'}`} />
                        <div>
                            <h3 className="font-bold text-xl text-slate-800">{inverter.name}</h3>
                            <p className="text-sm text-slate-500">Trạng thái: <span className="font-medium">{inverter.deviceStatus || "Gặp sự cố"}</span></p>
                        </div>
                    </div>

                    <div className="flex items-center gap-4">
                        <div className="px-3 py-1 bg-slate-100 text-slate-600 rounded-lg text-sm font-medium border border-slate-200">
                            Số chuỗi PV: <span className="text-slate-900">{inverter.strings.length}</span>
                        </div>
                        <button
                            onClick={onClose}
                            className="p-2 hover:bg-slate-200 rounded-full transition-colors text-slate-500 cursor-pointer"
                        >
                            <X size={24} />
                        </button>
                    </div>
                </div>

                {/* Body: Horizontal Split */}
                <div className="flex-1 flex overflow-hidden min-h-0">
                    {/* Left Panel: Details + Chart */}
                    <div className="flex-[9] overflow-y-auto border-r border-slate-100">
                        {/* Inverter Details — data-driven rendering */}
                        <div className="p-6 bg-white border-b border-slate-100">
                            <div className="grid grid-cols-3 gap-x-12 gap-y-4 text-sm">
                                {detailRows.map((row, idx) => (
                                    row.isStatus ? (
                                        <div key={idx} className="flex justify-between border-b border-slate-100 pb-1 items-center">
                                            <span className="text-slate-500">{row.label}</span>
                                            <span className={`font-medium ${inverter.deviceStatus === 'Grid connected' ? 'text-green-600' : 'text-slate-700'}`}>
                                                {row.value}
                                            </span>
                                        </div>
                                    ) : (
                                        <DetailRow key={idx} label={row.label} value={row.value} />
                                    )
                                ))}
                            </div>
                        </div>

                        {/* Power Chart — Lazy Loaded with Skeleton */}
                        <div className="p-2">
                            <Suspense fallback={<ChartSkeleton />}>
                                <InverterPowerChart inverterId={inverter.id} inverterName={inverter.name} />
                            </Suspense>
                        </div>
                    </div>

                    {/* Right Panel: PV Strings */}
                    <div className="flex-[1] overflow-y-auto bg-slate-50/50 p-5">
                        <h4 className="flex items-center gap-2 font-semibold text-slate-700 mb-4 sticky top-0 bg-slate-50/90 backdrop-blur-sm py-2 -mt-2 z-10">
                            <Battery size={18} /> Chuỗi PV
                        </h4>

                        {/* ─── Alert Summary Banner ─────────────────────── */}
                        {hasIssues && (
                            <div className="mb-3 p-2.5 rounded-lg border border-red-200 bg-red-50/80 text-xs">
                                <div className="flex items-center gap-1.5 font-semibold text-red-700 mb-1">
                                    <AlertTriangle size={14} />
                                    Phát hiện sự cố chuỗi PV
                                </div>
                                <div className="text-red-600 space-y-0.5">
                                    {faultCount > 0 && (
                                        <p>🔴 <strong>{faultCount}</strong> chuỗi mất dòng/mất áp</p>
                                    )}
                                    {warningCount > 0 && (
                                        <p>🟠 <strong>{warningCount}</strong> chuỗi vượt ngưỡng (&lt;80% trung bình)</p>
                                    )}
                                </div>
                            </div>
                        )}

                        {inactiveCount === inverter.strings.length && inactiveCount > 0 && (
                            <div className="mb-3 p-2.5 rounded-lg border border-slate-200 bg-slate-50 text-xs">
                                <div className="flex items-center gap-1.5 font-medium text-slate-500">
                                    <ShieldCheck size={14} />
                                    Inverter không hoạt động (ngoài giờ làm việc)
                                </div>
                            </div>
                        )}

                        <div className="grid grid-cols-3 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-2 2xl:grid-cols-2 gap-2 pb-2">
                            {stringAnalysis.map((analysis) => (
                                <PVString key={analysis.data.id} data={analysis.data} status={analysis.status} />
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </div>,
        document.body
    );
});
