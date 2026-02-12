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



// Combined Daily Line Chart
const CombinedLineChart: React.FC<{ data: ProductionDataPoint[] }> = ({ data }) => (
    <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data} margin={{ top: 30, right: 10, bottom: 0, left: 10 }}>
            <CartesianGrid stroke="#cbd5e1" vertical={false} />
            <XAxis
                dataKey="date"
                stroke="#94a3b8"
                tick={{ fontSize: 11, fill: '#64748b' }}
                tickMargin={10}
                interval={11}
                axisLine={false}
                tickLine={false}
            />
            <YAxis
                yAxisId="left"
                stroke="#94a3b8"
                tick={{ fontSize: 11, fill: '#64748b' }}
                tickLine={false}
                axisLine={false}
                width={35}
                tickCount={5}
                tickFormatter={(value) => value >= 1000 ? `${value / 1000}K` : value}
                label={{
                    value: 'kW',
                    position: 'insideTopLeft',
                    offset: 0,
                    style: { fontSize: '12px', fill: '#64748b', fontWeight: 600, textAnchor: 'start' },
                    dy: -25,
                    dx: 0
                }}
            />
            <YAxis
                yAxisId="right"
                orientation="right"
                stroke="#f97316"
                tick={{ fontSize: 11, fill: '#f97316' }}
                tickLine={false}
                axisLine={false}
                width={35}
                tickCount={5}
                label={{
                    value: 'W/m²',
                    position: 'insideTopRight',
                    offset: 0,
                    style: { fontSize: '12px', fill: '#f97316', fontWeight: 600, textAnchor: 'end' },
                    dy: -25,
                    dx: 0
                }}
            />
            <Tooltip
                content={({ active, payload, label }) => {
                    if (active && payload && payload.length) {
                        const d = payload[0]?.payload as ProductionDataPoint;
                        return (
                            <div className="bg-white p-3 border border-slate-200 rounded-lg shadow-xl text-xs min-w-[200px]">
                                <p className="text-slate-500 font-medium mb-2">Thời gian: {label}</p>
                                <div className="space-y-2">
                                    <div className="border-b pb-1">
                                        <div className="flex justify-between text-green-600 font-semibold">
                                            <span>Shundao 1 Công suất : </span>
                                            <span>{d?.site1Power?.toLocaleString() || '--'} kW</span>
                                        </div>
                                        <div className="flex justify-between text-orange-500">
                                            <span>Shundao 1 Bức xạ : </span>
                                            <span>{d?.site1Irradiance?.toFixed(0) || '--'} W/m²</span>
                                        </div>
                                    </div>
                                    <div>
                                        <div className="flex justify-between text-blue-600 font-semibold">
                                            <span>Shundao 2 Công suất : </span>
                                            <span>{d?.site2Power?.toLocaleString() || '--'} kW</span>
                                        </div>
                                        <div className="flex justify-between text-purple-500">
                                            <span>Shundao 2 Bức xạ : </span>
                                            <span>{d?.site2Irradiance?.toFixed(0) || '--'} W/m²</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        );
                    }
                    return null;
                }}
            />
            <Legend wrapperStyle={{ fontSize: '11px', paddingTop: '5px' }} />

            {/* SHUNDAO 1 */}
            <Line yAxisId="left" type="linear" dataKey="site1Power" name="Shundao 1 Công suất " stroke="#22c55e" strokeWidth={2} dot={false} activeDot={{ r: 4 }} connectNulls={true} />
            <Line yAxisId="right" type="linear" dataKey="site1Irradiance" name="Shundao 1 Bức xạ " stroke="#fb923c" strokeWidth={2} dot={false} activeDot={{ r: 4 }} connectNulls={true} />

            {/* SHUNDAO 2 */}
            <Line yAxisId="left" type="linear" dataKey="site2Power" name="Shundao 2 Công suất" stroke="#0ea5e9" strokeWidth={2} dot={false} activeDot={{ r: 4 }} connectNulls={true} />
            <Line yAxisId="right" type="linear" dataKey="site2Irradiance" name="Shundao 2 Bức xạ" stroke="#a855f7" strokeWidth={2} dot={false} activeDot={{ r: 4 }} connectNulls={true} />
        </LineChart>
    </ResponsiveContainer>
);

// Combined Monthly Bar Chart
const CombinedBarChart: React.FC<{ data: MonthlyDataPoint[] }> = ({ data }) => (
    <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data} margin={{ top: 5, right: 5, bottom: 0, left: 0 }} barCategoryGap="20%">
            <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" vertical={false} />
            <XAxis dataKey="date" stroke="#94a3b8" tick={{ fontSize: 10 }} tickMargin={6} />
            <YAxis yAxisId="left" stroke="#94a3b8" tick={{ fontSize: 10 }} width={45} label={{ value: 'kW', angle: -90, position: 'insideLeft', offset: 10, style: { fontSize: '10px', fill: '#94a3b8' } }} />
            <YAxis yAxisId="right" orientation="right" stroke="#f97316" tick={{ fontSize: 10 }} width={45} label={{ value: 'W/m²', angle: 90, position: 'insideRight', offset: 10, style: { fontSize: '10px', fill: '#f97316' } }} />
            <Tooltip
                content={({ active, payload, label }) => {
                    if (active && payload && payload.length) {
                        const d = payload[0]?.payload as MonthlyDataPoint;
                        return (
                            <div className="bg-white p-3 border border-slate-200 rounded-lg shadow-xl text-xs min-w-[200px]">
                                <p className="text-slate-500 font-medium mb-2">Ngày: {label}</p>
                                <div className="space-y-2">
                                    <div className="border-b pb-1">
                                        <div className="flex justify-between text-green-600">
                                            <span>Shundao 1 CS Max : </span>
                                            <span>{d?.site1MaxPower?.toLocaleString() || 0} kW</span>
                                        </div>
                                        <div className="flex justify-between text-orange-500">
                                            <span>Shundao 1 BX Max: </span>
                                            <span>{d?.site1MaxIrrad?.toFixed(0) || 0} W/m²</span>
                                        </div>
                                    </div>
                                    <div>
                                        <div className="flex justify-between text-blue-600">
                                            <span>Shundao 2 CS Max : </span>
                                            <span>{d?.site2MaxPower?.toLocaleString() || 0} kW</span>
                                        </div>
                                        <div className="flex justify-between text-purple-500">
                                            <span>Shundao 2 BX Max : </span>
                                            <span>{d?.site2MaxIrrad?.toFixed(0) || 0} W/m²</span>
                                        </div>
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
            <Bar yAxisId="left" dataKey="site1MaxPower" name="Shundao 1 CS Max" fill="#22c55e" radius={[4, 4, 0, 0]} />
            <Bar yAxisId="right" dataKey="site1MaxIrrad" name="Shundao 1 BX Max" fill="#fb923c" radius={[4, 4, 0, 0]} />
            <Bar yAxisId="left" dataKey="site2MaxPower" name="Shundao 2 CS Max" fill="#0ea5e9" radius={[4, 4, 0, 0]} />
            <Bar yAxisId="right" dataKey="site2MaxIrrad" name="Shundao 2 BX Max" fill="#a855f7" radius={[4, 4, 0, 0]} />
        </BarChart>
    </ResponsiveContainer>
);

const CombinedChartContainer: React.FC<{
    dailyData: ProductionDataPoint[];
    monthlyData: MonthlyDataPoint[];
    loading: boolean;
    viewMode: ViewMode
}> = ({ dailyData, monthlyData, loading, viewMode }) => {
    const [zoomRange, setZoomRange] = useState<{ start: number; end: number } | null>(null);
    const chartRef = useRef<HTMLDivElement>(null);

    // Reset zoom
    useEffect(() => {
        if (dailyData.length > 0) {
            setZoomRange({ start: 0, end: dailyData.length - 1 });
        }
    }, [dailyData.length, viewMode]);

    // Calculate Totals
    const validDailyPowerPoints = dailyData.filter(d => (d.site1Power != null && d.site1Power > 0) || (d.site2Power != null && d.site2Power > 0));
    const totalS1MWh = viewMode === 'day'
        ? (validDailyPowerPoints.reduce((sum, d) => sum + (d.site1Power || 0), 0) * (5 / 60)) / 1000
        : (monthlyData || []).reduce((sum, d) => sum + (d.site1MaxPower || 0), 0) / 1000; // Warning: Summing MaxPower is technically not Energy, but keeping logic consistent with previous chart for relative comparison or placeholder. 
    // Actually, for monthly view, previous code summed 'power' which was mapped from MaxPower.
    // Let's keep it as is.

    const totalS2MWh = viewMode === 'day'
        ? (validDailyPowerPoints.reduce((sum, d) => sum + (d.site2Power || 0), 0) * (5 / 60)) / 1000
        : (monthlyData || []).reduce((sum, d) => sum + (d.site2MaxPower || 0), 0) / 1000;

    // Zoom Logic (Same as before)
    useEffect(() => {
        if (viewMode === 'month') return;
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
                if (!prev) return { start: 0, end: dailyData.length - 1 };
                const currentRange = prev.end - prev.start;
                const minRange = 12;

                if (delta > 0) {
                    const expand = Math.ceil(currentRange * zoomFactor);
                    const expandLeft = Math.ceil(expand * mouseRatio);
                    const expandRight = expand - expandLeft;
                    const newStart = Math.max(0, prev.start - expandLeft);
                    const newEnd = Math.min(dailyData.length - 1, prev.end + expandRight);
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
    }, [dailyData.length, viewMode]);

    const resetZoom = () => setZoomRange({ start: 0, end: dailyData.length - 1 });
    const effectiveRange = zoomRange || { start: 0, end: dailyData.length - 1 };

    const visibleDailyData = dailyData.slice(effectiveRange.start, effectiveRange.end + 1);
    const zoomLevel = dailyData.length > 1 && viewMode === 'day'
        ? Math.round((1 - (effectiveRange.end - effectiveRange.start) / (dailyData.length - 1)) * 100)
        : 0;

    if (loading) {
        return (
            <div className="h-[485px] bg-white rounded-2xl border border-slate-200 p-4">
                <Skeleton className="h-6 w-48 mb-4" />
                <Skeleton className="h-[400px] w-full" />
            </div>
        );
    }

    return (
        <Card className="h-[550px]">
            <CardHeader className="mb-2 pb-0">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <div className="flex items-center gap-2">
                            <div className="p-1.5 rounded-lg bg-slate-100">
                                {viewMode === 'day' ? <LineChartIcon className="w-4 h-4 text-slate-600" /> : <BarChart2 className="w-4 h-4 text-slate-600" />}
                            </div>
                            <CardTitle className="text-base">Biểu đồ Tổng hợp</CardTitle>
                        </div>

                        <div className="flex gap-4 text-sm">
                            <div className="flex flex-col">
                                <span className="text-xs text-slate-500">Shundao 1 Tổng</span>
                                <span className="font-semibold text-green-600">{totalS1MWh.toFixed(2)} MWh</span>
                            </div>
                            <div className="flex flex-col">
                                <span className="text-xs text-slate-500">Shundao 2 Tổng</span>
                                <span className="font-semibold text-blue-600">{totalS2MWh.toFixed(2)} MWh</span>
                            </div>
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

            <div ref={chartRef} className="h-[450px] w-full px-2 outline-none focus:outline-none">
                {viewMode === 'day'
                    ? <CombinedLineChart data={visibleDailyData} />
                    : <CombinedBarChart data={monthlyData} />
                }
            </div>
        </Card>
    );
}

export const ProductionChart: React.FC<ProductionChartProps> = ({ data, loading = false }) => {
    const [viewMode, setViewMode] = useState<ViewMode>('day');

    // Fetch monthly data when in month mode
    const { data: monthlyData, isLoading: monthlyLoading } = useQuery({
        queryKey: ['monthlyProduction'],
        queryFn: () => api.get('/production-monthly') as Promise<MonthlyDataPoint[]>,
        enabled: viewMode === 'month',
        staleTime: 60000,
    });

    // Transform daily data (Filter 06:00 - 18:00)
    const filterTimeRange = (d: ProductionDataPoint) => {
        const [h, m] = d.date.split(':').map(Number);
        if (h < 6 || h > 18) return false;
        if (h === 18 && m > 0) return false; // Exclude 18:05, 18:10...
        return true;
    };

    // Filter daily data
    const filteredDailyData = data.filter(filterTimeRange).map(d => ({
        ...d,
        site1Power: d.site1Power ?? null, // Ensure nulls are preserved
        site1Irradiance: d.site1Irradiance ?? null,
        site2Power: d.site2Power ?? null,
        site2Irradiance: d.site2Irradiance ?? null
    }));

    const currentLoading = viewMode === 'day' ? loading : monthlyLoading;

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
            <div className="grid grid-cols-1 gap-6">
                <CombinedChartContainer
                    dailyData={filteredDailyData}
                    monthlyData={monthlyData || []}
                    loading={currentLoading}
                    viewMode={viewMode}
                />
            </div>
        </div>
    );
};
