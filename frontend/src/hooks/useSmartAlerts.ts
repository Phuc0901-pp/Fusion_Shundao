import { useRef, useMemo } from 'react';
import type { Site, DeviceAlert } from '../types';
import { playAlarmSound, playGentleChime, playHappyChime, speakAlert } from '../utils/audio';

// ─── Constants ─────────────────────────────────────────────────────
const WORKING_HOUR_START = 6;
const WORKING_HOUR_END = 18;
const THRESHOLD_PERCENT = 0.8;

// ─── Hook ──────────────────────────────────────────────────────────
export function useSmartAlerts(sites: Site[], serverAlerts: DeviceAlert[]): DeviceAlert[] {
    const prevHourRef = useRef<number | null>(null);
    const spokenAlertsRef = useRef<Set<string>>(new Set());
    const timeEventFiredRef = useRef<Set<string>>(new Set());

    const smartAlerts = useMemo(() => {
        const now = new Date();
        const currentHour = now.getHours();
        const generatedAlerts: DeviceAlert[] = [];
        const newSpokenKeys: string[] = [];

        // ─── Time-based Events ─────────────────────────────────
        const todayKey = `${now.getFullYear()}-${now.getMonth()}-${now.getDate()}`;

        // Start of Day (6:00)
        if (currentHour >= WORKING_HOUR_START && currentHour < WORKING_HOUR_START + 1) {
            const startKey = `start-${todayKey}`;
            if (!timeEventFiredRef.current.has(startKey)) {
                timeEventFiredRef.current.add(startKey);
                generatedAlerts.push({
                    id: `smart-start-${Date.now()}`,
                    timestamp: Date.now(),
                    level: 'success',
                    message: 'Hệ thống bắt đầu hoạt động. Chào mừng ngày mới!',
                    source: 'Hệ thống',
                    deviceType: 'system',
                });
                setTimeout(() => {
                    playHappyChime();
                    setTimeout(() => speakAlert('Hệ thống bắt đầu hoạt động. Chào mừng ngày mới!'), 800);
                }, 100);
            }
        }

        // End of Day (18:00)
        if (currentHour >= WORKING_HOUR_END && currentHour < WORKING_HOUR_END + 1) {
            const endKey = `end-${todayKey}`;
            if (!timeEventFiredRef.current.has(endKey)) {
                timeEventFiredRef.current.add(endKey);
                generatedAlerts.push({
                    id: `smart-end-${Date.now()}`,
                    timestamp: Date.now(),
                    level: 'info',
                    message: 'Cảnh báo: Hệ thống Inverter hết giờ hoạt động.',
                    source: 'Hệ thống',
                    deviceType: 'system',
                });
                setTimeout(() => {
                    playGentleChime();
                    setTimeout(() => speakAlert('Cảnh báo: Hệ thống Inverter hết giờ hoạt động.'), 600);
                }, 100);
            }
        }

        // Clean up old day keys
        if (prevHourRef.current !== null && prevHourRef.current !== currentHour) {
            if (currentHour === 0) {
                timeEventFiredRef.current.clear();
            }
        }
        prevHourRef.current = currentHour;

        const isWorkingHours = currentHour >= WORKING_HOUR_START && currentHour < WORKING_HOUR_END;

        // ─── Inverter String Analysis ──────────────────────────
        if (isWorkingHours) {
            for (const site of sites) {
                for (const logger of site.loggers || []) {
                    for (const inverter of logger.inverters || []) {
                        const strings = inverter.strings || [];
                        if (strings.length === 0) continue;

                        // Calculate averages from active strings
                        const activeStrings = strings.filter(s => s.voltage > 10 && s.current > 0);
                        const avgVoltage = activeStrings.length > 0
                            ? activeStrings.reduce((sum, s) => sum + s.voltage, 0) / activeStrings.length
                            : 0;
                        const avgCurrent = activeStrings.length > 0
                            ? activeStrings.reduce((sum, s) => sum + s.current, 0) / activeStrings.length
                            : 0;

                        const thresholdV = avgVoltage * THRESHOLD_PERCENT;
                        const thresholdI = avgCurrent * THRESHOLD_PERCENT;

                        for (const str of strings) {
                            const hasVoltage = str.voltage > 10;
                            const hasCurrent = str.current > 0;
                            const alertKey = `${inverter.id}-${str.id}`;

                            // Zero output during working hours
                            if (!hasVoltage && !hasCurrent) {
                                const key = `${alertKey}-dead`;
                                if (!spokenAlertsRef.current.has(key)) {
                                    newSpokenKeys.push(key);
                                    generatedAlerts.push({
                                        id: `smart-${key}-${Date.now()}`,
                                        timestamp: Date.now(),
                                        level: 'error',
                                        message: `${inverter.name} → ${str.id}: Mất dòng & mất điện áp`,
                                        source: inverter.name,
                                        deviceType: 'inverter',
                                        deviceId: inverter.id,
                                    });
                                }
                            } else if (!hasCurrent && hasVoltage) {
                                const key = `${alertKey}-noI`;
                                if (!spokenAlertsRef.current.has(key)) {
                                    newSpokenKeys.push(key);
                                    generatedAlerts.push({
                                        id: `smart-${key}-${Date.now()}`,
                                        timestamp: Date.now(),
                                        level: 'error',
                                        message: `${inverter.name} → ${str.id}: Mất dòng (${str.current.toFixed(2)}A)`,
                                        source: inverter.name,
                                        deviceType: 'inverter',
                                        deviceId: inverter.id,
                                    });
                                }
                            } else if (!hasVoltage && hasCurrent) {
                                const key = `${alertKey}-noV`;
                                if (!spokenAlertsRef.current.has(key)) {
                                    newSpokenKeys.push(key);
                                    generatedAlerts.push({
                                        id: `smart-${key}-${Date.now()}`,
                                        timestamp: Date.now(),
                                        level: 'error',
                                        message: `${inverter.name} → ${str.id}: Mất điện áp (${str.voltage.toFixed(1)}V)`,
                                        source: inverter.name,
                                        deviceType: 'inverter',
                                        deviceId: inverter.id,
                                    });
                                }
                            } else {
                                // Threshold check
                                const warnings: string[] = [];
                                if (avgCurrent > 0 && str.current < thresholdI) {
                                    warnings.push(`${str.current.toFixed(2)}A < ${avgCurrent.toFixed(2)}A`);
                                }
                                if (avgVoltage > 0 && str.voltage < thresholdV) {
                                    warnings.push(`${str.voltage.toFixed(1)}V < ${avgVoltage.toFixed(1)}V`);
                                }

                                if (warnings.length > 0) {
                                    const key = `${alertKey}-thresh`;
                                    if (!spokenAlertsRef.current.has(key)) {
                                        newSpokenKeys.push(key);
                                        generatedAlerts.push({
                                            id: `smart-${key}-${Date.now()}`,
                                            timestamp: Date.now(),
                                            level: 'warning',
                                            message: `${inverter.name} → ${str.id}: Vượt ngưỡng (${warnings.join(' | ')})`,
                                            source: inverter.name,
                                            deviceType: 'inverter',
                                            deviceId: inverter.id,
                                        });
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }

        // ─── Play Audio for NEW critical alerts ────────────────
        const newErrors = generatedAlerts.filter(a => a.level === 'error' && newSpokenKeys.some(k => a.id.includes(k)));
        if (newErrors.length > 0) {
            // Get first fault inverter name for TTS
            const firstFault = newErrors[0];
            setTimeout(() => {
                playAlarmSound();
                setTimeout(() => {
                    speakAlert(`Báo động! Thiết bị ${firstFault.source} gặp sự cố.${newErrors.length > 1 ? ` Tổng cộng ${newErrors.length} chuỗi bất thường.` : ''}`);
                }, 700);
            }, 100);
        }

        // Mark new spoken keys
        for (const key of newSpokenKeys) {
            spokenAlertsRef.current.add(key);
        }

        // Combine: server alerts (enriched) first, then smart alerts
        const enrichedServerAlerts = serverAlerts.map(alert => {
            if (alert.deviceId) return alert;

            // Try to find inverter by source name (Smart Matching)
            for (const site of sites) {
                for (const logger of site.loggers || []) {
                    // Strategy 1: Exact Inverter Name
                    let found = logger.inverters?.find(inv => inv.name === alert.source);

                    // Strategy 2: Source contains Logger Name AND Inverter Name (e.g., "Station 21 - Inverter 7")
                    if (!found && alert.source.includes(logger.name)) {
                        found = logger.inverters?.find(inv => alert.source.includes(inv.name));
                    }

                    // Strategy 3: Source ends with Inverter Name (e.g. "... - Inverter 7")
                    // Check for " - Inverter 7" or " Inverter 7" to avoid matching "Inverter 17" with "Inverter 7"
                    if (!found) {
                        found = logger.inverters?.find(inv =>
                            alert.source.endsWith(` ${inv.name}`) ||
                            alert.source.endsWith(`-${inv.name}`)
                        );
                    }

                    if (found) {
                        return {
                            ...alert,
                            deviceId: found.id,
                            deviceType: 'inverter' as const
                        };
                    }
                }
            }
            return alert;
        });

        return [...enrichedServerAlerts, ...generatedAlerts];
    }, [sites, serverAlerts]);

    return smartAlerts;
}
