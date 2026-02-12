export interface DeviceData {
    timestamp: number;
    device_name: string;
    device_id: string;
    data: Record<string, number | string>;
}

export interface KPI {
    dailyEnergy: number;
    dailyIncome: number;
    totalEnergy: number;
    ratedPower: number;
    gridSupplyToday: number;
    standardCoalSaved: number;
    co2Reduction: number;
    treesPlanted: number;
}

export interface Environment {
    co2_reduction: number;
    tree_planted: number;
}

export interface StationOverview {
    station_name: string;
    kpi: KPI;
    env: Environment;
}

// For UI State
export interface AlertMessage {
    id: string;
    timestamp: number;
    level: 'info' | 'warning' | 'error' | 'success';
    message: string;
    source: string;
}

// String Diagram Interfaces
export interface StringData {
    id: string; // pv01, pv02...
    current: number;
    voltage: number;
}

export interface Inverter {
    id: string;
    dbId?: string;
    name: string;
    defaultName?: string;
    numberStringSet?: string;
    deviceStatus: string;
    strings: StringData[];
    pOutKw?: number;
    powerFactor?: number;
    gridIaA?: number;
    gridVaV?: number;
    startupTime?: string;
    insulationResistanceMO?: number;
    eDailyKwh?: number;
    qOutKvar?: number;
    gridFreqHz?: number;
    gridIbA?: number;
    gridVbV?: number;
    shutdownTime?: string;
    dcPowerKw?: number;
    eTotalKwh?: number;
    ratedPowerKw?: number;
    outputMode?: string;
    gridIcA?: number;
    gridVcV?: number;
    internalTempDegC?: number;
}

export interface SmartLogger {
    id: string;
    dbId?: string;
    name: string;
    defaultName?: string;
    inverters: Inverter[];
}

export interface Site {
    id: string;
    dbId?: string;
    name: string;
    defaultName?: string;
    loggers: SmartLogger[];
    kpi?: KPI;
}

export interface Sensor {
    id: string;
    siteId: string; // Link to site
    name: string;
    irradiance: number; // W/m2
    ambientString: number; // Temp C
    moduleTemp: number; // Temp C
    windSpeed: number; // m/s
}

export interface Meter {
    id: string;
    siteId: string; // Link to site
    name: string;
    phaseA: { voltage: number; current: number };
    phaseB: { voltage: number; current: number };
    phaseC: { voltage: number; current: number };
    totalPower: number; // kW
    frequency: number; // Hz
    powerFactor: number;
}

// Production chart data point (per site)
export interface ProductionDataPoint {
    date: string;           // "01", "02", etc.
    // SHUNDAO 1
    site1DailyEnergy: number;
    site1GridFeedIn: number;
    site1Irradiation: number;
    // SHUNDAO 2
    site2DailyEnergy: number;
    site2GridFeedIn: number;
    site2Irradiation: number;
}

// Extended alert with device info
export interface DeviceAlert extends AlertMessage {
    deviceId?: string;
    deviceType?: 'inverter' | 'sensor' | 'meter' | 'system';
}

// Aggregated API response from /dashboard
export interface DashboardResponse {
    alerts: DeviceAlert[];
    sites: Site[];
    kpi: KPI;
    sensors: Sensor[];
    meters: Meter[];
    productionData: ProductionDataPoint[];
    siteData: {
        all: ProductionDataPoint[];
        siteA: ProductionDataPoint[];
        siteB: ProductionDataPoint[];
    };
    chartData: ProductionDataPoint[];
}
