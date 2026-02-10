import React, { useState, useRef, useEffect } from 'react';
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { useQuery } from '@tanstack/react-query';
import { Card, CardHeader, CardTitle } from '../ui/Card';
import { Skeleton } from '../ui/Skeleton';
import { LineChart as LineChartIcon, BarChart2, RotateCcw, Calendar, Clock } from 'lucide-react';
import api from '../../services/api';

interface ProductionDataPoint {
    date: string;
    site1Power: number;
    site1Irradiance: number;
    site2Power: number;
    site2Irradiance: number;
}

interface MonthlyDataPoint {
    date: string;
    site1MaxPower: number | null;
    site1MaxIrrad: number | null;
    site2MaxPower: number | null;
    site2MaxIrrad: number | null;
}

interface ProductionChartProps {
    data: ProductionDataPoint[];
    loading?: boolean;
}

type ViewMode = 'day' | 'month';

interface SingleSiteChartProps {
    siteName: string;
    data: { date: string; power: number; irradiance: number }[];
    color: string;
    loading?: boolean;
    viewMode: ViewMode;
}

// Daily Line Chart Component
const DailyLineChart: React.FC<{ data: any[]; color: string }> = ({ data, color }) => (
    <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data} margin={{ top: 5, right: 5, bottom: 0, left: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" vertical={false} />
            <XAxis dataKey="date" stroke="#94a3b8" tick={{ fontSize: 10 }} tickMargin={6} minTickGap={35} />
            <YAxis yAxisId="left" stroke="#94a3b8" tick={{ fontSize: 10 }} width={40} label={{ value: 'kW', angle: -90, position: 'insideLeft', offset: 10, style: { fontSize: '10px', fill: '#94a3b8' } }} />
            <YAxis yAxisId="right" orientation="right" stroke="#f97316" tick={{ fontSize: 10 }} width={40} label={{ value: 'W/m²', angle: 90, position: 'insideRight', offset: 10, style: { fontSize: '10px', fill: '#f97316' } }} />
            <Tooltip
                content={({ active, payload, label }) => {
                    if (active && payload && payload.length) {
                        const d = payload[0]?.payload;
                        return (
                            <div className="bg-white p-3 border border-slate-200 rounded-lg shadow-xl text-xs min-w-[180px]">
                                <p className="text-slate-500 font-medium mb-2">Thời gian: {label}</p>
                                <div className="space-y-1">
                                    <div className="flex justify-between">
                                        <span>Công suất</span>
                                        <span className="font-mono font-medium">{d?.power?.toLocaleString() || 0} kW</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span>Bức xạ</span>
                                        <span className="font-mono font-medium">{d?.irradiance?.toFixed(0) || 0} W/m²</span>
                                    </div>
                                </div>
                            </div>
                        );
                    }
                    return null;
                }}
            />
            <Legend wrapperStyle={{ fontSize: '11px', paddingTop: '5px' }} />
            <Line yAxisId="left" type="monotone" dataKey="power" name="Công suất" stroke={color} strokeWidth={2} dot={false} activeDot={{ r: 4 }} connectNulls={true} />
            <Line yAxisId="right" type="monotone" dataKey="irradiance" name="Bức xạ" stroke="#fb923c" strokeWidth={2} dot={false} activeDot={{ r: 4 }} connectNulls={true} />
        </LineChart>
    </ResponsiveContainer>
);

// Monthly Bar Chart Component
const MonthlyBarChart: React.FC<{ data: any[]; color: string }> = ({ data, color }) => (
    <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data} margin={{ top: 5, right: 5, bottom: 0, left: 0 }} barCategoryGap="20%">
            <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" vertical={false} />
            <XAxis dataKey="date" stroke="#94a3b8" tick={{ fontSize: 10 }} tickMargin={6} />
            <YAxis yAxisId="left" stroke="#94a3b8" tick={{ fontSize: 10 }} width={45} label={{ value: 'kW', angle: -90, position: 'insideLeft', offset: 10, style: { fontSize: '10px', fill: '#94a3b8' } }} />
            <YAxis yAxisId="right" orientation="right" stroke="#f97316" tick={{ fontSize: 10 }} width={45} label={{ value: 'W/m²', angle: 90, position: 'insideRight', offset: 10, style: { fontSize: '10px', fill: '#f97316' } }} />
            <Tooltip
                content={({ active, payload, label }) => {
                    if (active && payload && payload.length) {
                        const d = payload[0]?.payload;
                        return (
                            <div className="bg-white p-3 border border-slate-200 rounded-lg shadow-xl text-xs min-w-[180px]">
                                <p className="text-slate-500 font-medium mb-2">Ngày: {label}</p>
                                <div className="space-y-1">
                                    <div className="flex justify-between">
                                        <span>CS Max</span>
                                        <span className="font-mono font-medium text-green-600">{d?.power?.toLocaleString() || 0} kW</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span>BX Max</span>
                                        <span className="font-mono font-medium text-orange-500">{d?.irradiance?.toFixed(0) || 0} W/m²</span>
                                    </div>
                                </div>
                            </div>
                        );
                    }
                    return null;
                }}
                cursor={{ fill: 'rgba(0,0,0,0.05)' }}
            />
            <Legend wrapperStyle={{ fontSize: '11px', paddingTop: '5px' }} />
            <Bar yAxisId="left" dataKey="power" name="CS Max (kW)" fill={color} radius={[4, 4, 0, 0]} />
            <Bar yAxisId="right" dataKey="irradiance" name="BX Max (W/m²)" fill="#fb923c" radius={[4, 4, 0, 0]} />
        </BarChart>
    </ResponsiveContainer>
);

const SingleSiteChart: React.FC<SingleSiteChartProps> = ({ siteName, data, color, loading, viewMode }) => {
    const [zoomRange, setZoomRange] = useState<{ start: number; end: number } | null>(null);
    const chartRef = useRef<HTMLDivElement>(null);

    // Reset zoom when data or view mode changes
    useEffect(() => {
        if (data.length > 0) {
            setZoomRange({ start: 0, end: data.length - 1 });
        }
    }, [data.length, viewMode]);

    // Calculate Energy/Summary
    const validPowerPoints = data.filter(d => d.power != null && d.power > 0);
    const totalEnergyMWh = viewMode === 'day'
        ? (validPowerPoints.reduce((sum, d) => sum + (d.power || 0), 0) * (5 / 60)) / 1000
        : validPowerPoints.reduce((sum, d) => sum + (d.power || 0), 0) / 1000;

    // Wheel zoom handler (only for daily view)
    useEffect(() => {
        if (viewMode === 'month') return; // No zoom for bar chart

        const chartElement = chartRef.current;
        if (!chartElement) return;

        const handleWheel = (e: WheelEvent) => {
            if (!e.ctrlKey) return;
            e.preventDefault();
            e.stopPropagation();

            const rect = chartElement.getBoundingClientRect();
            const chartLeftMargin = 50;
            const chartRightMargin = 50;
            const chartWidth = rect.width - chartLeftMargin - chartRightMargin;
            const mouseX = e.clientX - rect.left - chartLeftMargin;
            const mouseRatio = Math.max(0, Math.min(1, mouseX / chartWidth));

            const delta = e.deltaY > 0 ? 1 : -1;
            const zoomFactor = 0.15;

            setZoomRange(prev => {
                if (!prev) return { start: 0, end: data.length - 1 };
                const currentRange = prev.end - prev.start;
                const minRange = 12;

                if (delta > 0) {
                    const expand = Math.ceil(currentRange * zoomFactor);
                    const expandLeft = Math.ceil(expand * mouseRatio);
                    const expandRight = expand - expandLeft;
                    const newStart = Math.max(0, prev.start - expandLeft);
                    const newEnd = Math.min(data.length - 1, prev.end + expandRight);
                    return { start: newStart, end: newEnd };
                } else {
                    if (currentRange > minRange) {
                        const shrink = Math.ceil(currentRange * zoomFactor);
                        const shrinkLeft = Math.ceil(shrink * mouseRatio);
                        const shrinkRight = shrink - shrinkLeft;
                        const newStart = prev.start + shrinkLeft;
                        const newEnd = prev.end - shrinkRight;
                        if (newEnd - newStart >= minRange) {
                            return { start: newStart, end: newEnd };
                        }
                    }
                    return prev;
                }
            });
        };

        chartElement.addEventListener('wheel', handleWheel, { passive: false });
        return () => chartElement.removeEventListener('wheel', handleWheel);
    }, [data.length, viewMode]);

    const resetZoom = () => setZoomRange({ start: 0, end: data.length - 1 });

    const effectiveRange = zoomRange || { start: 0, end: data.length - 1 };
    const visibleData = viewMode === 'day'
        ? data.slice(effectiveRange.start, effectiveRange.end + 1)
        : data; // Show all data for monthly view
    const zoomLevel = data.length > 1 && viewMode === 'day'
        ? Math.round((1 - (effectiveRange.end - effectiveRange.start) / (data.length - 1)) * 100)
        : 0;

    if (loading) {
        return (
            <div className="h-[240px] bg-white rounded-2xl border border-slate-200 p-4">
                <div className="mb-2 space-y-1">
                    <Skeleton className="h-4 w-28" />
                    <Skeleton className="h-3 w-20" />
                </div>
                <div className="h-[160px] w-full flex items-end gap-1">
                    {Array.from({ length: 8 }).map((_, i) => (
                        <Skeleton key={i} className="w-full" style={{ height: `${Math.random() * 60 + 20}%` }} />
                    ))}
                </div>
            </div>
        );
    }

    return (
        <Card className="h-[485px]">
            <CardHeader className="mb-2 pb-0">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        <div className="p-1.5 rounded-lg" style={{ backgroundColor: `${color}15` }}>
                            {viewMode === 'day'
                                ? <LineChartIcon className="w-4 h-4" style={{ color }} />
                                : <BarChart2 className="w-4 h-4" style={{ color }} />
                            }
                        </div>
                        <div>
                            <CardTitle className="text-base">{siteName}</CardTitle>
                            <p className="text-slate-500 text-xs">
                                {viewMode === 'day' ? 'Tổng: ' : 'Tổng Max: '}
                                <span className="font-semibold text-slate-700">{totalEnergyMWh.toFixed(2)} MWh</span>
                            </p>
                        </div>
                    </div>
                    <div className="flex items-center gap-1">
                        {viewMode === 'day' && zoomLevel > 0 && (
                            <>
                                <span className="text-xs text-slate-400 mr-1">Zoom: {zoomLevel}%</span>
                                <button onClick={resetZoom} className="p-1 rounded hover:bg-slate-100 text-slate-400 hover:text-slate-600" title="Reset zoom">
                                    <RotateCcw className="w-4 h-4" />
                                </button>
                            </>
                        )}
                        {viewMode === 'day' && <span className="text-xs text-slate-300 ml-2">Ctrl + Scroll</span>}
                    </div>
                </div>
            </CardHeader>

            <div ref={chartRef} className="h-[380px] w-full px-2" style={{ minWidth: '100px', minHeight: '160px' }}>
                {viewMode === 'day'
                    ? <DailyLineChart data={visibleData} color={color} />
                    : <MonthlyBarChart data={visibleData} color={color} />
                }
            </div>
        </Card>
    );
};

export const ProductionChart: React.FC<ProductionChartProps> = ({ data, loading = false }) => {
    const [viewMode, setViewMode] = useState<ViewMode>('day');

    // Fetch monthly data when in month mode
    const { data: monthlyData, isLoading: monthlyLoading } = useQuery({
        queryKey: ['monthlyProduction'],
        queryFn: () => api.get('/production-monthly') as Promise<MonthlyDataPoint[]>,
        enabled: viewMode === 'month',
        staleTime: 60000,
    });

    // Transform daily data
    const site1DayData = data.map(d => ({
        date: d.date,
        power: d.site1Power || 0,
        irradiance: d.site1Irradiance
    }));

    const site2DayData = data.map(d => ({
        date: d.date,
        power: d.site2Power || 0,
        irradiance: d.site2Irradiance
    }));

    // Transform monthly data
    const site1MonthData = (monthlyData || []).map(d => ({
        date: d.date,
        power: d.site1MaxPower || 0,
        irradiance: d.site1MaxIrrad || 0
    }));

    const site2MonthData = (monthlyData || []).map(d => ({
        date: d.date,
        power: d.site2MaxPower || 0,
        irradiance: d.site2MaxIrrad || 0
    }));

    const currentLoading = viewMode === 'day' ? loading : monthlyLoading;
    const site1Data = viewMode === 'day' ? site1DayData : site1MonthData;
    const site2Data = viewMode === 'day' ? site2DayData : site2MonthData;

    return (
        <div className="space-y-4">
            {/* View Mode Toggle */}
            <div className="flex justify-end">
                <div className="inline-flex bg-slate-100 rounded-lg p-1">
                    <button
                        onClick={() => setViewMode('day')}
                        className={`px-3 py-1.5 text-sm font-medium rounded-md flex items-center gap-1.5 transition-colors ${viewMode === 'day'
                            ? 'bg-white text-slate-800 shadow-sm'
                            : 'text-slate-500 hover:text-slate-700'
                            }`}
                    >
                        <Clock size={14} />
                        Hôm nay
                    </button>
                    <button
                        onClick={() => setViewMode('month')}
                        className={`px-3 py-1.5 text-sm font-medium rounded-md flex items-center gap-1.5 transition-colors ${viewMode === 'month'
                            ? 'bg-white text-slate-800 shadow-sm'
                            : 'text-slate-500 hover:text-slate-700'
                            }`}
                    >
                        <Calendar size={14} />
                        Theo tháng
                    </button>
                </div>
            </div>

            {/* Charts Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <SingleSiteChart
                    siteName="SHUNDAO 1"
                    data={site1Data}
                    color="#22c55e"
                    loading={currentLoading}
                    viewMode={viewMode}
                />
                <SingleSiteChart
                    siteName="SHUNDAO 2"
                    data={site2Data}
                    color="#0ea5e9"
                    loading={currentLoading}
                    viewMode={viewMode}
                />
            </div>
        </div>
    );
};
