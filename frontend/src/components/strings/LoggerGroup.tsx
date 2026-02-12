import React, { useState } from 'react';
import { Pencil } from 'lucide-react';
import { cn } from '../../utils/cn';
import type { SmartLogger } from '../../types';
import { InverterCard } from './InverterCard';
import { RenameModal } from '../modals/RenameModal';

interface LoggerGroupProps {
    logger: SmartLogger;
    selectedInverterId: string;
    isSiteFiltered: boolean;
}

export const LoggerGroup: React.FC<LoggerGroupProps> = React.memo(({ logger, selectedInverterId, isSiteFiltered }) => {
    const [showRename, setShowRename] = useState(false);
    const [displayName, setDisplayName] = useState(logger.name);

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
                <h4 className="text-base font-semibold text-slate-700">{displayName}</h4>
                {logger.dbId && (
                    <button
                        onClick={(e) => { e.stopPropagation(); setShowRename(true); }}
                        className="p-1 text-slate-300 hover:text-blue-500 hover:bg-blue-50 rounded-lg transition"
                        title="Đổi tên logger"
                    >
                        <Pencil size={12} />
                    </button>
                )}
                <span className="text-xs text-slate-400">({inverters.length} Biến tần)</span>
            </div>

            {/* Inverters Loop */}
            <div className="grid grid-cols-2 md:grid-cols-5 xl:grid-cols-10 gap-4">
                {inverters.map((inverter) => (
                    <InverterCard key={inverter.id} inverter={inverter} />
                ))}
            </div>

            {/* Rename Modal */}
            {showRename && logger.dbId && (
                <RenameModal
                    isOpen={showRename}
                    onClose={() => setShowRename(false)}
                    entityType="logger"
                    entityId={logger.dbId}
                    currentName={displayName}
                    defaultName={logger.defaultName || logger.id}
                    onRenamed={(n) => setDisplayName(n)}
                />
            )}
        </div>
    );
});
