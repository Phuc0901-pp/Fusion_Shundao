import React, { useState } from 'react';
import { Zap, Activity, DollarSign, RefreshCw } from 'lucide-react';
import { MetricCard } from '../widgets/MetricCard';
import { MetricDetailModal } from '../modals/MetricDetailModal';
import type { Site, KPI } from '../../types';

interface ProductionSectionProps {
    kpi?: KPI;
    sites?: Site[];
    isLoading: boolean;
}

export const ProductionSection: React.FC<ProductionSectionProps> = ({ kpi, sites, isLoading }) => {
    const [selectedMetric, setSelectedMetric] = useState<{
        title: string;
        unit: string;
        field: keyof KPI;
        icon: any;
        color: string;
    } | null>(null);

    const handleCardClick = (title: string, unit: string, field: keyof KPI, icon: any, color: string) => {
        setSelectedMetric({ title, unit, field, icon, color });
    };

    const getSiteData = (field: keyof KPI) => {
        return sites?.map(s => ({
            id: s.id,
            name: s.name,
            value: s.kpi ? s.kpi[field] : 0
        })) || [];
    };

    return (
        <>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
                <MetricCard
                    title="Sản lượng hôm nay"
                    value={kpi?.dailyEnergy ? kpi.dailyEnergy.toLocaleString('en-US', { maximumFractionDigits: 3 }) : "0"}
                    unit="kWh"
                    icon={Zap}
                    color="solar"
                    loading={isLoading}
                    variant="flat"
                    onClick={() => handleCardClick("Sản lượng hôm nay", "kWh", "dailyEnergy", Zap, "solar")}
                />
                <MetricCard
                    title="Doanh thu hôm nay"
                    value={kpi?.dailyIncome ? kpi.dailyIncome.toLocaleString('en-US', { maximumFractionDigits: 3 }) : "0"}
                    unit="VND"
                    icon={DollarSign}
                    color="green"
                    loading={isLoading}
                    variant="flat"
                    onClick={() => handleCardClick("Doanh thu hôm nay", "VND", "dailyIncome", DollarSign, "green")}
                />
                <MetricCard
                    title="Tổng sản lượng"
                    value={kpi?.totalEnergy ? (kpi.totalEnergy / 1000).toLocaleString('en-US', { maximumFractionDigits: 3 }) : "0"}
                    unit="MWh"
                    icon={Activity}
                    color="blue"
                    loading={isLoading}
                    variant="flat"
                    onClick={() => handleCardClick("Tổng sản lượng", "MWh", "totalEnergy", Activity, "blue")}
                />
                <MetricCard
                    title="Công suất định mức"
                    value={kpi?.ratedPower ? kpi.ratedPower.toLocaleString('en-US', { maximumFractionDigits: 3 }) : "0"}
                    unit="kW"
                    icon={Zap}
                    color="slate"
                    loading={isLoading}
                    variant="flat"
                    onClick={() => handleCardClick("Công suất định mức", "kW", "ratedPower", Zap, "slate")}
                />
                <MetricCard
                    title="Từ lưới điện hôm nay"
                    value={kpi?.gridSupplyToday ? kpi.gridSupplyToday.toLocaleString('en-US', { maximumFractionDigits: 3 }) : "0"}
                    unit="kWh"
                    icon={RefreshCw}
                    color="orange"
                    loading={isLoading}
                    variant="flat"
                    onClick={() => handleCardClick("Từ lưới điện hôm nay", "kWh", "gridSupplyToday", RefreshCw, "orange")}
                />
            </div>

            {selectedMetric && kpi && (
                <MetricDetailModal
                    isOpen={!!selectedMetric}
                    onClose={() => setSelectedMetric(null)}
                    title={selectedMetric.title}
                    unit={selectedMetric.unit}
                    icon={selectedMetric.icon}
                    color={selectedMetric.color}
                    totalValue={selectedMetric.field === 'totalEnergy' ? (kpi.totalEnergy / 1000) : kpi[selectedMetric.field]}
                    sites={getSiteData(selectedMetric.field)}
                />
            )}
        </>
    );
};
