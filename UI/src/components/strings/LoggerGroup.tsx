import React from 'react';
import { cn } from '../../utils/cn';
import type { SmartLogger } from '../../types';
import { InverterCard } from './InverterCard';

interface LoggerGroupProps {
    logger: SmartLogger;
    selectedInverterId: string;
    isSiteFiltered: boolean;
}

export const LoggerGroup: React.FC<LoggerGroupProps> = React.memo(({ logger, selectedInverterId, isSiteFiltered }) => {
    // Filter inverters for this logger based on selection
    const inverters = React.useMemo(() => {
        return selectedInverterId === "all"
            ? logger.inverters
            : logger.inverters.filter(i => i.id === selectedInverterId);
    }, [logger.inverters, selectedInverterId]);

    if (inverters.length === 0) return null;

    return (
        <div className={cn("space-y-4", !isSiteFiltered ? "pl-4 md:pl-8 border-l-2 border-slate-50" : "")}>
            {/* Logger Header */}
            <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-blue-500" />
                <h4 className="text-base font-semibold text-slate-700">{logger.name}</h4>
                <span className="text-xs text-slate-400">({inverters.length} Biến tần)</span>
            </div>

            {/* Inverters Loop */}
            <div className="space-y-8">
                {inverters.map((inverter) => (
                    <InverterCard key={inverter.id} inverter={inverter} />
                ))}
            </div>
        </div>
    );
});
