import React, { useState, useRef, useCallback, useEffect, useMemo } from 'react';
import { AreaChart, Area, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import type { ProductionDataPoint } from '../../types';
import { CHART_MIN_VISIBLE_POINTS, CHART_X_AXIS_TICK_COUNT } from '../../config/constants';

interface DailyLineChartProps {
    data: ProductionDataPoint[];
    visibleSites: { site1: boolean; site2: boolean };
}

// Helper to get nice round ticks for the left axis
const getNiceTicks = (dataMax: number, tickCount: number = 6) => {
    if (dataMax === 0) return [0, 2000, 4000, 6000, 8000, 10000];
    const step = Math.ceil(dataMax / (tickCount - 1) / 100) * 100; // Round step to nearest 100
    const ticks = [];
    for (let i = 0; i < tickCount; i++) {
        ticks.push(i * step);
    }
    return ticks;
};

export const DailyLineChart: React.FC<DailyLineChartProps> = ({ data, visibleSites }) => {
    const [zoomStart, setZoomStart] = useState(0);
    const [zoomEnd, setZoomEnd] = useState(data.length);
    const chartContainerRef = useRef<HTMLDivElement>(null);

    // Initial sync
    useEffect(() => {
        if (data.length > 0) {
            setZoomStart(0);
            setZoomEnd(data.length);
        }
    }, [data.length]);

    // Handle Ctrl + Scroll zoom
    const handleWheel = useCallback((e: WheelEvent) => {
        if (!e.ctrlKey || data.length === 0) return;
        e.preventDefault();
        e.stopPropagation();

        const container = chartContainerRef.current;
        if (!container) return;

        const rect = container.getBoundingClientRect();
        const chartLeft = 45; // Approx Y-Axis width
        const chartRight = 15; // Approx Right padding
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

    // Calculate max power to determine left axis ticks
    const maxPower = Math.max(
        ...visibleData.map(d => Math.max(d.site1DailyEnergy || 0, d.site2DailyEnergy || 0)),
        100
    );

    // Generate exactly 6 ticks for left axis
    const leftTicks = getNiceTicks(maxPower, 6);
    const rightTicks = [0, 300, 600, 900, 1200, 1500];

    return (
        <div ref={chartContainerRef} className="w-full h-full">
            <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={visibleData} margin={{ top: 10, right: 10, bottom: 10, left: 10 }}>
                    <defs>
                        {/* Shundao 1 Gradients */}
                        <linearGradient id="site1PowerGrad" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#22c55e" stopOpacity={0.25} />
                            <stop offset="95%" stopColor="#22c55e" stopOpacity={0.02} />
                        </linearGradient>
                        <linearGradient id="site1IrradGrad" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#fb923c" stopOpacity={0.25} />
                            <stop offset="95%" stopColor="#fb923c" stopOpacity={0.02} />
                        </linearGradient>
                        {/* Shundao 2 Gradients */}
                        <linearGradient id="site2PowerGrad" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#0ea5e9" stopOpacity={0.25} />
                            <stop offset="95%" stopColor="#0ea5e9" stopOpacity={0.02} />
                        </linearGradient>
                        <linearGradient id="site2IrradGrad" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#a855f7" stopOpacity={0.25} />
                            <stop offset="95%" stopColor="#a855f7" stopOpacity={0.02} />
                        </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="1 1" stroke="#8d8e8fff" horizontal={false} vertical={false} />
                    <XAxis
                        dataKey="date"
                        stroke="#94a3b8"
                        tick={{ fontSize: 10, fill: '#94a3b8' }}
                        tickLine={false}
                        axisLine={{ stroke: '#e2e8f0' }}
                        interval={Math.max(0, Math.floor(visibleData.length / CHART_X_AXIS_TICK_COUNT) - 1)}
                    />
                    <YAxis
                        yAxisId="left"
                        tick={{ fontSize: 10, fill: '#0d09ffff' }}
                        tickLine={false}
                        axisLine={false}
                        width={45}
                        domain={[leftTicks[0], leftTicks[leftTicks.length - 1]]}
                        ticks={leftTicks}
                        tickFormatter={(v: number) => `${v.toLocaleString()}`}
                        label={{ value: 'kW', angle: -90, position: 'insideLeft', offset: 0, style: { fontSize: 10, fill: '#0d09ffff' } }}
                    />
                    <YAxis
                        yAxisId="right"
                        orientation="right"
                        tick={{ fontSize: 10, fill: '#f97316' }}
                        tickLine={false}
                        axisLine={false}
                        width={45}
                        domain={[0, 1500]}
                        ticks={rightTicks}
                        label={{ value: 'W/m²', angle: 90, position: 'insideRight', offset: 0, style: { fontSize: 10, fill: '#f97316' } }}
                    />
                    <Tooltip
                        content={({ active, payload, label }) => {
                            if (active && payload && payload.length) {
                                const d = payload[0]?.payload as ProductionDataPoint;
                                return (
                                    <div className="bg-white p-3 border border-slate-200 rounded-lg shadow-xl text-xs min-w-[200px]">
                                        <p className="text-slate-500 font-medium mb-2">Thời gian: {label}</p>
                                        <div className="space-y-2">
                                            {visibleSites.site1 && (
                                                <div className="border-b pb-1 last:border-0 last:pb-0">
                                                    <div className="flex justify-between text-green-600 font-semibold">
                                                        <span>Shundao 1 Công suất : </span>
                                                        <span>{d?.site1DailyEnergy?.toLocaleString() || '--'} kW</span>
                                                    </div>
                                                    <div className="flex justify-between text-orange-500">
                                                        <span>Shundao 1 Bức xạ : </span>
                                                        <span>{d?.site1Irradiation?.toFixed(0) || '--'} W/m²</span>
                                                    </div>
                                                </div>
                                            )}
                                            {visibleSites.site2 && (
                                                <div className="border-b pb-1 last:border-0 last:pb-0">
                                                    <div className="flex justify-between text-blue-600 font-semibold">
                                                        <span>Shundao 2 Công suất : </span>
                                                        <span>{d?.site2DailyEnergy?.toLocaleString() || '--'} kW</span>
                                                    </div>
                                                    <div className="flex justify-between text-purple-500">
                                                        <span>Shundao 2 Bức xạ : </span>
                                                        <span>{d?.site2Irradiation?.toFixed(0) || '--'} W/m²</span>
                                                    </div>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                );
                            }
                            return null;
                        }}
                    />
                    <Legend wrapperStyle={{ fontSize: '11px', paddingTop: '5px' }} />

                    {/* SHUNDAO 1 */}
                    {visibleSites.site1 && (
                        <>
                            <Area yAxisId="left" type="monotone" dataKey="site1DailyEnergy" name="Shundao 1 Công suất " stroke="#22c55e" strokeWidth={2} fill="url(#site1PowerGrad)" dot={false} activeDot={{ r: 4 }} connectNulls={true} />
                            <Area yAxisId="right" type="monotone" dataKey="site1Irradiation" name="Shundao 1 Bức xạ " stroke="#fb923c" strokeWidth={2} fill="url(#site1IrradGrad)" dot={false} activeDot={{ r: 4 }} connectNulls={true} />
                        </>
                    )}

                    {/* SHUNDAO 2 */}
                    {visibleSites.site2 && (
                        <>
                            <Area yAxisId="left" type="monotone" dataKey="site2DailyEnergy" name="Shundao 2 Công suất" stroke="#0ea5e9" strokeWidth={2} fill="url(#site2PowerGrad)" dot={false} activeDot={{ r: 4 }} connectNulls={true} />
                            <Area yAxisId="right" type="monotone" dataKey="site2Irradiation" name="Shundao 2 Bức xạ" stroke="#a855f7" strokeWidth={2} fill="url(#site2IrradGrad)" dot={false} activeDot={{ r: 4 }} connectNulls={true} />
                        </>
                    )}
                </AreaChart>
            </ResponsiveContainer>
        </div>
    );
};
