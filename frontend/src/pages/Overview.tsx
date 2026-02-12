import { RefreshCw, Zap } from 'lucide-react';
import { ProductionChart } from '../components/charts/ProductionChart';
import { AlertBox } from '../components/widgets/AlertBox';
import { StringDiagram } from '../components/strings/StringDiagram';
import { Badge } from '../components/ui/Badge';
import { ProductionSection } from '../components/dashboard/ProductionSection';
import { DeviceSection } from '../components/dashboard/DeviceSection';
import { DashboardSection } from '../components/dashboard/DashboardSection';
import { ErrorBoundary } from '../components/common/ErrorBoundary';
import { WelcomeOverlay } from '../components/common/WelcomeOverlay';
import { LoadingScreen } from '../components/common/LoadingScreen';
import { useDashboardLogic } from '../hooks/useDashboardLogic';

export const Overview = () => {
    const {
        data,
        isLoading,
        isFetching,
        isError,
        error,
        refetch,
        smartAlerts,
        sites,
        kpi,
        sensors,
        meters,
        productionData
    } = useDashboardLogic();

    if (isError) {
        return (
            <div className="flex flex-col items-center justify-center h-[80vh] space-y-4 animate-fade-in">
                <div className="p-4 bg-red-50 rounded-full">
                    <Zap size={48} className="text-red-500" />
                </div>
                <h2 className="text-xl font-bold text-slate-800">Không thể kết nối đến máy chủ</h2>
                <p className="text-slate-500 max-w-md text-center">
                    {error instanceof Error ? error.message : "Đã xảy ra lỗi không xác định."}
                </p>
                <button
                    onClick={() => refetch()}
                    className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors flex items-center gap-2"
                >
                    <RefreshCw size={18} />
                    Thử lại
                </button>
            </div>
        );
    }

    return (
        <div className="space-y-6 animate-fade-in pb-10">
            {isLoading && !data && <LoadingScreen />}
            <WelcomeOverlay />
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
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                        <ProductionSection kpi={kpi} sites={sites} isLoading={isLoading} />
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
                        <AlertBox alerts={smartAlerts} sites={sites} loading={isLoading} />
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
