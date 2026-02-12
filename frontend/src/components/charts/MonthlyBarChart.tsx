import React from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';

interface MonthlyDataPoint {
    date: string;
    site1MaxPower: number | null;
    site1MaxIrrad: number | null;
    site2MaxPower: number | null;
    site2MaxIrrad: number | null;
}

interface MonthlyBarChartProps {
    data: MonthlyDataPoint[];
    visibleSites: { site1: boolean; site2: boolean };
}

// Reuse helper for consistency
const getNiceTicks = (dataMax: number, tickCount: number = 6) => {
    if (dataMax === 0) return [0, 2000, 4000, 6000, 8000, 10000];
    const step = Math.ceil(dataMax / (tickCount - 1) / 100) * 100;
    const ticks = [];
    for (let i = 0; i < tickCount; i++) {
        ticks.push(i * step);
    }
    return ticks;
};

export const MonthlyBarChart: React.FC<MonthlyBarChartProps> = ({ data, visibleSites }) => {
    // Determine max power for scaling
    const maxPower = Math.max(
        ...data.map(d => Math.max(d.site1MaxPower || 0, d.site2MaxPower || 0)),
        100
    );
    const leftTicks = getNiceTicks(maxPower, 6);
    const rightTicks = [0, 300, 600, 900, 1200, 1500];

    return (
        <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data} margin={{ top: 5, right: 5, bottom: 0, left: 0 }} barCategoryGap="20%">
                <CartesianGrid strokeDasharray="3 3" stroke="#8d8e8fff" horizontal={false} vertical={false} />
                <XAxis dataKey="date" stroke="#94a3b8" tick={{ fontSize: 10 }} tickMargin={6} />
                <YAxis
                    yAxisId="left"
                    tick={{ fontSize: 10, fill: '#94a3b8' }}
                    tickLine={false}
                    axisLine={false}
                    width={45}
                    domain={[leftTicks[0], leftTicks[leftTicks.length - 1]]}
                    ticks={leftTicks}
                    label={{ value: 'kW', angle: -90, position: 'insideLeft', offset: 10, style: { fontSize: 10, fill: '#94a3b8' } }}
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
                    label={{ value: 'W/m²', angle: 90, position: 'insideRight', offset: 10, style: { fontSize: 10, fill: '#f97316' } }}
                />
                <Tooltip
                    content={({ active, payload, label }) => {
                        if (active && payload && payload.length) {
                            const d = payload[0]?.payload as MonthlyDataPoint;
                            return (
                                <div className="bg-white p-3 border border-slate-200 rounded-lg shadow-xl text-xs min-w-[200px]">
                                    <p className="text-slate-500 font-medium mb-2">Ngày: {label}</p>
                                    <div className="space-y-2">
                                        {visibleSites.site1 && (
                                            <div className="border-b pb-1 last:border-0 last:pb-0">
                                                <div className="flex justify-between text-green-600">
                                                    <span>Shundao 1 CS Max : </span>
                                                    <span>{d?.site1MaxPower?.toLocaleString() || 0} kW</span>
                                                </div>
                                                <div className="flex justify-between text-orange-500">
                                                    <span>Shundao 1 BX Max: </span>
                                                    <span>{d?.site1MaxIrrad?.toFixed(0) || 0} W/m²</span>
                                                </div>
                                            </div>
                                        )}
                                        {visibleSites.site2 && (
                                            <div className="border-b pb-1 last:border-0 last:pb-0">
                                                <div className="flex justify-between text-blue-600">
                                                    <span>Shundao 2 CS Max : </span>
                                                    <span>{d?.site2MaxPower?.toLocaleString() || 0} kW</span>
                                                </div>
                                                <div className="flex justify-between text-purple-500">
                                                    <span>Shundao 2 BX Max : </span>
                                                    <span>{d?.site2MaxIrrad?.toFixed(0) || 0} W/m²</span>
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            );
                        }
                        return null;
                    }}
                    cursor={{ fill: 'rgba(0,0,0,0.05)' }}
                />
                <Legend wrapperStyle={{ fontSize: '11px', paddingTop: '5px' }} />

                {visibleSites.site1 && (
                    <>
                        <Bar yAxisId="left" dataKey="site1MaxPower" name="Shundao 1 CS Max" fill="#22c55e" radius={[4, 4, 0, 0]} />
                        <Bar yAxisId="right" dataKey="site1MaxIrrad" name="Shundao 1 BX Max" fill="#fb923c" radius={[4, 4, 0, 0]} />
                    </>
                )}

                {visibleSites.site2 && (
                    <>
                        <Bar yAxisId="left" dataKey="site2MaxPower" name="Shundao 2 CS Max" fill="#0ea5e9" radius={[4, 4, 0, 0]} />
                        <Bar yAxisId="right" dataKey="site2MaxIrrad" name="Shundao 2 BX Max" fill="#a855f7" radius={[4, 4, 0, 0]} />
                    </>
                )}
            </BarChart>
        </ResponsiveContainer>
    );
};
