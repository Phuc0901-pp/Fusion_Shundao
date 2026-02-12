import React, { useState } from 'react';
import { Zap, Pencil } from 'lucide-react';
import type { Site } from '../../types';
import { LoggerGroup } from './LoggerGroup';
import { RenameModal } from '../modals/RenameModal';

interface SiteGroupProps {
    site: Site;
    selectedLoggerId: string;
    selectedInverterId: string;
    selectedSiteId: string;
}

export const SiteGroup: React.FC<SiteGroupProps> = React.memo(({ site, selectedLoggerId, selectedInverterId, selectedSiteId }) => {
    const [showRename, setShowRename] = useState(false);
    const [displayName, setDisplayName] = useState(site.name);

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
                    <h3 className="text-lg font-bold text-slate-800">{displayName}</h3>
                    {site.dbId && (
                        <button
                            onClick={(e) => { e.stopPropagation(); setShowRename(true); }}
                            className="p-1 text-slate-300 hover:text-blue-500 hover:bg-blue-50 rounded-lg transition"
                            title="Đổi tên trạm"
                        >
                            <Pencil size={14} />
                        </button>
                    )}
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

            {/* Rename Modal */}
            {showRename && site.dbId && (
                <RenameModal
                    isOpen={showRename}
                    onClose={() => setShowRename(false)}
                    entityType="site"
                    entityId={site.dbId}
                    currentName={displayName}
                    defaultName={site.defaultName || site.id}
                    onRenamed={(n) => setDisplayName(n)}
                />
            )}
        </div>
    );
});
