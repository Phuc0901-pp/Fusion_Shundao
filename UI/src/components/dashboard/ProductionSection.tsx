import React from 'react';
import { Zap, Activity, DollarSign, RefreshCw } from 'lucide-react';
import { MetricCard } from '../widgets/MetricCard';

interface KPI {
    dailyEnergy: number;
    dailyIncome: number;
    totalEnergy: number;
    ratedPower: number;
    gridSupplyToday: number;
    standardCoalSaved: number;
    co2Reduction: number;
    treesPlanted: number;
}

interface ProductionSectionProps {
    kpi?: KPI;
    isLoading: boolean;
}

export const ProductionSection: React.FC<ProductionSectionProps> = ({ kpi, isLoading }) => {
    return (
        <div className="space-y-4">
            <h3 className="text-lg font-semibold text-slate-700">Sản Xuất & Tài Chính</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
                <MetricCard
                    title="Sản lượng hôm nay"
                    value={kpi?.dailyEnergy ? kpi.dailyEnergy.toLocaleString() : "0"}
                    unit="kWh"
                    icon={Zap}
                    trend={12}
                    color="solar"
                    loading={isLoading}
                />
                <MetricCard
                    title="Doanh thu hôm nay"
                    value={`$${kpi?.dailyIncome || "0"}`}
                    unit="USD"
                    icon={DollarSign}
                    trend={8}
                    color="green"
                    loading={isLoading}
                />
                <MetricCard
                    title="Tổng sản lượng"
                    value={kpi?.totalEnergy ? (kpi.totalEnergy / 1000).toFixed(2) : "0"}
                    unit="MWh"
                    icon={Activity}
                    color="blue"
                    loading={isLoading}
                />
                <MetricCard
                    title="Công suất định mức"
                    value={kpi?.ratedPower || "0"}
                    unit="kW"
                    icon={Zap}
                    color="slate"
                    loading={isLoading}
                />
                <MetricCard
                    title="Từ lưới điện hôm nay"
                    value={kpi?.gridSupplyToday || "0"}
                    unit="kWh"
                    icon={RefreshCw}
                    color="orange"
                    loading={isLoading}
                />
            </div>
        </div>
    );
};
