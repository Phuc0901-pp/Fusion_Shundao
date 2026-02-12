import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardHeader, CardTitle } from '../ui/Card';
import { Skeleton } from '../ui/Skeleton';
import { LineChart as LineChartIcon, BarChart2 } from 'lucide-react';
import api from '../../services/api';
import { ChartControls, type ViewMode } from './ChartControls';
import { DailyLineChart } from './DailyLineChart';
import { MonthlyBarChart } from './MonthlyBarChart';
import type { ProductionDataPoint } from '../../types';

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

const ProductionChartComponent: React.FC<ProductionChartProps> = ({ data, loading = false }) => {
    const [viewMode, setViewMode] = useState<ViewMode>('day');
    const [visibleSites, setVisibleSites] = useState({ site1: true, site2: true });

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
    const filteredDailyData = data.filter(filterTimeRange);

    const currentLoading = viewMode === 'day' ? loading : monthlyLoading;

    // Calculate Totals based on visible sites (using GridFeedIn as Power)
    const validDailyPowerPoints = data.filter(d =>
        (d.site1GridFeedIn != null && d.site1GridFeedIn > 0) || (d.site2GridFeedIn != null && d.site2GridFeedIn > 0)
    );

    const totalS1MWh = viewMode === 'day'
        ? (validDailyPowerPoints.reduce((sum, d) => sum + (d.site1GridFeedIn || 0), 0) * (5 / 60)) / 1000
        : (monthlyData || []).reduce((sum, d) => sum + (d.site1MaxPower || 0), 0) / 1000;

    const totalS2MWh = viewMode === 'day'
        ? (validDailyPowerPoints.reduce((sum, d) => sum + (d.site2GridFeedIn || 0), 0) * (5 / 60)) / 1000
        : (monthlyData || []).reduce((sum, d) => sum + (d.site2MaxPower || 0), 0) / 1000;

    const toggleSite = (site: 'site1' | 'site2') => {
        setVisibleSites(prev => ({ ...prev, [site]: !prev[site] }));
    };

    if (currentLoading) {
        return (
            <div className="h-[550px] bg-white rounded-2xl border border-slate-200 p-4">
                <Skeleton className="h-10 w-full mb-4" />
                <Skeleton className="h-[480px] w-full" />
            </div>
        );
    }

    return (
        <Card className="h-[600px] flex flex-col">
            <CardHeader className="mb-0 pb-4 border-b border-slate-100/50">
                <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                    <div className="flex items-center gap-4">
                        <div className="flex items-center gap-3">
                            <div className="p-2 rounded-xl bg-slate-100/80 text-slate-600">
                                {viewMode === 'day' ? <LineChartIcon className="w-5 h-5" /> : <BarChart2 className="w-5 h-5" />}
                            </div>
                            <div>
                                <CardTitle className="text-lg">Biểu đồ Tổng hợp</CardTitle>
                                <div className="flex gap-4 text-xs mt-1">
                                    {visibleSites.site1 && (
                                        <div className="flex items-baseline gap-1">
                                            <span className="text-[14px] font-bold text-blue-400">Tổng công suất Shundao1:</span>
                                            <span className="text-[16px] font-semibold text-black">{totalS1MWh.toFixed(2)} MWh</span>
                                        </div>
                                    )}
                                    {visibleSites.site2 && (
                                        <div className="flex items-baseline gap-1">
                                            <span className="text-[14px] font-bold text-blue-400">Tổng công suất Shundao2:</span>
                                            <span className="text-[16px] font-semibold text-black">{totalS2MWh.toFixed(2)} MWh</span>
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>
                    </div>

                    <ChartControls
                        viewMode={viewMode}
                        onViewModeChange={setViewMode}
                        visibleSites={visibleSites}
                        onToggleSite={toggleSite}
                    />
                </div>
            </CardHeader>

            <div className="flex-grow w-full px-2 py-4 min-h-0">
                {viewMode === 'day'
                    ? <DailyLineChart
                        data={filteredDailyData}
                        visibleSites={visibleSites}
                    />
                    : <MonthlyBarChart
                        data={monthlyData || []}
                        visibleSites={visibleSites}
                    />
                }
            </div>
        </Card>
    );
};

export const ProductionChart = React.memo(ProductionChartComponent);
