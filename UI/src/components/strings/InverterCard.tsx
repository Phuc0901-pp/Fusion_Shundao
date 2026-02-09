import React, { useState } from 'react';
import type { Inverter } from '../../types';
import { InverterDetailModal } from './InverterDetailModal';

interface InverterCardProps {
    inverter: Inverter;
}

export const InverterCard: React.FC<InverterCardProps> = React.memo(({ inverter }) => {
    const [showModal, setShowModal] = useState(false);

    // Determine Status
    // "Grid connected" -> Green
    // Others -> Red
    const status = inverter.deviceStatus || "";
    const isConnected = status.toLowerCase() === 'grid connected';

    return (
        <>
            <div
                onClick={() => setShowModal(true)}
                className={`
                    cursor-pointer relative overflow-hidden rounded-xl border p-4 flex flex-col items-center justify-center gap-2 transition-all duration-300 hover:shadow-lg hover:scale-[1.02]
                    ${isConnected
                        ? 'bg-white border-green-200 hover:border-green-400 group'
                        : 'bg-red-50 border-red-200 hover:border-red-400'
                    }
                `}
            >
                {/* Status Indicator (Pulse) */}
                <div className={`
                    w-4 h-4 rounded-full border-2 
                    ${isConnected
                        ? 'bg-green-500 border-green-600 shadow-[0_0_12px_rgba(34,197,94,0.6)]'
                        : 'bg-red-500 border-red-600 animate-pulse'
                    }
                `} />

                <h5 className={`font-bold text-sm text-center ${isConnected ? 'text-slate-700' : 'text-red-700'}`}>
                    {inverter.name.replace('HF1 ', '').replace('HF5 ', '')}
                </h5>

                {/* Status Text Label (Optional, matches color) */}
                <span className={`text-[10px] uppercase tracking-wider font-semibold ${isConnected ? 'text-green-600' : 'text-red-500'}`}>
                    {isConnected ? 'Running' : 'Fault'}
                </span>
            </div>

            {/* Inverter Detail Modal */}
            {showModal && (
                <InverterDetailModal
                    isOpen={showModal}
                    onClose={() => setShowModal(false)}
                    inverter={inverter}
                />
            )}
        </>
    );
});
