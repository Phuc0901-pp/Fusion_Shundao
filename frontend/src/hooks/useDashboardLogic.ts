import { useQuery, keepPreviousData } from '@tanstack/react-query';
import { fetchDashboardData } from '../services/dashboardService';
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
    const { data, isLoading, isFetching, isError, error, refetch } = useQuery({
        queryKey: ['dashboardData'],
        queryFn: fetchDashboardData,
        refetchInterval: 10000, // Fast update: every 10 seconds
        staleTime: 5000, // Data is fresh for 5 seconds
        placeholderData: keepPreviousData,
        retry: 3,
    });

    const {
        alerts: serverAlerts = [],
        sites = [],
        kpi,
        sensors = [],
        meters = [],
        productionData = []
    } = data || {};

    // Smart alerts: combines server alerts + auto-generated inverter/time alerts with TTS
    const smartAlerts = useSmartAlerts(sites, serverAlerts);

    return {
        data,
        isLoading,
        isFetching,
        isError,
        error: error as Error | null,
        refetch,
        serverAlerts,
        smartAlerts,
        sites,
        kpi,
        sensors,
        meters,
        productionData,
    };
};
