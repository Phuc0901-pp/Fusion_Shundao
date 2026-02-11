import React, { useState } from 'react';
import { Leaf, Trees, Wind, type LucideIcon } from 'lucide-react';
import { MetricCard } from '../widgets/MetricCard';
import { MetricDetailModal } from '../modals/MetricDetailModal';
import type { Site, KPI } from '../../types';
import { formatCompactNumber } from '../../utils/formatters';

interface EnvironmentalSectionProps {
    kpi?: KPI;
    sites?: Site[];
    isLoading: boolean;
}

export const EnvironmentalSection: React.FC<EnvironmentalSectionProps> = ({ kpi, sites, isLoading }) => {
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
        { title: "Tiết kiệm than chuẩn", value: kpi?.standardCoalSaved ? formatCompactNumber(kpi.standardCoalSaved) : "0", unit: "Tấn", icon: Leaf, color: "green", field: "standardCoalSaved" as keyof KPI },
        { title: "Giảm thải CO2", value: kpi?.co2Reduction ? formatCompactNumber(kpi.co2Reduction) : "0", unit: "Tấn", icon: Wind, color: "emerald", field: "co2Reduction" as keyof KPI },
        { title: "Cây trồng tương đương", value: kpi?.treesPlanted ? formatCompactNumber(kpi.treesPlanted) : "0", unit: "Cây", icon: Trees, color: "green", field: "treesPlanted" as keyof KPI }
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
                    delay={index + 5}
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
                    totalValue={kpi[selectedMetric.field]}
                    sites={getSiteData(selectedMetric.field)}
                />
            )}
        </>
    );
};
