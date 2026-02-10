import React, { useState } from 'react';
import { Zap, Activity, DollarSign, RefreshCw, type LucideIcon } from 'lucide-react';
import { MetricCard } from '../widgets/MetricCard';
import { MetricDetailModal } from '../modals/MetricDetailModal';
import type { Site, KPI } from '../../types';
import { formatCompactNumber } from '../../utils/formatters';

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
        icon: LucideIcon;
        color: string;
    } | null>(null);

    const handleCardClick = (title: string, unit: string, field: keyof KPI, icon: LucideIcon, color: string) => {
        setSelectedMetric({ title, unit, field, icon, color });
    };

    const getSiteData = (field: keyof KPI) => {
        return sites?.map(s => ({
            id: s.id,
            name: s.name,
            value: s.kpi ? s.kpi[field] : 0
        })) || [];
    };

    const metrics = [
        { title: "Sản lượng hôm nay", value: kpi?.dailyEnergy ? formatCompactNumber(kpi.dailyEnergy) : "0", unit: "kWh", icon: Zap, color: "solar", field: "dailyEnergy" as keyof KPI },
        { title: "Doanh thu hôm nay", value: kpi?.dailyIncome ? formatCompactNumber(kpi.dailyIncome) : "0", unit: "VND", icon: DollarSign, color: "green", field: "dailyIncome" as keyof KPI },
        { title: "Tổng sản lượng", value: kpi?.totalEnergy ? formatCompactNumber(kpi.totalEnergy / 1000) : "0", unit: "MWh", icon: Activity, color: "blue", field: "totalEnergy" as keyof KPI },
        { title: "Công suất định mức", value: kpi?.ratedPower ? formatCompactNumber(kpi.ratedPower) : "0", unit: "kW", icon: Zap, color: "slate", field: "ratedPower" as keyof KPI },
        { title: "Lưới điện hôm nay", value: kpi?.gridSupplyToday ? formatCompactNumber(kpi.gridSupplyToday) : "0", unit: "kWh", icon: RefreshCw, color: "orange", field: "gridSupplyToday" as keyof KPI }
    ];

    return (
        <>
            {metrics.map((metric, index) => (
                <MetricCard
                    key={metric.field}
                    title={metric.title}
                    value={metric.value}
                    unit={metric.unit}
                    icon={metric.icon}
                    color={metric.color}
                    loading={isLoading}
                    variant="flat"
                    delay={index}
                    onClick={() => handleCardClick(metric.title, metric.unit, metric.field, metric.icon, metric.color)}
                />
            ))}

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
