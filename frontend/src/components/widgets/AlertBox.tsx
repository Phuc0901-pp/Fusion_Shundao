import React, { useEffect, useRef, useState, useCallback, memo } from 'react';
import { AlertTriangle, CheckCircle, Info, XCircle, Filter, ExternalLink } from 'lucide-react';
import type { AlertMessage, Site, Inverter } from '../../types';
import { cn } from '../../utils/cn';
import { Card } from '../ui/Card';
import { Skeleton } from '../ui/Skeleton';
import { InverterDetailModal } from '../strings/InverterDetailModal';

// Extended alert type with device reference
interface DeviceAlert extends AlertMessage {
    deviceId?: string;
    deviceType?: 'inverter' | 'sensor' | 'meter' | 'system';
}

interface AlertBoxProps {
    alerts: DeviceAlert[];
    sites?: Site[];
    loading?: boolean;
}

const LEVEL_ICONS = {
    info: Info,
    warning: AlertTriangle,
    error: XCircle,
    success: CheckCircle,
};

const LEVEL_COLORS = {
    info: "text-blue-600 bg-blue-50/80 border-blue-200",
    warning: "text-amber-600 bg-amber-50/80 border-amber-200",
    error: "text-red-600 bg-red-50/80 border-red-200",
    success: "text-green-600 bg-green-50/80 border-green-200",
};

type FilterLevel = 'all' | 'error' | 'warning' | 'info' | 'success';

// Memoized Alert Item for performance
const AlertItem = memo<{
    alert: DeviceAlert;
    onClick?: () => void;
    isClickable: boolean;
}>(({ alert, onClick, isClickable }) => {
    const Icon = LEVEL_ICONS[alert.level];
    const time = new Date(alert.timestamp).toLocaleTimeString('vi-VN', {
        hour: '2-digit',
        minute: '2-digit'
    });

    return (
        <div
            onClick={onClick}
            className={cn(
                "p-2.5 rounded-lg border text-sm transition-colors duration-150",
                isClickable && "cursor-pointer hover:brightness-95",
                LEVEL_COLORS[alert.level]
            )}
        >
            <div className="flex items-start gap-2">
                <Icon size={14} className="mt-0.5 shrink-0" />
                <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2">
                        <span className="font-medium text-xs truncate flex items-center gap-1">
                            {alert.source}
                            {isClickable && <ExternalLink size={9} className="opacity-40" />}
                        </span>
                        <span className="text-[10px] opacity-50 shrink-0 tabular-nums">{time}</span>
                    </div>
                    <p className="text-[11px] opacity-70 leading-snug mt-0.5 line-clamp-1">{alert.message}</p>
                </div>
            </div>
        </div>
    );
});

AlertItem.displayName = 'AlertItem';

// Helper to find inverter by ID
const findInverterById = (sites: Site[], deviceId: string): Inverter | null => {
    for (const site of sites) {
        for (const logger of site.loggers || []) {
            for (const inverter of logger.inverters || []) {
                if (inverter.id === deviceId) return inverter;
            }
        }
    }
    return null;
};

// Alert sound using Web Audio API (more reliable than Audio element)
const playAlertSound = (level: 'error' | 'warning') => {
    try {
        const audioContext = new (window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext)();
        const oscillator = audioContext.createOscillator();
        const gainNode = audioContext.createGain();

        oscillator.connect(gainNode);
        gainNode.connect(audioContext.destination);

        // Different tones for error vs warning
        oscillator.frequency.value = level === 'error' ? 800 : 600;
        oscillator.type = 'sine';

        gainNode.gain.setValueAtTime(0.3, audioContext.currentTime);
        gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.5);

        oscillator.start(audioContext.currentTime);
        oscillator.stop(audioContext.currentTime + 0.5);

        // Play 2 beeps for error
        if (level === 'error') {
            setTimeout(() => {
                const osc2 = audioContext.createOscillator();
                const gain2 = audioContext.createGain();
                osc2.connect(gain2);
                gain2.connect(audioContext.destination);
                osc2.frequency.value = 800;
                osc2.type = 'sine';
                gain2.gain.setValueAtTime(0.3, audioContext.currentTime);
                gain2.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.3);
                osc2.start(audioContext.currentTime);
                osc2.stop(audioContext.currentTime + 0.3);
            }, 200);
        }
    } catch {
        console.log('Audio not supported');
    }
};

export const AlertBox: React.FC<AlertBoxProps> = memo(({ alerts, sites = [], loading = false }) => {
    const scrollRef = useRef<HTMLDivElement>(null);
    const [filter, setFilter] = useState<FilterLevel>('all');
    const [showFilter, setShowFilter] = useState(false);
    const [selectedInverter, setSelectedInverter] = useState<Inverter | null>(null);
    const prevAlertCountRef = useRef({ error: 0, warning: 0 });
    const isFirstLoad = useRef(true);

    // Play sound when new error/warning alerts arrive (or on initial load)
    useEffect(() => {
        const currentErrors = alerts.filter(a => a.level === 'error').length;
        const currentWarnings = alerts.filter(a => a.level === 'warning').length;

        // On first load with existing alerts, play sound
        if (isFirstLoad.current && alerts.length > 0) {
            isFirstLoad.current = false;
            if (currentErrors > 0) {
                playAlertSound('error');
            } else if (currentWarnings > 0) {
                playAlertSound('warning');
            }
            prevAlertCountRef.current = { error: currentErrors, warning: currentWarnings };
            return;
        }

        isFirstLoad.current = false;

        // Check for new errors
        if (currentErrors > prevAlertCountRef.current.error) {
            playAlertSound('error');
        }
        // Check for new warnings (only if no new errors)
        else if (currentWarnings > prevAlertCountRef.current.warning) {
            playAlertSound('warning');
        }

        prevAlertCountRef.current = { error: currentErrors, warning: currentWarnings };
    }, [alerts]);

    // Auto-scroll to bottom on new alerts
    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [alerts.length]);

    const handleAlertClick = useCallback((alert: DeviceAlert) => {
        if (alert.deviceType === 'inverter' && alert.deviceId) {
            const inverter = findInverterById(sites, alert.deviceId);
            if (inverter) setSelectedInverter(inverter);
        }
    }, [sites]);

    const closeModal = useCallback(() => setSelectedInverter(null), []);

    // Filter alerts
    const filteredAlerts = filter === 'all' ? alerts : alerts.filter(a => a.level === filter);
    const errorCount = alerts.filter(a => a.level === 'error').length;
    const warningCount = alerts.filter(a => a.level === 'warning').length;

    if (loading) {
        return (
            <Card className="h-[540px] flex flex-col" noPadding>
                <div className="p-3 border-b border-slate-200">
                    <Skeleton className="h-5 w-32" />
                </div>
                <div className="p-3 space-y-2 flex-1">
                    {[...Array(5)].map((_, i) => <Skeleton key={i} className="h-12 w-full rounded-lg" />)}
                </div>
            </Card>
        );
    }

    return (
        <>
            <Card className="h-[540px] flex flex-col" noPadding>
                {/* Header - Simplified */}
                <div className="px-3 py-2.5 border-b border-slate-200 flex justify-between items-center bg-slate-50/80">
                    <div className="flex items-center gap-2">
                        <div className="p-1.5 bg-amber-500 rounded-md text-white">
                            <AlertTriangle size={14} />
                        </div>
                        <span className="font-semibold text-slate-700 text-sm">Nhật Ký Cảnh Báo</span>
                    </div>
                    <div className="flex items-center gap-1.5">
                        {errorCount > 0 && (
                            <span className="px-1.5 py-0.5 text-[10px] font-medium bg-red-500 text-white rounded">
                                {errorCount}
                            </span>
                        )}
                        {warningCount > 0 && (
                            <span className="px-1.5 py-0.5 text-[10px] font-medium bg-amber-500 text-white rounded">
                                {warningCount}
                            </span>
                        )}

                        {/* Filter */}
                        <div className="relative ml-1">
                            <button
                                onClick={() => setShowFilter(!showFilter)}
                                className="p-1 hover:bg-slate-200 rounded transition-colors text-slate-500"
                            >
                                <Filter size={12} />
                            </button>
                            {showFilter && (
                                <>
                                    <div className="fixed inset-0 z-10" onClick={() => setShowFilter(false)} />
                                    <div className="absolute right-0 top-full mt-1 bg-white border border-slate-200 rounded-lg shadow-lg z-20 py-1 min-w-[100px]">
                                        {(['all', 'error', 'warning', 'success'] as FilterLevel[]).map(level => (
                                            <button
                                                key={level}
                                                onClick={() => { setFilter(level); setShowFilter(false); }}
                                                className={cn(
                                                    "w-full px-2.5 py-1 text-left text-xs hover:bg-slate-50",
                                                    filter === level && "bg-slate-100 font-medium"
                                                )}
                                            >
                                                {level === 'all' ? 'Tất cả' : level === 'error' ? 'Lỗi' : level === 'warning' ? 'Cảnh báo' : 'OK'}
                                            </button>
                                        ))}
                                    </div>
                                </>
                            )}
                        </div>
                    </div>
                </div>

                {/* Alert List - Optimized */}
                <div ref={scrollRef} className="flex-1 overflow-y-auto p-2 space-y-1.5 scrollbar-thin">
                    {filteredAlerts.length === 0 ? (
                        <div className="flex flex-col items-center justify-center h-full text-slate-400">
                            <CheckCircle size={32} className="mb-2 text-green-400" />
                            <p className="text-sm font-medium">Hệ thống bình thường</p>
                        </div>
                    ) : (
                        filteredAlerts.map(alert => (
                            <AlertItem
                                key={alert.id}
                                alert={alert}
                                isClickable={alert.deviceType === 'inverter' && !!alert.deviceId}
                                onClick={() => handleAlertClick(alert)}
                            />
                        ))
                    )}
                </div>

                {/* Footer */}
                <div className="px-3 py-1.5 border-t border-slate-100 bg-slate-50/50 text-[10px] text-slate-400 flex justify-between">
                    <span>{filteredAlerts.length} sự kiện</span>
                    <span>{new Date().toLocaleTimeString('vi-VN')}</span>
                </div>
            </Card>

            {/* Modal */}
            {selectedInverter && (
                <InverterDetailModal
                    isOpen={true}
                    onClose={closeModal}
                    inverter={selectedInverter}
                />
            )}
        </>
    );
});

AlertBox.displayName = 'AlertBox';
