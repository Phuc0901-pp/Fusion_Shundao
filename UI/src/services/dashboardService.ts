import type { AlertMessage } from '../types';

// Mock Data Generators
export const generateMockData = async () => {
    // Simulate network delay
    await new Promise(resolve => setTimeout(resolve, 1500));

    const MOCK_ALERTS: AlertMessage[] = [
        { id: '1', timestamp: Date.now(), level: 'warning', message: 'Phát hiện hiệu suất biến tần 1 giảm', source: 'Phân tích' },
        { id: '2', timestamp: Date.now() - 500000, level: 'info', message: 'Đã tạo báo cáo hàng ngày', source: 'Hệ thống' },
        { id: '3', timestamp: Date.now() - 1000000, level: 'success', message: 'Đồng bộ sao lưu hoàn tất', source: 'Cơ sở dữ liệu' },
    ];

    // Generate Hierarchical String Data
    const generateStrings = (count: number, prefix: string) => {
        return Array.from({ length: count }).map((_, i) => ({
            id: `${prefix}-Chuỗi${String(i + 1).padStart(2, '0')}`,
            current: Math.random() > 0.1 ? 8 + Math.random() * 4 : 0,
            voltage: Math.random() > 0.1 ? 600 + Math.random() * 50 : 0,
        }));
    };

    const generateHierarchy = () => {
        const sites = [
            { id: 'site1', name: 'Dự án Shundao 2', loggers: 2 }, // Shundao 2
            { id: 'site2', name: 'Dự án Shundao 1', loggers: 1 }  // Shundao 1
        ];

        return sites.map(site => ({
            id: site.id,
            name: site.name,
            loggers: Array.from({ length: site.loggers }).map((_, i) => ({
                id: `${site.id}-logger-${i + 1}`,
                name: `Bộ ghi ${i + 1}`,
                inverters: Array.from({ length: 3 }).map((__, j) => ({ // 3 inverters per logger
                    id: `${site.id}-logger${i + 1}-inv${j + 1}`,
                    name: `Biến tần ${j + 1}`,
                    strings: generateStrings(16, `PV`)
                }))
            }))
        }));
    };

    const MOCK_SITES = generateHierarchy(); // Structured Data

    // Helper to generate a day profile
    const generateDayProfile = (peak: number, noise: number) => {
        return Array.from({ length: 24 }).map((_, i) => {
            const hour = i;
            let val = 0;
            // Solar curve roughly 6am to 6pm
            if (hour > 5 && hour < 19) {
                val = Math.sin((hour - 6) / 13 * Math.PI) * peak;
            }
            return Math.max(0, val + (Math.random() - 0.5) * noise);
        });
    };

    const siteAPV = generateDayProfile(600, 50);
    const siteAConsume = generateDayProfile(400, 30).map(v => v + 50); // Base load

    const siteBPV = generateDayProfile(450, 40);
    const siteBConsume = generateDayProfile(300, 20).map(v => v + 30);

    const generateTimeData = (pvArr: number[], consumeArr: number[]) => {
        return pvArr.map((pv, i) => ({
            time: `${i}:00`,
            power: pv, // Generic power field
            pvPower: pv,
            gridPower: Math.max(0, consumeArr[i] - pv), // Grid supplements if PV < Consume (simplistic)
            consumptionPower: consumeArr[i]
        }));
    };

    const siteAData = generateTimeData(siteAPV, siteAConsume);
    const siteBData = generateTimeData(siteBPV, siteBConsume);

    // Aggregate All
    const allData = siteAData.map((d, i) => ({
        time: d.time,
        power: d.pvPower + siteBData[i].pvPower,
        pvPower: d.pvPower + siteBData[i].pvPower,
        gridPower: d.gridPower + siteBData[i].gridPower,
        consumptionPower: d.consumptionPower + siteBData[i].consumptionPower,
        // Add individual site data for multi-site charting
        siteAPower: d.pvPower,
        siteBPower: siteBData[i].pvPower
    }));

    // Mock Sensors & Meters
    const MOCK_SENSORS = [
        {
            id: 'sensor-1',
            siteId: 'site-a',
            name: 'Trạm Quan Trắc - Nhà Máy 1',
            irradiance: 850 + Math.random() * 50,
            ambientString: 32 + Math.random() * 2,
            moduleTemp: 45 + Math.random() * 5,
            windSpeed: 3.5 + Math.random()
        },
        {
            id: 'sensor-2',
            siteId: 'site-b',
            name: 'Trạm Quan Trắc - Nhà Máy 2',
            irradiance: 920 + Math.random() * 60,
            ambientString: 30 + Math.random() * 2,
            moduleTemp: 42 + Math.random() * 5,
            windSpeed: 4.2 + Math.random()
        }
    ];

    const MOCK_METERS = [
        {
            id: 'meter-1',
            siteId: 'site-a',
            name: 'Đồng Hồ Tổng - NM1',
            phaseA: { voltage: 220 + Math.random() * 5, current: 15 + Math.random() * 2 },
            phaseB: { voltage: 221 + Math.random() * 5, current: 14 + Math.random() * 2 },
            phaseC: { voltage: 219 + Math.random() * 5, current: 16 + Math.random() * 2 },
            totalPower: 125.5,
            frequency: 50.0,
            powerFactor: 0.98
        },
        {
            id: 'meter-2',
            siteId: 'site-b',
            name: 'Đồng Hồ Tổng - NM2',
            phaseA: { voltage: 222 + Math.random() * 5, current: 20 + Math.random() * 3 },
            phaseB: { voltage: 220 + Math.random() * 5, current: 21 + Math.random() * 3 },
            phaseC: { voltage: 221 + Math.random() * 5, current: 19 + Math.random() * 3 },
            totalPower: 145.2,
            frequency: 50.1,
            powerFactor: 0.99
        }
    ];

    return {
        alerts: MOCK_ALERTS,
        sites: MOCK_SITES,
        chartData: allData,
        sensors: MOCK_SENSORS,
        meters: MOCK_METERS,
        siteData: {
            all: allData,
            siteA: siteAData,
            siteB: siteBData
        },
        kpi: {
            // Production & Financial
            dailyEnergy: 2450.5,      // Output Today (kWh)
            dailyIncome: 320.5,       // Revenue Today ($)
            totalEnergy: 154200.8,    // Total Energy (kWh)
            ratedPower: 600.0,        // Rated Power (kW)
            gridSupplyToday: 45.2,    // From Grid Today (kWh)

            // Environmental
            standardCoalSaved: 850.4, // Standard Coal (Tons)
            co2Reduction: 1.2,        // CO2 Reduced (Tons)
            treesPlanted: 150         // Equivalent Trees
        }
    };
};
