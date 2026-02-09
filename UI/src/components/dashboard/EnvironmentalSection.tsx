import React, { useState } from 'react';
import { Leaf } from 'lucide-react';
import { MetricCard } from '../widgets/MetricCard';
import { MetricDetailModal } from '../modals/MetricDetailModal';
import type { Site, KPI } from '../../types';

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
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <MetricCard
                    title="Tiết kiệm than chuẩn"
                    value={kpi?.standardCoalSaved ? kpi.standardCoalSaved.toLocaleString('en-US', { maximumFractionDigits: 3 }) : "0"}
                    unit="Tấn"
                    icon={Leaf}
                    color="green"
                    loading={isLoading}
                    variant="flat"
                    onClick={() => handleCardClick("Tiết kiệm than chuẩn", "Tấn", "standardCoalSaved", Leaf, "green")}
                />
                <MetricCard
                    title="Giảm thải CO2"
                    value={kpi?.co2Reduction ? kpi.co2Reduction.toLocaleString('en-US', { maximumFractionDigits: 3 }) : "0"}
                    unit="Tấn"
                    icon={Leaf}
                    color="green"
                    loading={isLoading}
                    variant="flat"
                    onClick={() => handleCardClick("Giảm thải CO2", "Tấn", "co2Reduction", Leaf, "green")}
                />
                <MetricCard
                    title="Cây trồng tương đương"
                    value={kpi?.treesPlanted ? kpi.treesPlanted.toLocaleString('en-US', { maximumFractionDigits: 0 }) : "0"}
                    unit="Cây"
                    icon={Leaf}
                    color="green"
                    loading={isLoading}
                    variant="flat"
                    onClick={() => handleCardClick("Cây trồng tương đương", "Cây", "treesPlanted", Leaf, "green")}
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
                    totalValue={kpi[selectedMetric.field]}
                    sites={getSiteData(selectedMetric.field)}
                />
            )}
        </>
    );
};
