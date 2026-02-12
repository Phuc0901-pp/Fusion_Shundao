import api from './api';
import type { DashboardResponse, Site } from '../types';

/**
 * Cleans raw dashboard data by removing empty loggers and sites.
 * Returns a new object without mutating the original.
 */
const cleanDashboardData = (data: DashboardResponse): DashboardResponse => {
    const cleanedSites = (data.sites ?? [])
        .map((site: Site) => ({
            ...site,
            loggers: (site.loggers ?? []).filter(
                logger => logger.inverters && logger.inverters.length > 0
            ),
        }))
        .filter(site => site.loggers.length > 0);

    return {
        ...data,
        sites: cleanedSites,
        siteData: data.siteData ?? { all: [], siteA: [], siteB: [] },
        chartData: data.chartData ?? [],
        alerts: data.alerts ?? [],
        sensors: data.sensors ?? [],
        meters: data.meters ?? [],
        productionData: data.productionData ?? [],
    };
};

/**
 * Fetches dashboard data from the Go backend.
 * Errors are propagated to React Query for proper isError handling.
 */
export const fetchDashboardData = async (): Promise<DashboardResponse> => {
    const data = await api.get('/dashboard') as unknown as DashboardResponse;
    return cleanDashboardData(data);
};
