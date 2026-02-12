import React, { useEffect, useState, useRef, useCallback, useMemo } from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { Activity, Loader2, ZoomIn, Eye, EyeOff } from 'lucide-react';
import api from '../../services/api';
import {
    CHART_REFRESH_INTERVAL,
    CHART_MIN_VISIBLE_POINTS,
    CHART_X_AXIS_TICK_COUNT,
    CHART_ANIMATION_DURATION,
    COLORS,
} from '../../config/constants';

interface PowerPoint {
    time: string;
    dcPower?: number;
    acPower?: number;
}

interface InverterPowerChartProps {
    inverterId: string;
    inverterName: string;
}

export const InverterPowerChart: React.FC<InverterPowerChartProps> = ({ inverterId, inverterName }) => {
    const [data, setData] = useState<PowerPoint[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const isFirstLoad = useRef(true);

    // Zoom state
    const [zoomStart, setZoomStart] = useState(0);
    const [zoomEnd, setZoomEnd] = useState(0);
    const chartContainerRef = useRef<HTMLDivElement>(null);

    // Line visibility toggles
    const [showDc, setShowDc] = useState(true);
    const [showAc, setShowAc] = useState(true);

    useEffect(() => {
        let cancelled = false;
        const fetchData = async () => {
            // Only show loading spinner on first load
            if (isFirstLoad.current) setLoading(true);
            setError(null);
            try {
                const response: any = await api.get(`/inverter/dc-power?device=${encodeURIComponent(inverterId)}`);
                if (!cancelled) {
                    const newData = response.data || [];
                    setData(prev => {
                        // On first load or if zoomed out, update zoom range
                        if (isFirstLoad.current || zoomEnd === prev.length) {
                            setZoomStart(0);
                            setZoomEnd(newData.length);
                        }
                        return newData;
                    });
                    isFirstLoad.current = false;
                }
            } catch (err) {
                if (!cancelled) {
                    setError('Không thể tải dữ liệu biểu đồ');
                    console.error('Error fetching power data:', err);
                }
            } finally {
                if (!cancelled) setLoading(false);
            }
        };
        fetchData();
        const interval = setInterval(fetchData, CHART_REFRESH_INTERVAL);
        return () => { cancelled = true; clearInterval(interval); };
    }, [inverterId]);

    useEffect(() => {
        if (data.length > 0) {
            setZoomStart(0);
            setZoomEnd(data.length);
        }
    }, [data.length]);

    // Get latest values from data
    const latestValues = useMemo(() => {
        if (data.length === 0) return { dc: null, ac: null, time: '' };
        // Find last point that has data
        for (let i = data.length - 1; i >= 0; i--) {
            const point = data[i];
            if (point.dcPower != null || point.acPower != null) {
                return {
                    dc: point.dcPower ?? null,
                    ac: point.acPower ?? null,
                    time: point.time,
                };
            }
        }
        return { dc: null, ac: null, time: '' };
    }, [data]);

    // Handle Ctrl + Scroll zoom
    const handleWheel = useCallback((e: WheelEvent) => {
        if (!e.ctrlKey || data.length === 0) return;
        e.preventDefault();
        e.stopPropagation();

        const container = chartContainerRef.current;
        if (!container) return;

        const rect = container.getBoundingClientRect();
        const chartLeft = 45;
        const chartRight = 15;
        const chartWidth = rect.width - chartLeft - chartRight;
        const mouseX = e.clientX - rect.left - chartLeft;
        const ratio = Math.max(0, Math.min(1, mouseX / chartWidth));

        const currentRange = zoomEnd - zoomStart;
        const zoomFactor = e.deltaY > 0 ? 1.15 : 0.85;
        let newRange = Math.round(currentRange * zoomFactor);
        newRange = Math.max(CHART_MIN_VISIBLE_POINTS, Math.min(data.length, newRange));

        const centerIndex = zoomStart + ratio * currentRange;
        let newStart = Math.round(centerIndex - ratio * newRange);
        let newEnd = newStart + newRange;

        if (newStart < 0) { newStart = 0; newEnd = newRange; }
        if (newEnd > data.length) { newEnd = data.length; newStart = Math.max(0, newEnd - newRange); }

        setZoomStart(newStart);
        setZoomEnd(newEnd);
    }, [data.length, zoomStart, zoomEnd]);

    useEffect(() => {
        const container = chartContainerRef.current;
        if (!container) return;
        container.addEventListener('wheel', handleWheel, { passive: false });
        return () => container.removeEventListener('wheel', handleWheel);
    }, [handleWheel]);

    const visibleData = useMemo(() => {
        if (data.length === 0) return [];
        return data.slice(zoomStart, zoomEnd);
    }, [data, zoomStart, zoomEnd]);

    const isZoomed = data.length > 0 && (zoomEnd - zoomStart) < data.length;

    const resetZoom = () => {
        setZoomStart(0);
        setZoomEnd(data.length);
    };

    // Custom tooltip
    const CustomTooltip = ({ active, payload, label }: any) => {
        if (active && payload && payload.length) {
            return (
                <div className="bg-white px-3 py-2 rounded-lg shadow-lg border border-slate-200 text-xs">
                    <p className="text-slate-500 font-medium mb-1">{label}</p>
                    {payload.map((entry: any, idx: number) => (
                        <p key={idx} style={{ color: entry.color }} className="font-bold">
                            {entry.name}: {entry.value != null ? entry.value.toFixed(2) : '—'} kW
                        </p>
                    ))}
                </div>
            );
        }
        return null;
    };

    return (
        <div className="bg-white border border-slate-200 rounded-xl overflow-hidden">
            {/* Chart Header */}
            <div className="px-4 py-3 border-b border-slate-100">
                {/* Row 1: Title + latest values */}
                <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                        <Activity size={16} className="text-amber-500" />
                        <h4 className="font-semibold text-sm text-slate-700">
                            Đồ thị công suất — {inverterName}
                        </h4>
                    </div>
                    {/* Latest values */}
                    {data.length > 0 && (
                        <div className="flex items-center gap-3">
                            <div className="flex items-center gap-1.5 text-xs">
                                <div className="w-2 h-2 rounded-full bg-amber-500" />
                                <span className="text-slate-500">Tổng công suất đầu vào:</span>
                                <span className="font-bold text-amber-600">
                                    {latestValues.dc != null ? latestValues.dc.toFixed(2) : '—'} kW
                                </span>
                            </div>
                            <div className="flex items-center gap-1.5 text-xs">
                                <div className="w-2 h-2 rounded-full bg-blue-500" />
                                <span className="text-slate-500">Công suất thuần:</span>
                                <span className="font-bold text-blue-600">
                                    {latestValues.ac != null ? latestValues.ac.toFixed(2) : '—'} kW
                                </span>
                            </div>
                            <span className="text-[9px] text-slate-400">({latestValues.time})</span>
                        </div>
                    )}
                </div>

                {/* Row 2: Toggles + zoom controls */}
                <div className="flex items-center justify-between">
                    {/* Line toggles */}
                    <div className="flex items-center gap-2">
                        <button
                            onClick={() => setShowDc(!showDc)}
                            className={`flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium transition-all border cursor-pointer ${showDc
                                ? 'bg-amber-50 border-amber-200 text-amber-700'
                                : 'bg-slate-50 border-slate-200 text-slate-400 line-through'
                                }`}
                        >
                            {showDc ? <Eye size={10} /> : <EyeOff size={10} />}
                            Tổng công suất đầu vào
                        </button>
                        <button
                            onClick={() => setShowAc(!showAc)}
                            className={`flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium transition-all border cursor-pointer ${showAc
                                ? 'bg-blue-50 border-blue-200 text-blue-700'
                                : 'bg-slate-50 border-slate-200 text-slate-400 line-through'
                                }`}
                        >
                            {showAc ? <Eye size={10} /> : <EyeOff size={10} />}
                            Công suất thuần
                        </button>
                    </div>

                    {/* Zoom controls */}
                    <div className="flex items-center gap-3">
                        {isZoomed && (
                            <button
                                onClick={resetZoom}
                                className="text-[10px] text-blue-500 hover:text-blue-700 font-medium transition-colors cursor-pointer"
                            >
                                ↩ Thu nhỏ
                            </button>
                        )}
                        <span className="text-[10px] text-slate-400 font-medium flex items-center gap-1">
                            <ZoomIn size={10} />
                            Ctrl + Scroll để zoom
                        </span>
                    </div>
                </div>
            </div>

            {/* Chart Body */}
            <div ref={chartContainerRef} className="px-2 py-3" style={{ height: 240 }}>
                {loading ? (
                    <div className="flex items-center justify-center h-full text-slate-400 gap-2">
                        <Loader2 size={18} className="animate-spin" />
                        <span className="text-sm">Đang tải...</span>
                    </div>
                ) : error ? (
                    <div className="flex items-center justify-center h-full text-red-400 text-sm">
                        {error}
                    </div>
                ) : data.length === 0 ? (
                    <div className="flex items-center justify-center h-full text-slate-400 text-sm">
                        Chưa có dữ liệu hôm nay
                    </div>
                ) : (
                    <ResponsiveContainer width="100%" height="100%">
                        <AreaChart data={visibleData} margin={{ top: 5, right: 15, left: 0, bottom: 5 }}>
                            <defs>
                                <linearGradient id={`dcGrad-${inverterId}`} x1="0" y1="0" x2="0" y2="1">
                                    <stop offset="5%" stopColor={COLORS.dcPower.stroke} stopOpacity={0.25} />
                                    <stop offset="95%" stopColor={COLORS.dcPower.stroke} stopOpacity={0.02} />
                                </linearGradient>
                                <linearGradient id={`acGrad-${inverterId}`} x1="0" y1="0" x2="0" y2="1">
                                    <stop offset="5%" stopColor={COLORS.acPower.stroke} stopOpacity={0.25} />
                                    <stop offset="95%" stopColor={COLORS.acPower.stroke} stopOpacity={0.02} />
                                </linearGradient>
                            </defs>
                            <CartesianGrid strokeDasharray="3 3" stroke="#f1f5f9" />
                            <XAxis
                                dataKey="time"
                                tick={{ fontSize: 10, fill: '#94a3b8' }}
                                tickLine={false}
                                axisLine={{ stroke: '#e2e8f0' }}
                                interval={Math.max(0, Math.floor(visibleData.length / CHART_X_AXIS_TICK_COUNT) - 1)}
                            />
                            <YAxis
                                tick={{ fontSize: 10, fill: '#94a3b8' }}
                                tickLine={false}
                                axisLine={false}
                                width={45}
                                tickFormatter={(v: number) => `${v.toFixed(0)}`}
                                label={{ value: 'kW', position: 'insideTopLeft', offset: 0, style: { fontSize: 10, fill: '#94a3b8' } }}
                            />
                            <Tooltip content={<CustomTooltip />} />
                            {showDc && (
                                <Area
                                    name="Tổng công suất đầu vào"
                                    type="monotone"
                                    dataKey="dcPower"
                                    stroke={COLORS.dcPower.stroke}
                                    strokeWidth={2}
                                    fill={`url(#dcGrad-${inverterId})`}
                                    dot={false}
                                    activeDot={{ r: 4, fill: COLORS.dcPower.stroke, stroke: '#fff', strokeWidth: 2 }}
                                    connectNulls
                                    isAnimationActive={true}
                                    animationDuration={CHART_ANIMATION_DURATION}
                                    animationEasing="ease-in-out"
                                />
                            )}
                            {showAc && (
                                <Area
                                    name="Công suất thuần"
                                    type="monotone"
                                    dataKey="acPower"
                                    stroke={COLORS.acPower.stroke}
                                    strokeWidth={2}
                                    fill={`url(#acGrad-${inverterId})`}
                                    dot={false}
                                    activeDot={{ r: 4, fill: COLORS.acPower.stroke, stroke: '#fff', strokeWidth: 2 }}
                                    connectNulls
                                    isAnimationActive={true}
                                    animationDuration={CHART_ANIMATION_DURATION}
                                    animationEasing="ease-in-out"
                                />
                            )}
                        </AreaChart>
                    </ResponsiveContainer>
                )}
            </div>
        </div>
    );
};
