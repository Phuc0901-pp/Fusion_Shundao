import React, { useState } from 'react';
import type { StringData } from '../../types';
import { PVDetailModal } from './PVDetailModal';

interface PVStringProps {
    data: StringData;
}

export const PVString: React.FC<PVStringProps> = React.memo(({ data }) => {
    const [showModal, setShowModal] = useState(false);

    // Determine status color
    // Green: I > 0, U > 0
    // Red: I == 0, U > 0
    // Gray: I == 0, U == 0
    let statusClass = "bg-slate-300 border-slate-400"; // Gray (Default/Disconnected)

    if (data.current > 0 && data.voltage > 0) {
        statusClass = "bg-green-500 border-green-600 shadow-[0_0_8px_rgba(34,197,94,0.6)]"; // Green (On)
    } else if (data.voltage > 0) {
        statusClass = "bg-red-500 border-red-600 animate-pulse"; // Red (Fault/Review)
    }

    return (
        <>
            <div
                className="flex flex-col items-center gap-1 group cursor-pointer"
                onClick={() => setShowModal(true)}
                title={`String ${data.id}: ${data.voltage}V / ${data.current}A`}
            >
                {/* Status Light */}
                <div className={`w-6 h-6 rounded-full border-2 transition-transform transform group-hover:scale-125 ${statusClass}`} />

                {/* Label */}
                <span className="text-[10px] font-mono text-slate-400 group-hover:text-slate-700 transition-colors">
                    {data.id.split('-').pop()?.replace('PV', '')}
                </span>
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
    // Custom comparison for performance optimization
    // Only re-render if voltage or current changes significantly or status changes
    // But since we removed animations, simple ID/Current/Voltage check is fine.
    // React.memo with default shallow compare might be enough if props are stable, 
    // but explicit check is safer for heavy lists.
    return (
        prev.data.id === next.data.id &&
        prev.data.current === next.data.current &&
        prev.data.voltage === next.data.voltage
    );
});
