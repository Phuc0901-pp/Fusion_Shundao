import React, { useState } from 'react';

import type { Inverter } from '../../types';
import { InverterDetailModal } from './InverterDetailModal';
import { Zap, AlertTriangle } from 'lucide-react';

interface InverterCardProps {
    inverter: Inverter;
    index?: number;
}

export const InverterCard: React.FC<InverterCardProps> = React.memo(({ inverter }) => {
    const [showModal, setShowModal] = useState(false);

    // Determine Status
    const status = inverter.deviceStatus || "";
    const isConnected = status.toLowerCase() === 'grid connected';

    return (
        <>
            <div
                onClick={() => setShowModal(true)}
                className={`
                    cursor-pointer relative overflow-hidden rounded-xl border p-4 flex flex-col items-center justify-center gap-2 transition-all duration-200 ease-out transform hover:scale-105 hover:-translate-y-1 active:scale-95 will-change-transform
                    ${isConnected
                        ? 'bg-gradient-to-br from-white to-green-50/50 border-green-200/60 hover:border-green-400 hover:shadow-lg hover:shadow-green-500/10 group'
                        : 'bg-gradient-to-br from-red-50 to-red-100/50 border-red-200 hover:border-red-400 hover:shadow-lg hover:shadow-red-500/10'
                    }
                `}
            >
                {/* Status Indicator */}
                <div className={`
                    w-4 h-4 rounded-full border-2 flex items-center justify-center
                    ${isConnected
                        ? 'bg-gradient-to-br from-green-400 to-green-600 border-green-600 shadow-[0_0_12px_rgba(34,197,94,0.5)]'
                        : 'bg-gradient-to-br from-red-400 to-red-600 border-red-600 animate-pulse'
                    }
                `}>
                    {isConnected
                        ? <Zap size={8} className="text-white" />
                        : <AlertTriangle size={8} className="text-white" />
                    }
                </div>

                <h5 className={`font-bold text-sm text-center transition-colors ${isConnected ? 'text-slate-700 group-hover:text-green-700' : 'text-red-700'}`}>
                    {inverter.name.replace('HF1 ', '').replace('HF5 ', '')}
                </h5>

                {/* Status Text */}
                <span className={`
                    text-[10px] uppercase tracking-wider font-semibold px-2 py-0.5 rounded-full
                    ${isConnected
                        ? 'text-green-700 bg-green-100'
                        : 'text-red-700 bg-red-100'
                    }
                `}>
                    {isConnected ? 'Running' : 'Fault'}
                </span>

                {/* Decorative Glow */}
                <div className={`
                    absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none
                    ${isConnected
                        ? 'bg-gradient-to-br from-green-400/5 to-transparent'
                        : 'bg-gradient-to-br from-red-400/10 to-transparent'
                    }
                `} />
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
