import React, { useMemo } from 'react';
import { PVString } from './PVString';
import type { Inverter } from '../../types';

interface InverterCardProps {
    inverter: Inverter;
}

export const InverterCard: React.FC<InverterCardProps> = React.memo(({ inverter }) => {
    // Calculate Totals per Inverter
    const { totalPower, avgVoltage } = useMemo(() => {
        const active = inverter.strings.filter(s => s.current > 0 && s.voltage > 0);
        const power = active.reduce((sum, s) => sum + (s.current * s.voltage), 0); // Watts
        const voltage = active.length > 0
            ? active.reduce((sum, s) => sum + s.voltage, 0) / active.length
            : 0;
        return { activeStrings: active, totalPower: power, avgVoltage: voltage };
    }, [inverter.strings]);


    return (
        <div className="space-y-3 pl-4 border-l border-slate-100">
            {/* Inverter Header */}
            <div className="flex items-center gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-solar-500" />
                <h5 className="text-sm font-medium text-slate-600">{inverter.name}</h5>
                <span className="text-xs text-slate-400 ml-auto bg-slate-50 px-2 py-0.5 rounded-full border border-slate-100">
                    {inverter.strings.length} Chuỗi
                </span>
            </div>

            <div className="flex flex-col xl:flex-row gap-4">
                {/* Summary Side Panel */}
                <div className="flex flex-row xl:flex-col gap-2 shrink-0 w-full xl:w-32">
                    <div className="bg-emerald-500 rounded-lg p-3 flex flex-col items-center justify-center text-white shadow-sm flex-1">
                        <span className="text-lg font-bold">{(avgVoltage / 1000).toFixed(2)}</span>
                        <span className="text-xs opacity-90">kV (Avg)</span>
                    </div>
                    <div className="bg-blue-500 rounded-lg p-3 flex flex-col items-center justify-center text-white shadow-sm flex-1">
                        <span className="text-lg font-bold">{(totalPower / 1000).toFixed(1)}</span>
                        <span className="text-xs opacity-90">kW</span>
                    </div>
                </div>

                {/* Strings Grid */}
                <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8 gap-4 flex-1">
                    {inverter.strings.map((str) => (
                        <PVString key={str.id} data={str} />
                    ))}
                </div>
            </div>
        </div>
    );
});
