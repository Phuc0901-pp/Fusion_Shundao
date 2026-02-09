

import api from './api';

// Fetch real data from Go backend
export const generateMockData = async () => {
    try {
        const data: any = await api.get('/dashboard');
        // Ensure data structure matches what UI expects
        if (!data.siteData) {
            data.siteData = { all: [], siteA: [], siteB: [] };
        }
        if (!data.chartData) data.chartData = [];
        return data;
    } catch (error) {
        console.error("Failed to fetch dashboard data:", error);
        // Return empty structure to prevent UI crash
        return {
            alerts: [],
            sites: [],
            chartData: [],
            sensors: [],
            meters: [],
            siteData: { all: [], siteA: [], siteB: [] },
            productionData: [],
            kpi: {
                dailyEnergy: 0,
                dailyIncome: 0,
                totalEnergy: 0,
                ratedPower: 0,
                gridSupplyToday: 0,
                standardCoalSaved: 0,
                co2Reduction: 0,
                treesPlanted: 0
            }
        };
    }
};
