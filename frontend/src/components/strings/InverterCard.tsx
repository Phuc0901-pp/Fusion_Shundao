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
                className={`
                    relative group cursor-pointer overflow-hidden rounded-2xl transition-all duration-300
                    bg-white border border-slate-200 shadow-sm hover:shadow-xl hover:-translate-y-1
                    ${isConnected ? 'hover:border-green-200' : 'hover:border-red-200'}
                `}
            >
                {/* Status Stripe (Left Border) */}
                <div className={`absolute left-0 top-0 bottom-0 w-1 ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />

                <div className="p-4 pl-5">
                    {/* Header: Name and Edit */}
                    <div className="flex justify-between items-start mb-3">
                        <div>
                            <h5 className="font-bold text-slate-700 text-sm truncate pr-2 group-hover:text-blue-600 transition-colors">
                                {displayName.replace('HF1 ', '').replace('HF5 ', '')}
                            </h5>
                            <div className="flex items-center gap-1.5 mt-1">
                                <span className={`flex h-2 w-2 rounded-full ${isConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`} />
                                <span className={`text-[10px] font-medium uppercase tracking-wider ${isConnected ? 'text-green-600' : 'text-red-600'}`}>
                                    {isConnected ? 'Running' : 'Fault'}
                                </span>
                            </div>
                        </div>

                        {inverter.dbId && (
                            <button
                                onClick={(e) => { e.stopPropagation(); setShowRename(true); }}
                                className="p-1.5 text-slate-300 hover:text-blue-500 hover:bg-blue-50 rounded-lg transition-colors opacity-0 group-hover:opacity-100"
                            >
                                <Pencil size={14} />
                            </button>
                        )}
                    </div>

                    {/* Main Metrics Grid */}
                    <div className="grid grid-cols-2 gap-4 mb-4">
                        <div>
                            <p className="text-[8px] text-slate-400 font-medium uppercase mb-0.5">Công suất thuần</p>
                            <div className="flex items-baseline gap-1">
                                <span className={`text-[12px] font-bold font-mono tracking-tight ${isConnected ? 'text-slate-800' : 'text-slate-400'}`}>
                                    {inverter.pOutKw?.toFixed(2) || '0.00'}
                                </span>
                                <span className="text-xs text-slate-500 font-medium">kW</span>
                            </div>
                        </div>

                        <div>
                            <p className="text-[8px] text-slate-400 font-medium uppercase mb-0.5">Sản lượng ngày</p>
                            <div className="flex items-baseline gap-1">
                                <span className="text-[12px] font-bold font-mono tracking-tight text-slate-600">
                                    {inverter.eDailyKwh?.toFixed(1) || '0.0'}
                                </span>
                                <span className="text-xs text-slate-500 font-medium">kWh</span>
                            </div>
                        </div>
                    </div>

                    {/* Footer: String Status (Moved up slightly as no border needed now) */}
                    {setupCount > 0 && (
                        <div className="mt-3 flex items-center justify-between">
                            <div className="flex flex-col">
                                <span className="text-[9px] text-slate-400 font-medium">Chuỗi PV hoạt động</span>
                                <div className="flex items-center gap-1 mt-0.5">
                                    <span className={`text-xs font-bold ${warningCount > 0 ? "text-amber-500" : "text-green-600"}`}>
                                        {normalCount}
                                    </span>
                                    <span className="text-[10px] text-slate-300">/</span>
                                    <span className="text-xs font-medium text-slate-400">{setupCount}</span>
                                </div>
                            </div>

                            {/* Visual String Indicators (Dots) */}
                            <div className="flex gap-0.5 max-w-[80px] flex-wrap justify-end">
                                {Array.from({ length: Math.min(setupCount, 10) }).map((_, idx) => {
                                    const s = getString(idx + 1);
                                    const isOk = s.current > 0; // Simplified check
                                    return (
                                        <div
                                            key={idx}
                                            className={`w-1.5 h-1.5 rounded-full ${isOk ? 'bg-green-400' : 'bg-slate-200'}`}
                                        />
                                    );
                                })}
                                {setupCount > 10 && <span className="text-[8px] text-slate-300 leading-none self-end">+</span>}
                            </div>
                        </div>
                    )}
                </div>

                {/* Warning details overlay (optional, simplified from previous version) */}
                {warningCount > 0 && (
                    <div className="absolute top-2 right-8">
                        <span className="flex h-2 w-2 relative">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-amber-500"></span>
                        </span>
                    </div>
                )}
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
