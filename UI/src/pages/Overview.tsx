import { RefreshCw } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { PowerChart } from '../components/charts/PowerChart';
import { AlertBox } from '../components/widgets/AlertBox';
import { StringDiagram } from '../components/strings/StringDiagram';
import { Badge } from '../components/ui/Badge';
import { generateMockData } from '../services/dashboardService';
import { ProductionSection } from '../components/dashboard/ProductionSection';
import { EnvironmentalSection } from '../components/dashboard/EnvironmentalSection';
import { DeviceSection } from '../components/dashboard/DeviceSection';

export const Overview = () => {
    const { data, isLoading, isFetching, refetch } = useQuery({
        queryKey: ['dashboardData'],
        queryFn: generateMockData,
        refetchInterval: 30000, // Auto refresh every 30s
    });

    // Safe defaults if data is undefined during loading
    const { alerts = [], sites = [], chartData = [], kpi, sensors = [], meters = [] } = data || {};

    return (
        <div className="space-y-6 animate-fade-in pb-10">
            <div className="flex justify-between items-center">
                <h2 className="text-2xl font-bold text-slate-900">Tổng Quan Hệ Thống</h2>
                <div className="flex items-center gap-3">
                    {isFetching && <Badge variant="warning" className="animate-pulse">Đang cập nhật...</Badge>}
                    <button
                        onClick={() => refetch()}
                        className="p-2 bg-white hover:bg-slate-100 rounded-lg text-slate-400 hover:text-slate-900 border border-slate-200 transition-colors shadow-sm"
                    >
                        <RefreshCw size={18} className={isFetching ? "animate-spin" : ""} />
                    </button>
                </div>
            </div>

            {/* Section 1: Production & Financial */}
            <ProductionSection kpi={kpi} isLoading={isLoading} />

            {/* Section 2: Environmental Benefits */}
            <EnvironmentalSection kpi={kpi} isLoading={isLoading} />

            {/* Row 2: Charts & Alerts */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="lg:col-span-2">
                    <PowerChart data={chartData} siteData={data?.siteData} loading={isLoading} />
                </div>
                <div>
                    <AlertBox alerts={alerts} loading={isLoading} />
                </div>
            </div>

            {/* Row 3: Sensors & Meters */}
            <DeviceSection sensors={sensors} meters={meters} loading={isLoading} />

            {/* Row 4: Detail String View */}
            <div>
                <StringDiagram
                    sites={sites}
                    loading={isLoading}
                />
            </div>
        </div>
    );
};

