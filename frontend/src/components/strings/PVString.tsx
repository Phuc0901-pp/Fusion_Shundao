import React, { useState } from 'react';
import { AlertTriangle } from 'lucide-react';
import type { StringData } from '../../types';
import { PVDetailModal } from './PVDetailModal';

// Status object structure for smart alerts
export interface PVStringStatus {
    state: 'normal' | 'warning' | 'error' | 'inactive';
    message: string;
    detail?: string; // e.g. "0.4A < 2.0A (80%)"
}

interface PVStringProps {
    data: StringData;
    status?: PVStringStatus;
}

export const PVString: React.FC<PVStringProps> = React.memo(({ data, status }) => {
    const [showModal, setShowModal] = useState(false);

    // Determine visual style based on status prop (priority) or fallback to legacy logic
    let containerClass = "bg-slate-50 border-slate-200 text-slate-400"; // Gray (default)
    let valueClass = "text-slate-500";
    let showWarningIcon = false;

    if (status) {
        switch (status.state) {
            case 'error':
                containerClass = "bg-red-50 border-red-300 text-red-700 hover:border-red-400 hover:bg-red-100 ring-1 ring-red-200";
                valueClass = "text-red-700";
                showWarningIcon = true;
                break;
            case 'warning':
                containerClass = "bg-amber-50 border-amber-300 text-amber-700 hover:border-amber-400 hover:bg-amber-100 ring-1 ring-amber-200";
                valueClass = "text-amber-700";
                showWarningIcon = true;
                break;
            case 'inactive':
                containerClass = "bg-slate-50 border-slate-200 text-slate-400 opacity-60";
                valueClass = "text-slate-400";
                break;
            case 'normal':
                containerClass = "bg-green-50/50 border-green-200 text-green-700 hover:border-green-300 hover:bg-green-50";
                valueClass = "text-green-700";
                break;
        }
    } else {
        // Legacy fallback
        if (data.current > 0 && data.voltage > 10) {
            containerClass = "bg-green-50/50 border-green-200 text-green-700 hover:border-green-300 hover:bg-green-50";
            valueClass = "text-green-700";
        } else if (data.voltage > 10) {
            containerClass = "bg-red-50/50 border-red-200 text-red-600 hover:border-red-300 hover:bg-red-50";
            valueClass = "text-red-700";
        }
    }

    const tooltipText = status
        ? `${status.message}${status.detail ? `\n${status.detail}` : ''}`
        : 'Click for details';

    return (
        <>
            <div
                className={`relative flex flex-col items-center justify-center p-1.5 rounded-lg border transition-all duration-200 cursor-pointer hover:shadow-sm ${containerClass}`}
                onClick={() => setShowModal(true)}
                title={tooltipText}
            >
                {/* Warning Icon */}
                {showWarningIcon && (
                    <div className="absolute -top-1.5 -right-1.5 z-10">
                        <AlertTriangle size={12} className={status?.state === 'error' ? 'text-red-500' : 'text-amber-500'} fill="currentColor" />
                    </div>
                )}

                {/* ID */}
                <span className="text-[8px] font-bold uppercase tracking-wider opacity-60 mb-0.5">
                    {data.id.replace('PV', '')}
                </span>

                {/* Values */}
                <div className="flex flex-col items-center gap-0">
                    <div className={`text-[8px] font-semibold leading-none opacity-80 ${valueClass}`}>
                        {data.voltage.toFixed(1)} <span className="text-[8px] font-bold opacity-70">V</span>
                    </div>
                    <div className={`text-[8px] font-semibold leading-none opacity-80 ${valueClass}`}>
                        {data.current.toFixed(2)} <span className="text-[8px] font-bold opacity-70">A</span>
                    </div>
                </div>

                {/* Short status detail underneath */}
                {status?.detail && (
                    <div className={`text-[6px] font-medium mt-0.5 leading-tight text-center ${status.state === 'error' ? 'text-red-500' : 'text-amber-500'}`}>
                        {status.detail}
                    </div>
                )}
            </div>

            {/* Detail Modal */}
            {showModal && (
                <PVDetailModal
                    isOpen={showModal}
                    onClose={() => setShowModal(false)}
                    data={data}
                />
            )}
        </>
    );
}, (prev, next) => {
    return (
        prev.data.id === next.data.id &&
        prev.data.current === next.data.current &&
        prev.data.voltage === next.data.voltage &&
        prev.status?.state === next.status?.state &&
        prev.status?.detail === next.status?.detail
    );
});
