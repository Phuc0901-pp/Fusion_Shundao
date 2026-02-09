import React from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { Card, CardHeader, CardTitle } from '../ui/Card';
import { Skeleton } from '../ui/Skeleton';
import { BarChart3 } from 'lucide-react';

interface ProductionDataPoint {
    date: string;
    site1DailyEnergy: number;
    site1GridFeedIn: number;
    site1Irradiation: number;
    site2DailyEnergy: number;
    site2GridFeedIn: number;
    site2Irradiation: number;
}

interface ProductionChartProps {
    data: ProductionDataPoint[];
    loading?: boolean;
}

interface SingleSiteChartProps {
    siteName: string;
    data: { date: string; dailyEnergy: number; gridFeedIn: number; irradiation: number }[];
    color: string;
    loading?: boolean;
}

const SingleSiteChart: React.FC<SingleSiteChartProps> = ({ siteName, data, color, loading }) => {
    const totalEnergy = data.reduce((sum, d) => sum + d.dailyEnergy, 0) / 1000;

    const CustomTooltip = ({ active, payload, label }: any) => {
        if (active && payload && payload.length) {
            const d = payload[0]?.payload;
            return (
                <div className="bg-white p-3 border border-slate-200 rounded-lg shadow-xl text-xs min-w-[180px]">
                    <p className="text-slate-500 font-medium mb-2">Ngày {label}</p>
                    <div className="space-y-1">
                        <div className="flex justify-between">
                            <span>Đã tiêu thụ</span>
                            <span className="font-mono font-medium">{d?.dailyEnergy?.toLocaleString()} kWh</span>
                        </div>

                        <div className="flex justify-between">
                            <span>Bức xạ</span>
                            <span className="font-mono font-medium">{d?.irradiation?.toFixed(2)} MJ/m²</span>
                        </div>
                    </div>
                </div>
            );
        }
        return null;
    };

    if (loading) {
        return (
            <Card className="h-[380px]">
                <div className="mb-4 space-y-2">
                    <Skeleton className="h-5 w-32" />
                    <Skeleton className="h-4 w-24" />
                </div>
                <div className="h-[280px] w-full flex items-end gap-1">
                    {Array.from({ length: 10 }).map((_, i) => (
                        <Skeleton key={i} className="w-full" style={{ height: `${Math.random() * 60 + 20}%` }} />
                    ))}
                </div>
            </Card>
        );
    }

    return (
        <Card className="h-[380px]">
            <CardHeader className="mb-3 pb-0">
                <div className="flex items-center gap-2">
                    <div className={`p-1.5 rounded-lg`} style={{ backgroundColor: `${color}15` }}>
                        <BarChart3 className="w-4 h-4" style={{ color }} />
                    </div>
                    <div>
                        <CardTitle className="text-base">{siteName}</CardTitle>
                        <p className="text-slate-500 text-xs">
                            Tổng: <span className="font-semibold text-slate-700">{totalEnergy.toFixed(2)} MWh</span>
                        </p>
                    </div>
                </div>
            </CardHeader>

            <div className="h-[280px] w-full px-2" style={{ minWidth: '100px', minHeight: '280px' }}>
                <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={data} barGap={0} barCategoryGap="20%">
                        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" vertical={false} />
                        <XAxis dataKey="date" stroke="#94a3b8" tick={{ fontSize: 10 }} tickMargin={6} />
                        <YAxis yAxisId="left" stroke="#94a3b8" tick={{ fontSize: 10 }} width={55} />
                        <YAxis yAxisId="right" orientation="right" stroke="#f97316" tick={{ fontSize: 10 }} width={45} />
                        <Tooltip content={<CustomTooltip />} />
                        <Legend wrapperStyle={{ fontSize: '11px', paddingTop: '5px' }} />

                        <Bar yAxisId="left" dataKey="dailyEnergy" name="Tiêu thụ" fill={color} radius={[2, 2, 0, 0]} />

                        <Bar yAxisId="right" dataKey="irradiation" name="Bức xạ" fill="#fb923c" radius={[2, 2, 0, 0]} />
                    </BarChart>
                </ResponsiveContainer>
            </div>
        </Card>
    );
};

export const ProductionChart: React.FC<ProductionChartProps> = ({ data, loading = false }) => {
    // Transform data for each site
    const site1Data = data.map(d => ({
        date: d.date,
        dailyEnergy: d.site1DailyEnergy,
        gridFeedIn: d.site1GridFeedIn,
        irradiation: d.site1Irradiation
    }));

    const site2Data = data.map(d => ({
        date: d.date,
        dailyEnergy: d.site2DailyEnergy,
        gridFeedIn: d.site2GridFeedIn,
        irradiation: d.site2Irradiation
    }));

    return (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <SingleSiteChart
                siteName="SHUNDAO 1"
                data={site1Data}
                color="#22c55e"
                loading={loading}
            />
            <SingleSiteChart
                siteName="SHUNDAO 2"
                data={site2Data}
                color="#0ea5e9"
                loading={loading}
            />
        </div>
    );
};
