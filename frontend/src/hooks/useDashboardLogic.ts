import { useSSEDashboard } from './useSSEDashboard';
import { useSmartAlerts } from './useSmartAlerts';
import type { DashboardResponse, DeviceAlert, Site, KPI, Sensor, Meter, ProductionDataPoint } from '../types';

interface DashboardLogicResult {
    data: DashboardResponse | undefined;
    isLoading: boolean;
    isFetching: boolean;
    isError: boolean;
    error: Error | null;
    refetch: () => void;
    // Derived data
    serverAlerts: DeviceAlert[];
    smartAlerts: DeviceAlert[];
    sites: Site[];
    kpi: KPI | undefined;
    sensors: Sensor[];
    meters: Meter[];
    productionData: ProductionDataPoint[];
}

export const useDashboardLogic = (): DashboardLogicResult => {
    // SSE replaces React Query polling – data is pushed by the server
    const { data, isLoading, isError, error } = useSSEDashboard();

    const {
        alerts: serverAlerts = [],
        sites = [],
        kpi,
        sensors = [],
        meters = [],
        productionData = []
    } = data || {};

    const smartAlerts = useSmartAlerts(sites, serverAlerts);

    return {
        data,
        isLoading,
        isFetching: false, // SSE has no explicit fetching state
        isError,
        error: error as Error | null,
        refetch: () => {}, // No-op: SSE is always connected
        serverAlerts,
        smartAlerts,
        sites,
        kpi,
        sensors,
        meters,
        productionData,
    };
};
