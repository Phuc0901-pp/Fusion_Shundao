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
    name: string;
    deviceStatus: string;
    strings: StringData[];
}

export interface SmartLogger {
    id: string;
    name: string;
    inverters: Inverter[];
}

export interface Site {
    id: string;
    name: string;
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
