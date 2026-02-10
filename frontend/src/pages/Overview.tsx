import { RefreshCw } from 'lucide-react';
import { useQuery, keepPreviousData } from '@tanstack/react-query';
import { ProductionChart } from '../components/charts/ProductionChart';
import { AlertBox } from '../components/widgets/AlertBox';
import { StringDiagram } from '../components/strings/StringDiagram';
import { Badge } from '../components/ui/Badge';
import { fetchDashboardData } from '../services/dashboardService';
import { ProductionSection } from '../components/dashboard/ProductionSection';
import { EnvironmentalSection } from '../components/dashboard/EnvironmentalSection';
import { DeviceSection } from '../components/dashboard/DeviceSection';
import { DashboardSection } from '../components/dashboard/DashboardSection';
import { ErrorBoundary } from '../components/ErrorBoundary';
import { Zap } from 'lucide-react';

export const Overview = () => {
    const { data, isLoading, isFetching, refetch } = useQuery({
        queryKey: ['dashboardData'],
        queryFn: fetchDashboardData,
        refetchInterval: 10000, // Fast update: every 10 seconds
        placeholderData: keepPreviousData,
    });

    const { alerts = [], sites = [], kpi, sensors = [], meters = [], productionData = [] } = data || {};

    return (
        <div className="space-y-6 animate-fade-in pb-10">
            {/* Header */}
            <div className="flex justify-between items-center">
                <h2 className="text-2xl font-bold text-slate-800">Tổng Quan Hệ Thống</h2>
                <div className="flex items-center gap-3">
                    {isFetching && <Badge variant="warning" className="animate-pulse">Đang cập nhật...</Badge>}
                    <button
                        onClick={() => refetch()}
                        className="p-2 bg-white hover:bg-slate-100 rounded-lg text-slate-400 hover:text-slate-800 border border-slate-200 transition-colors shadow-sm"
                    >
                        <RefreshCw size={18} className={isFetching ? "animate-spin" : ""} />
                    </button>
                </div>
            </div>

            {/* Section 1: ALL Metrics in ONE row */}
            <ErrorBoundary sectionName="Hiệu Quả Vận Hành">
                <DashboardSection title="Hiệu Quả Vận Hành" icon={Zap}>
                    <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-3">
                        <ProductionSection kpi={kpi} sites={sites} isLoading={isLoading} />
                        <EnvironmentalSection kpi={kpi} sites={sites} isLoading={isLoading} />
                    </div>
                </DashboardSection>
            </ErrorBoundary>

            {/* Row 2: Charts (LEFT 3/4) + Alerts (RIGHT 1/4) */}
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
                {/* LEFT: Production Charts */}
                <div className="lg:col-span-3">
                    <ErrorBoundary sectionName="Biểu đồ sản lượng">
                        <ProductionChart data={productionData} loading={isLoading} />
                    </ErrorBoundary>
                </div>

                {/* RIGHT: Alert Log */}
                <div className="lg:col-span-1">
                    <ErrorBoundary sectionName="Cảnh báo">
                        <AlertBox alerts={alerts} sites={sites} loading={isLoading} />
                    </ErrorBoundary>
                </div>
            </div>

            {/* Row 3: Sensors & Meters */}
            <ErrorBoundary sectionName="Thiết bị">
                <DeviceSection sensors={sensors} meters={meters} loading={isLoading} />
            </ErrorBoundary>

            {/* Row 4: Detail String View */}
            <ErrorBoundary sectionName="Sơ đồ String">
                <StringDiagram sites={sites} loading={isLoading} />
            </ErrorBoundary>
        </div>
    );
};
