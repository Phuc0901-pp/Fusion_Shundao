import React from 'react';
import { Zap, Activity } from 'lucide-react';
import { DetailedMetricCard } from '../widgets/DetailedMetricCard';
import type { Site, KPI } from '../../types';

interface ProductionSectionProps {
    kpi?: KPI;
    sites?: Site[];
    isLoading: boolean;
}

export const ProductionSection: React.FC<ProductionSectionProps> = ({ kpi, sites, isLoading }) => {
    const getSiteData = (field: keyof KPI) => {
        return sites?.map(s => ({
            id: s.id,
            name: s.name,
            value: s.kpi ? s.kpi[field] : 0
        })) || [];
    };

    const metrics = [
        {
            title: "Sản lượng hôm nay",
            unit: "MWh",
            icon: Zap,
            color: "solar",
            field: "dailyEnergy" as keyof KPI,
            totalValue: kpi?.dailyEnergy || 0
        },
        {
            title: "Tổng sản lượng",
            unit: "GWh",
            icon: Activity,
            color: "blue",
            field: "totalEnergy" as keyof KPI,
            totalValue: kpi?.totalEnergy ? kpi.totalEnergy / 1000 : 0
        },
        {
            title: "Công suất định mức",
            unit: "MW",
            icon: Zap,
            color: "slate",
            field: "ratedPower" as keyof KPI,
            totalValue: kpi?.ratedPower || 0
        }
    ];

    return (
        <>
            {metrics.map((metric, index) => (
                <DetailedMetricCard
                    key={metric.field}
                    title={metric.title}
                    unit={metric.unit}
                    icon={metric.icon}
                    color={metric.color}
                    totalValue={metric.totalValue}
                    sites={getSiteData(metric.field)}
                    loading={isLoading}
                    delay={index}
                />
            ))}
        </>
    );
};
