import React from 'react';
import { Battery } from 'lucide-react';
import { Card, CardTitle } from '../ui/Card';
import { Skeleton } from '../ui/Skeleton';
import type { Site } from '../../types';
import { SiteGroup } from './SiteGroup';

interface StringDiagramProps {
    sites?: Site[];
    loading?: boolean;
}

export const StringDiagram: React.FC<StringDiagramProps> = ({ sites = [], loading = false }) => {
    // Selection state - Default to 'all' for "Show Full"
    const [selectedSiteId, setSelectedSiteId] = React.useState<string>("all");
    const [selectedLoggerId, setSelectedLoggerId] = React.useState<string>("all");
    const [selectedInverterId, setSelectedInverterId] = React.useState<string>("all");

    // Reset children filters when parent changes
    React.useEffect(() => {
        // Reset to "all" whenever site changes to simplify UX
        setSelectedLoggerId("all");
        setSelectedInverterId("all");
    }, [selectedSiteId]);

    React.useEffect(() => {
        if (selectedLoggerId !== "all") {
            setSelectedInverterId("all");
        }
    }, [selectedLoggerId]);

    // Derived Data

    // 1. Available Loggers based on Site selection (for Dropdown)
    const availableLoggers = React.useMemo(() => {
        if (selectedSiteId === "all") {
            return sites.flatMap(s => s.loggers);
        }
        return sites.find(s => s.id === selectedSiteId)?.loggers || [];
    }, [sites, selectedSiteId]);

    // 2. Filter Loggers based on Logger selection (for Display)
    const displayedLoggers = React.useMemo(() => {
        if (selectedLoggerId === "all") return availableLoggers;
        return availableLoggers.filter(l => l.id === selectedLoggerId);
    }, [availableLoggers, selectedLoggerId]);

    // 3. Filter Sites for Display Loop
    const displayedSites = React.useMemo(() => {
        if (selectedSiteId === "all") return sites;
        return sites.filter(s => s.id === selectedSiteId);
    }, [sites, selectedSiteId]);

    if (loading) {
        return (
            <Card>
                <div className="flex justify-between items-center mb-6">
                    <div className="space-y-2">
                        <Skeleton className="h-6 w-32" />
                        <Skeleton className="h-4 w-48" />
                    </div>
                </div>
                <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8 gap-4">
                    {Array.from({ length: 16 }).map((_, i) => (
                        <Skeleton key={i} className="h-24 w-full rounded-xl" />
                    ))}
                </div>
            </Card>
        )
    }

    return (
        <Card className="flex flex-col gap-6">
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                <div>
                    <CardTitle className="mb-1">
                        <Battery size={20} className="text-solar-500" />
                        Trạng Thái Chuỗi PV
                    </CardTitle>
                    {/* Selectors Row */}
                    <div className="flex flex-wrap gap-2 mt-2">
                        <select
                            className="bg-slate-50 border border-slate-200 text-slate-700 text-xs rounded-lg p-1.5 outline-none focus:border-solar-500"
                            value={selectedSiteId}
                            onChange={(e) => setSelectedSiteId(e.target.value)}
                        >
                            <option value="all">Tất cả dự án</option>
                            {sites.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
                        </select>
                        <span className="text-slate-300">/</span>
                        <select
                            className="bg-slate-50 border border-slate-200 text-slate-700 text-xs rounded-lg p-1.5 outline-none focus:border-solar-500"
                            value={selectedLoggerId}
                            onChange={(e) => setSelectedLoggerId(e.target.value)}
                        >
                            <option value="all">Tất cả bộ ghi</option>
                            {availableLoggers.map(l => <option key={l.id} value={l.id}>{l.name}</option>)}
                        </select>
                        <span className="text-slate-300">/</span>
                        <select
                            className="bg-slate-50 border border-slate-200 text-slate-700 text-xs rounded-lg p-1.5 outline-none focus:border-solar-500"
                            value={selectedInverterId}
                            onChange={(e) => setSelectedInverterId(e.target.value)}
                        >
                            <option value="all">Tất cả biến tần</option>
                            {displayedLoggers.flatMap(l => l.inverters).map(inv => (
                                <option key={inv.id} value={inv.id}>{inv.name}</option>
                            ))}
                        </select>
                    </div>
                </div>
                <div className="flex gap-4 text-xs">
                    <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)]" />
                        <span className="text-slate-600">Hoạt động</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-slate-400" />
                        <span className="text-slate-400">Mất kết nối</span>
                    </div>
                </div>
            </div>

            <div className="space-y-12">
                {displayedSites.map(site => (
                    <SiteGroup
                        key={site.id}
                        site={site}
                        selectedLoggerId={selectedLoggerId}
                        selectedInverterId={selectedInverterId}
                        selectedSiteId={selectedSiteId}
                    />
                ))}
            </div>
        </Card>
    );
};
