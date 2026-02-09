import api from './api';
import type { AlertMessage, Site } from '../types';

// Extended alert with device reference
export interface DeviceAlert extends AlertMessage {
    deviceId?: string;
    deviceType?: 'inverter' | 'sensor' | 'meter' | 'system';
}

// Generate alerts from device status
const generateDeviceAlerts = (sites: Site[]): DeviceAlert[] => {
    const alerts: DeviceAlert[] = [];
    const now = Date.now();

    sites?.forEach(site => {
        site.loggers?.forEach(logger => {
            logger.inverters?.forEach(inverter => {
                const status = inverter.deviceStatus?.toLowerCase() || '';

                // Device not connected or has error
                if (status !== 'grid connected' && status !== '') {
                    alerts.push({
                        id: `inv-${inverter.id}-${now}`,
                        timestamp: now,
                        level: status === 'fault' || status.includes('error') ? 'error' : 'warning',
                        message: `Trạng thái: ${inverter.deviceStatus || 'Không xác định'}`,
                        source: inverter.name || inverter.id,
                        deviceId: inverter.id,
                        deviceType: 'inverter'
                    });
                }

                // Check for no data (no strings or empty strings)
                if (!inverter.strings || inverter.strings.length === 0) {
                    alerts.push({
                        id: `inv-nodata-${inverter.id}-${now}`,
                        timestamp: now,
                        level: 'warning',
                        message: 'Không có dữ liệu string',
                        source: inverter.name || inverter.id,
                        deviceId: inverter.id,
                        deviceType: 'inverter'
                    });
                }

                // Check for zero power when should be producing
                const hour = new Date().getHours();
                if (hour >= 6 && hour <= 18) {
                    if (inverter.pOutKw !== undefined && inverter.pOutKw === 0 && status === 'grid connected') {
                        alerts.push({
                            id: `inv-nopower-${inverter.id}-${now}`,
                            timestamp: now,
                            level: 'warning',
                            message: 'Công suất đầu ra = 0 kW trong giờ làm việc',
                            source: inverter.name || inverter.id,
                            deviceId: inverter.id,
                            deviceType: 'inverter'
                        });
                    }
                }
            });
        });
    });

    // Sort by level (error first, then warning)
    return alerts.sort((a, b) => {
        const order = { error: 0, warning: 1, info: 2, success: 3 };
        return order[a.level] - order[b.level];
    });
};

// Fetch real data from Go backend
export const generateMockData = async () => {
    try {
        const data: any = await api.get('/dashboard');

        if (!data.siteData) {
            data.siteData = { all: [], siteA: [], siteB: [] };
        }
        if (!data.chartData) data.chartData = [];

        // Generate alerts from device status
        if (!data.alerts || data.alerts.length === 0) {
            data.alerts = generateDeviceAlerts(data.sites || []);
        } else {
            const deviceAlerts = generateDeviceAlerts(data.sites || []);
            data.alerts = [...data.alerts, ...deviceAlerts];
        }

        // Add system status alert
        const timestamp = Date.now();
        const connectedCount = data.sites?.reduce((acc: number, site: Site) => {
            return acc + site.loggers?.reduce((logAcc, log) => {
                return logAcc + log.inverters?.filter(inv =>
                    inv.deviceStatus?.toLowerCase() === 'grid connected'
                ).length || 0;
            }, 0) || 0;
        }, 0) || 0;

        const totalCount = data.sites?.reduce((acc: number, site: Site) => {
            return acc + site.loggers?.reduce((logAcc, log) => logAcc + (log.inverters?.length || 0), 0) || 0;
        }, 0) || 0;

        if (totalCount > 0) {
            data.alerts.unshift({
                id: `system-status-${timestamp}`,
                timestamp,
                level: connectedCount === totalCount ? 'success' : 'info',
                message: `${connectedCount}/${totalCount} inverter đang kết nối`,
                source: 'Hệ thống',
                deviceType: 'system'
            });
        }

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
