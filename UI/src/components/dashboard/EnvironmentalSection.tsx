import React from 'react';
import { Leaf } from 'lucide-react';
import { MetricCard } from '../widgets/MetricCard';

interface KPI {
    standardCoalSaved: number;
    co2Reduction: number;
    treesPlanted: number;
}

interface EnvironmentalSectionProps {
    kpi?: KPI;
    isLoading: boolean;
}

export const EnvironmentalSection: React.FC<EnvironmentalSectionProps> = ({ kpi, isLoading }) => {
    return (
        <div className="space-y-4">
            <h3 className="text-lg font-semibold text-slate-700">Lợi Ích Môi Trường</h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <MetricCard
                    title="Tiết kiệm than chuẩn"
                    value={kpi?.standardCoalSaved || "0"}
                    unit="Tấn"
                    icon={Leaf}
                    color="slate"
                    loading={isLoading}
                />
                <MetricCard
                    title="Giảm thải CO2"
                    value={kpi?.co2Reduction || "0"}
                    unit="Tấn"
                    icon={Leaf}
                    color="emerald"
                    loading={isLoading}
                />
                <MetricCard
                    title="Cây trồng tương đương"
                    value={kpi?.treesPlanted || "0"}
                    unit="Cây"
                    icon={Leaf}
                    color="green"
                    loading={isLoading}
                />
            </div>
        </div>
    );
};
