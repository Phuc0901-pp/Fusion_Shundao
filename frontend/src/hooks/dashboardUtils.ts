import type { DashboardResponse, Site } from '../types';

/**
 * Shared helper: strips empty loggers/sites from raw API response.
 * Used by both dashboardService.ts and useSSEDashboard.ts.
 */
export const cleanDashboardData = (data: DashboardResponse): DashboardResponse => {
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
