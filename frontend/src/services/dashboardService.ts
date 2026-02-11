import api from './api';
import type { AlertMessage, Site } from '../types';

// Extended alert with device reference
export interface DeviceAlert extends AlertMessage {
    deviceId?: string;
    deviceType?: 'inverter' | 'sensor' | 'meter' | 'system';
}

// DeviceAlert interface exported for types if needed elsewhere, or remove if not used
export interface DeviceAlert extends AlertMessage {
    deviceId?: string;
    deviceType?: 'inverter' | 'sensor' | 'meter' | 'system';
}

// Fetch real data from Go backend
export const fetchDashboardData = async () => {
    try {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const data: any = await api.get('/dashboard');

        if (!data.siteData) {
            data.siteData = { all: [], siteA: [], siteB: [] };
        }
        if (!data.chartData) data.chartData = [];

        // Clean up data: Remove duplicate loggers and empty stations
        if (data.sites) {
            data.sites.forEach((site: Site) => {
                if (site.loggers) {
                    site.loggers = site.loggers.filter(logger => {
                        // Remove loggers with 0 inverters
                        if (!logger.inverters || logger.inverters.length === 0) return false;
                        return true;
                    });
                }
            });

            // Optional: Remove sites with no loggers if that contributes to empty "Station"
            data.sites = data.sites.filter((site: Site) => site.loggers && site.loggers.length > 0);
        }

        // Backend now handles alert generation, but we can augment if needed
        // For now, return as is
        return data;
    } catch (error) {
        console.error("Failed to fetch dashboard data:", error);
        return {
            alerts: [{
                id: 'error-fetch',
                timestamp: Date.now(),
                level: 'error' as const,
                message: 'Không thể kết nối đến server',
                source: 'API',
                deviceType: 'system' as const
            }],
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
