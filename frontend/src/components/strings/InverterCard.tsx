import React, { useState } from 'react';
import type { Inverter } from '../../types';
import { InverterDetailModal } from './InverterDetailModal';
import { RenameModal } from '../modals/RenameModal';
import { Zap, Pencil } from 'lucide-react';

interface InverterCardProps {
    inverter: Inverter;
    index?: number;
}

export const InverterCard: React.FC<InverterCardProps> = React.memo(({ inverter }) => {
    const [showModal, setShowModal] = useState(false);
    const [showRename, setShowRename] = useState(false);
    const [displayName, setDisplayName] = useState(inverter.name);
    const [stringSet, setStringSet] = useState(inverter.numberStringSet || '');

    // Determine Status
    const status = inverter.deviceStatus || "";
    const isConnected = status.toLowerCase() === 'grid connected';

    // Calculate Active Strings
    const setupCount = parseInt(stringSet) || 0; // Only use setup count, do not fallback

    // Helper to get string data by index (1-based)
    const getString = (idx: number) => {
        const id = `PV${idx.toString().padStart(2, '0')}`;
        return inverter.strings.find(s => s.id === id) || { current: 0, voltage: 0 };
    };

    let normalCount = 0;
    let warningCount = 0;
    const abnormalStrings: string[] = [];
    const [showAbnormalDetails, setShowAbnormalDetails] = useState(false);

    // Time-based logic
    const currentHour = new Date().getHours();
    const isWorkingHours = currentHour >= 6 && currentHour < 18;

    if (setupCount > 0) {
        // 1. Identify "Actual/Connected" strings (Voltage > 0)
        const connectedStrings: { index: number, current: number, id: string }[] = [];

        for (let i = 1; i <= setupCount; i++) {
            const s = getString(i);
            if (s.voltage > 10) { // Assume > 10V means connected
                connectedStrings.push({ index: i, current: s.current, id: `PV${i}` });
            }
        }

        // 2. During off-hours, skip threshold analysis
        if (!isWorkingHours) {
            normalCount = connectedStrings.length;
        } else if (connectedStrings.length > 0) {
            // 3. Calculate Average of Actual strings
            const totalCurrent = connectedStrings.reduce((a, b) => a + b.current, 0);
            const avgCurrent = totalCurrent / connectedStrings.length;

            // 4. Compare each Actual string against 80% of Avg
            const threshold = avgCurrent * 0.8;

            for (const item of connectedStrings) {
                if (item.current >= threshold) {
                    normalCount++;
                } else {
                    warningCount++;
                    abnormalStrings.push(item.id);
                }
            }
        }
    }

    return (
        <>
            <div
                onClick={() => setShowModal(true)}
                className="bg-white p-3 rounded-xl shadow-sm border border-slate-100 hover:shadow-md transition group relative cursor-pointer"
            >
                {/* Rename Button - top right */}
                {inverter.dbId && (
                    <button
                        onClick={(e) => { e.stopPropagation(); setShowRename(true); }}
                        className="absolute top-2 right-2 p-1.5 text-slate-400 hover:text-blue-500 hover:bg-blue-50 rounded-lg opacity-0 group-hover:opacity-100 transition"
                        title="Đổi tên / Cấu hình"
                    >
                        <Pencil size={14} />
                    </button>
                )}

                <div className="flex flex-col items-center justify-center py-2">
                    {/* Status Indicator */}
                    <div className={`w-12 h-12 rounded-full flex items-center justify-center mb-2 ${isConnected ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'}`}>
                        <Zap size={24} className={isConnected ? "animate-pulse" : ""} />
                    </div>

                    <div className="text-center w-full">
                        <h5 className={`font-bold text-sm transition-colors ${isConnected ? 'text-slate-700 group-hover:text-green-700' : 'text-red-700'}`}>
                            {displayName.replace('HF1 ', '').replace('HF5 ', '')}
                        </h5>

                        {/* String Count Display */}
                        {setupCount > 0 && (
                            <div className="flex flex-col items-center mt-1 relative">
                                <span className="text-[10px] text-gray-500 font-medium">
                                    Active: <span className={warningCount > 0 ? "text-orange-500 font-bold" : "text-green-600"}>{normalCount}</span>/{setupCount}
                                </span>
                                {warningCount > 0 && (
                                    <>
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                setShowAbnormalDetails(!showAbnormalDetails);
                                            }}
                                            className="text-[9px] text-red-500 font-bold animate-pulse hover:scale-105 transition cursor-pointer flex items-center gap-0.5 mt-0.5"
                                        >
                                            ⚠️ {warningCount} abnormal {showAbnormalDetails ? '▲' : '▼'}
                                        </button>

                                        {/* Dropdown for Abnormal Strings */}
                                        {showAbnormalDetails && (
                                            <div className="absolute top-full mt-1 z-10 bg-white border border-red-100 shadow-lg rounded-lg p-2 w-max min-w-[100px] animate-in fade-in zoom-in duration-200" onClick={(e) => e.stopPropagation()}>
                                                <p className="text-[10px] text-slate-500 font-medium mb-1 border-b border-slate-100 pb-1">Chi tiết sự cố:</p>
                                                <div className="grid grid-cols-3 gap-1">
                                                    {abnormalStrings.map(id => (
                                                        <span key={id} className="text-[9px] font-bold text-red-600 bg-red-50 px-1 py-0.5 rounded text-center block">
                                                            {id}
                                                        </span>
                                                    ))}
                                                </div>
                                            </div>
                                        )}
                                    </>
                                )}
                            </div>
                        )}
                    </div>

                    {/* Status Text overlay */}
                    <span className={`
                        absolute top-2 left-2 text-[9px] uppercase tracking-wider font-semibold px-1.5 py-0.5 rounded-full
                        ${isConnected ? 'text-green-700 bg-green-100' : 'text-red-700 bg-red-100'}
                    `}>
                        {isConnected ? 'Running' : 'Fault'}
                    </span>
                </div>
            </div>

            {/* Inverter Detail Modal */}
            {showModal && (
                <InverterDetailModal
                    isOpen={showModal}
                    onClose={() => setShowModal(false)}
                    inverter={inverter}
                />
            )}

            {/* Rename Modal */}
            {showRename && inverter.dbId && (
                <RenameModal
                    isOpen={showRename}
                    onClose={() => setShowRename(false)}
                    entityType="device"
                    entityId={inverter.dbId}
                    currentName={displayName}
                    defaultName={inverter.defaultName || inverter.id}
                    currentStringSet={stringSet}
                    onRenamed={(n, s) => {
                        setDisplayName(n);
                        if (s !== undefined) setStringSet(s);
                    }}
                />
            )}
        </>
    );
});
