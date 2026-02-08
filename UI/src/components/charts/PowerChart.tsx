import React from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { Card, CardHeader, CardTitle } from '../ui/Card';
import { Skeleton } from '../ui/Skeleton';

interface DataPoint {
    time: string;
    // Aggregated
    power: number;

    // Detailed Metrics (available for all sites or specific site)
    gridPower?: number;
    pvPower?: number;
    consumptionPower?: number;

    // Per Site (optional, if we want to separate lines per site)
    [key: string]: string | number | undefined;
}

interface PowerChartProps {
    data: DataPoint[]; // Default/Aggregated
    siteData?: Record<string, DataPoint[]>; // Data per site
    loading?: boolean;
}

export const PowerChart: React.FC<PowerChartProps> = ({ data, siteData, loading = false }) => {
    const [selectedSite, setSelectedSite] = React.useState<string>('all');

    // Filter displayed data based on selection
    const displayedData = React.useMemo(() => {
        if (selectedSite === 'all' || !siteData) return data;
        return siteData[selectedSite] || data;
    }, [selectedSite, data, siteData]);

    if (loading) {
        return (
            <Card className="h-[400px]">
                <div className="mb-6 space-y-2">
                    <Skeleton className="h-6 w-48" />
                    <Skeleton className="h-4 w-32" />
                </div>
                <div className="h-[300px] w-full flex items-end gap-2">
                    {Array.from({ length: 12 }).map((_, i) => (
                        <Skeleton key={i} className="w-full" style={{ height: `${Math.random() * 60 + 20}%` }} />
                    ))}
                </div>
            </Card>
        );
    }

    const CustomTooltip = ({ active, payload, label }: any) => {
        if (active && payload && payload.length) {
            const data = payload[0].payload;
            return (
                <div className="bg-white p-4 border border-slate-200 rounded-lg shadow-xl text-xs min-w-[300px]">
                    <div className="mb-3 border-b border-slate-100 pb-2">
                        <p className="text-slate-500">{new Date().toLocaleDateString()} {label}</p>
                    </div>
                    <div className="grid grid-cols-2 gap-8">
                        {/* Production */}
                        <div>
                            <p className="text-slate-400 mb-2 font-medium">Sản xuất</p>
                            <div className="space-y-2">
                                <div className="flex justify-between items-center">
                                    <div className="flex items-center gap-1.5">
                                        <div className="w-2 h-1 bg-slate-400 rounded-full"></div>
                                        <span className="text-slate-600">Công suất điện lưới</span>
                                    </div>
                                    <span className="font-mono font-medium text-slate-800">
                                        {(data.gridPower || 0).toLocaleString()} kW
                                    </span>
                                </div>
                                <div className="flex justify-between items-center">
                                    <div className="flex items-center gap-1.5">
                                        <div className="w-2 h-1 bg-green-500 rounded-full"></div>
                                        <span className="text-slate-600">Công suất PV</span>
                                    </div>
                                    <span className="font-mono font-medium text-slate-800">
                                        {(data.pvPower || 0).toLocaleString()} kW
                                    </span>
                                </div>
                            </div>
                        </div>

                        {/* Consumption */}
                        <div>
                            <p className="text-slate-400 mb-2 font-medium">Mức sử dụng</p>
                            <div className="space-y-2">
                                <div className="flex justify-between items-center">
                                    <div className="flex items-center gap-1.5">
                                        <div className="w-2 h-1 bg-slate-400 rounded-full"></div>
                                        <span className="text-slate-600">Công suất nguồn cấp</span>
                                    </div>
                                    <span className="font-mono font-medium text-slate-800">--</span>
                                </div>
                                <div className="flex justify-between items-center">
                                    <div className="flex items-center gap-1.5">
                                        <div className="w-2 h-1 bg-amber-500 rounded-full"></div>
                                        <span className="text-slate-600">Công suất tiêu thụ</span>
                                    </div>
                                    <span className="font-mono font-medium text-slate-800">
                                        {(data.consumptionPower || 0).toLocaleString()} kW
                                    </span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            );
        }
        return null;
    };

    return (
        <Card className="h-[400px]">
            <CardHeader className="mb-6 flex flex-row items-center justify-between">
                <div>
                    <CardTitle>Biểu Đồ Công Suất</CardTitle>
                    <p className="text-slate-500 text-sm">Sản lượng thời gian thực (kW)</p>
                </div>

                {/* Site Selector */}
                <select
                    className="bg-slate-50 border border-slate-200 text-slate-700 text-sm rounded-lg focus:ring-solar-500 focus:border-solar-500 block p-2 outline-none"
                    value={selectedSite}
                    onChange={(e) => setSelectedSite(e.target.value)}
                >
                    <option value="all">Tất cả dự án</option>
                    <option value="siteA">Dự án A (Shundao 2)</option>
                    <option value="siteB">Dự án B (Shundao 1)</option>
                </select>
            </CardHeader>

            <div className="h-[300px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={displayedData}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" vertical={false} />
                        <XAxis
                            dataKey="time"
                            stroke="#64748b"
                            tick={{ fontSize: 12 }}
                            tickMargin={10}
                        />
                        <YAxis
                            stroke="#64748b"
                            tick={{ fontSize: 12 }}
                            unit=" kW"
                        />
                        <Tooltip content={<CustomTooltip />} />
                        <Legend wrapperStyle={{ paddingTop: '20px' }} />

                        {selectedSite === 'all' ? (
                            <>
                                {/* Site A Line */}
                                <Line
                                    type="monotone"
                                    dataKey="siteAPower"
                                    name="Dự án Shundao 2 (A)"
                                    stroke="#0ea5e9" // Sky Blue
                                    strokeWidth={3}
                                    dot={false}
                                    activeDot={{ r: 6, strokeWidth: 0 }}
                                />
                                {/* Site B Line */}
                                <Line
                                    type="monotone"
                                    dataKey="siteBPower"
                                    name="Dự án Shundao 1 (B)"
                                    stroke="#8b5cf6" // Violet
                                    strokeWidth={3}
                                    dot={false}
                                    activeDot={{ r: 6, strokeWidth: 0 }}
                                />
                            </>
                        ) : (
                            <>
                                {/* Consumption Line */}
                                <Line
                                    type="monotone"
                                    dataKey="consumptionPower"
                                    name="Tiêu thụ"
                                    stroke="#f59e0b" // Orange
                                    strokeWidth={3}
                                    dot={false}
                                    activeDot={{ r: 6, strokeWidth: 0 }}
                                />

                                {/* PV Production Line */}
                                <Line
                                    type="monotone"
                                    dataKey="pvPower"
                                    name="Sản lượng PV"
                                    stroke="#22c55e" // Green
                                    strokeWidth={3}
                                    dot={false}
                                    activeDot={{ r: 6, strokeWidth: 0 }}
                                />
                            </>
                        )}
                    </LineChart>
                </ResponsiveContainer>
            </div>
        </Card>
    );
};
