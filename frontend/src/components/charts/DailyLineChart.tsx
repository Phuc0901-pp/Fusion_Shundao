import React, { useMemo } from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import type { ProductionDataPoint } from '../../types';
import { CHART_X_AXIS_TICK_COUNT } from '../../config/constants';
import { useChartZoom } from '../../hooks/useChartZoom';
import { useYAxisTicks } from '../../hooks/useYAxisTicks';

interface DailyLineChartProps {
    data: ProductionDataPoint[];
    visibleSites: { site1: boolean; site2: boolean };
}

// Static ticks for the irradiance (right) axis – never changes
const RIGHT_TICKS = [0, 300, 600, 900, 1200, 1500];

// DailyLineChart is memoized to avoid costly re-renders when parent state
// updates (e.g. hover, sidebar toggle) don't affect chart data.
const DailyLineChartInner: React.FC<DailyLineChartProps> = ({ data, visibleSites }) => {
    // ── Custom Hooks ──────────────────────────────────────────────────────────
    // Trim trailing incomplete data points to prevent the chart from dropping to 0
    // or showing '--' in tooltips during the ~30s it takes the crawler to fetch all devices.
    const sanitizedData = useMemo(() => {
        let validEnd = data.length;
        for (let i = data.length - 1; i >= 0; i--) {
            const d = data[i];
            // Check if this point has all required metrics for the currently visible sites
            const s1Ok = !visibleSites.site1 || (d.site1DailyEnergy != null && d.site1Irradiation != null);
            const s2Ok = !visibleSites.site2 || (d.site2DailyEnergy != null && d.site2Irradiation != null);
            
            if (s1Ok && s2Ok) {
                validEnd = i + 1;
                break;
            }
        }
        // If everything was incomplete, just return original to avoid blank chart
        if (validEnd === 0) return data;
        return data.slice(0, validEnd);
    }, [data, visibleSites]);

    // All zoom logic lives in useChartZoom
    const { containerRef, visibleRange } = useChartZoom({ dataLength: sanitizedData.length });

    // Slice data to what's visible in the current zoom window
    const visibleData = useMemo(
        () => sanitizedData.slice(visibleRange.start, visibleRange.end),
        [sanitizedData, visibleRange],
    );

    // Find max power across visible data for the left Y-axis
    const maxPower = useMemo(() => {
        let max = 100;
        visibleData.forEach((d) => {
            if (visibleSites.site1) max = Math.max(max, d.site1DailyEnergy || 0);
            if (visibleSites.site2) max = Math.max(max, d.site2DailyEnergy || 0);
        });
        return max;
    }, [visibleData, visibleSites]);

    // Nice ticks from hook
    const leftTicks = useYAxisTicks(maxPower, 6);

    // ── Render ────────────────────────────────────────────────────────────────
    return (
        <div ref={containerRef} className="w-full h-full">
            <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={visibleData} margin={{ top: 10, right: 10, bottom: 10, left: 10 }}>
                    <defs>
                        <linearGradient id="site1PowerGrad" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#22c55e" stopOpacity={0.25} />
                            <stop offset="95%" stopColor="#22c55e" stopOpacity={0.02} />
                        </linearGradient>
                        <linearGradient id="site1IrradGrad" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#fb923c" stopOpacity={0.25} />
                            <stop offset="95%" stopColor="#fb923c" stopOpacity={0.02} />
                        </linearGradient>
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
                        ticks={RIGHT_TICKS}
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
                    <Area yAxisId="left" type="monotone" dataKey="site1DailyEnergy"
                        name="Shundao 1 Công suất " stroke="#22c55e" strokeWidth={2}
                        fill="url(#site1PowerGrad)" dot={false} activeDot={{ r: 4 }}
                        connectNulls hide={!visibleSites.site1} isAnimationActive={false} />
                    <Area yAxisId="right" type="monotone" dataKey="site1Irradiation"
                        name="Shundao 1 Bức xạ " stroke="#fb923c" strokeWidth={2}
                        fill="url(#site1IrradGrad)" dot={false} activeDot={{ r: 4 }}
                        connectNulls hide={!visibleSites.site1} isAnimationActive={false} />

                    {/* SHUNDAO 2 */}
                    <Area yAxisId="left" type="monotone" dataKey="site2DailyEnergy"
                        name="Shundao 2 Công suất" stroke="#0ea5e9" strokeWidth={2}
                        fill="url(#site2PowerGrad)" dot={false} activeDot={{ r: 4 }}
                        connectNulls hide={!visibleSites.site2} isAnimationActive={false} />
                    <Area yAxisId="right" type="monotone" dataKey="site2Irradiation"
                        name="Shundao 2 Bức xạ" stroke="#a855f7" strokeWidth={2}
                        fill="url(#site2IrradGrad)" dot={false} activeDot={{ r: 4 }}
                        connectNulls hide={!visibleSites.site2} isAnimationActive={false} />
                </AreaChart>
            </ResponsiveContainer>
        </div>
    );
};

// Only re-render when data values or site visibility changes
export const DailyLineChart = React.memo(
    DailyLineChartInner,
    (prev, next) => JSON.stringify(prev) === JSON.stringify(next)
);
