import React from 'react';
import { Zap } from 'lucide-react';
import type { Site } from '../../types';
import { LoggerGroup } from './LoggerGroup';

interface SiteGroupProps {
    site: Site;
    selectedLoggerId: string;
    selectedInverterId: string;
    selectedSiteId: string;
}

export const SiteGroup: React.FC<SiteGroupProps> = React.memo(({ site, selectedLoggerId, selectedInverterId, selectedSiteId }) => {
    // Filter loggers for this site based on selection
    const loggers = React.useMemo(() => {
        return selectedLoggerId === "all"
            ? site.loggers
            : site.loggers.filter(l => l.id === selectedLoggerId);
    }, [site.loggers, selectedLoggerId]);

    if (loggers.length === 0) return null;

    return (
        <div className="space-y-6">
            {/* Site Header (only show if multiple sites are displayed) */}
            {selectedSiteId === "all" && (
                <div className="flex items-center gap-3 pb-2 border-b-2 border-slate-100">
                    <div className="p-2 bg-slate-100 rounded-lg text-slate-500">
                        <Zap size={20} />
                    </div>
                    <h3 className="text-lg font-bold text-slate-800">{site.name}</h3>
                    <span className="text-xs font-medium px-2 py-0.5 bg-slate-100 text-slate-500 rounded-full">
                        Mã trạm: {site.id}
                    </span>
                </div>
            )}

            {/* Loggers Loop */}
            {loggers.map(logger => (
                <LoggerGroup
                    key={logger.id}
                    logger={logger}
                    selectedInverterId={selectedInverterId}
                    isSiteFiltered={selectedSiteId !== "all"}
                />
            ))}
        </div>
    );
});
