import { useRef, useMemo, useCallback, useEffect, useState } from 'react';
import type { Site, DeviceAlert, Inverter } from '../types';
import { playAlarmSound, playGentleChime, playHappyChime } from '../utils/audio';
import { ALERT_CONFIG, SOLAR_START_HOUR, SOLAR_END_HOUR } from '../config/constants';

// ─── Debounce TTL ──────────────────────────────────────────────────
// A fault must persist continuously for this duration before it is
// CONFIRMED (red alarm fires). Faults that self-heal within this window
// are silently discarded → eliminates cloud-shadow false positives.
// Configured via ALERT_CONFIG.debounceMs in src/config/constants.ts

// ─── Types ──────────────────────────────────────────────────────────
interface InverterEntry {
    inverter: Inverter;
    loggerName: string;
}

interface PendingFault {
    firstSeenAt: number;
    faultType: 'dead' | 'noI' | 'noV' | 'thresh';
    message: string;
    inverterId: string;
    inverterName: string;
    loggerName: string;
}

// ─── Helpers ────────────────────────────────────────────────────────
function buildInverterIndex(sites: Site[]): Map<string, InverterEntry> {
    const map = new Map<string, InverterEntry>();
    for (const site of sites) {
        for (const logger of site.loggers || []) {
            for (const inverter of logger.inverters || []) {
                map.set(inverter.id, { inverter, loggerName: logger.name });
            }
        }
    }
    return map;
}

// ─── Hook ──────────────────────────────────────────────────────────
export function useSmartAlerts(sites: Site[], serverAlerts: DeviceAlert[]): DeviceAlert[] {
    const prevHourRef = useRef<number | null>(null);
    const timeEventFiredRef = useRef<Set<string>>(new Set());

    // STATEFUL DEBOUNCE: track first-seen timestamps per fault key
    const pendingFaultsRef = useRef<Map<string, PendingFault>>(new Map());
    // Track which faults already fired an alarm (to avoid re-alarm spam)
    const confirmedFaultsRef = useRef<Set<string>>(new Set());

    const inverterIndex = useMemo(() => buildInverterIndex(sites), [sites]);

    const enrichAlert = useCallback((alert: DeviceAlert): DeviceAlert => {
        if (alert.deviceId) return alert;
        for (const [, entry] of inverterIndex) {
            const inv = entry.inverter;
            const loggerName = entry.loggerName;
            const src = alert.source;
            if (
                src === inv.name ||
                (src.includes(loggerName) && src.includes(inv.name)) ||
                src.endsWith(` ${inv.name}`) ||
                src.endsWith(`-${inv.name}`)
            ) {
                return { 
                    ...alert, 
                    deviceId: inv.id, 
                    deviceType: 'inverter' as const,
                    source: `${loggerName} - ${inv.name}`
                };
            }
        }
        return alert;
    }, [inverterIndex]);

    const [audioQueue, setAudioQueue] = useState<string[]>([]);

    const smartAlerts = useMemo(() => {
        const now = new Date();
        const currentHour = now.getHours();
        const nowMs = now.getTime();
        const generatedAlerts: DeviceAlert[] = [];
        const newAudioToPlay: string[] = [];

        // ── Time-based Events ──────────────────────────────────────────
        const todayKey = `${now.getFullYear()}-${now.getMonth()}-${now.getDate()}`;

        if (currentHour >= SOLAR_START_HOUR && currentHour < SOLAR_START_HOUR + 1) {
            const startKey = `start-${todayKey}`;
            if (!timeEventFiredRef.current.has(startKey)) {
                timeEventFiredRef.current.add(startKey);
                generatedAlerts.push({
                    id: `smart-start-${nowMs}`,
                    timestamp: nowMs,
                    level: 'success',
                    message: 'Hệ thống bắt đầu hoạt động. Chào mừng ngày mới!',
                    source: 'Hệ thống',
                    deviceType: 'system',
                });
                newAudioToPlay.push('happy');
            }
        }

        if (currentHour >= SOLAR_END_HOUR && currentHour < SOLAR_END_HOUR + 1) {
            const endKey = `end-${todayKey}`;
            if (!timeEventFiredRef.current.has(endKey)) {
                timeEventFiredRef.current.add(endKey);
                generatedAlerts.push({
                    id: `smart-end-${nowMs}`,
                    timestamp: nowMs,
                    level: 'info',
                    message: 'Cảnh báo: Hệ thống Inverter hết giờ hoạt động.',
                    source: 'Hệ thống',
                    deviceType: 'system',
                });
                newAudioToPlay.push('gentle');
            }
        }

        // Midnight state reset
        if (prevHourRef.current !== null && prevHourRef.current !== currentHour && currentHour === 0) {
            timeEventFiredRef.current.clear();
            pendingFaultsRef.current.clear();
            confirmedFaultsRef.current.clear();
        }
        prevHourRef.current = currentHour;

        const isWorkingHours = currentHour >= SOLAR_START_HOUR && currentHour < SOLAR_END_HOUR;

        if (isWorkingHours) {
            // Track which fault keys we see THIS cycle (to detect healed faults)
            // ── PASS 1: Detect faults in current data snapshot ────────────
            const currentCycleFaultKeys = new Set<string>();

            for (const [, { inverter, loggerName }] of inverterIndex) {
                const allStrings = inverter.strings || [];
                if (allStrings.length === 0) continue;

                const setupCount = parseInt(inverter.numberStringSet || '') || 0;
                
                const excludedIndices = new Set(
                    (inverter.excludedStrings || '')
                        .split(',')
                        .map(s => parseInt(s.trim(), 10))
                        .filter(n => !isNaN(n))
                );

                // Only consider strings that are <= setupCount and not excluded
                const strings = allStrings.filter(str => {
                    const idx = parseInt(str.id.replace('PV', ''), 10);
                    if (setupCount > 0 && idx > setupCount) return false;
                    if (excludedIndices.has(idx)) return false;
                    return true;
                });

                if (strings.length === 0) continue;

                // Per-inverter averages from healthy strings only
                const activeStrings = strings.filter(s => s.voltage > 10 && s.current > ALERT_CONFIG.minCurrentThresholdA);
                const avgVoltage = activeStrings.length > 0
                    ? activeStrings.reduce((sum, s) => sum + s.voltage, 0) / activeStrings.length : 0;
                const avgCurrent = activeStrings.length > 0
                    ? activeStrings.reduce((sum, s) => sum + s.current, 0) / activeStrings.length : 0;
                const thresholdV = avgVoltage * ALERT_CONFIG.thresholdPercent;
                const thresholdI = avgCurrent * ALERT_CONFIG.thresholdPercent;

                for (const str of strings) {
                    const hasVoltage = str.voltage > 10;
                    const hasCurrent = str.current > ALERT_CONFIG.minCurrentThresholdA;
                    const baseKey = `${inverter.id}-${str.id}`;
                    let faultType: PendingFault['faultType'] | null = null;
                    let faultMessage = '';

                    if (!hasVoltage && !hasCurrent) {
                        faultType = 'dead';
                        faultMessage = `${inverter.name} → ${str.id}: Mất dòng & mất điện áp`;
                    } else if (!hasCurrent && hasVoltage) {
                        faultType = 'noI';
                        faultMessage = `${inverter.name} → ${str.id}: Mất dòng (${str.current.toFixed(2)}A)`;
                    } else if (!hasVoltage && hasCurrent) {
                        faultType = 'noV';
                        faultMessage = `${inverter.name} → ${str.id}: Mất điện áp (${str.voltage.toFixed(1)}V)`;
                    } else if (activeStrings.length >= 2) {
                        // Threshold check – only meaningful if inverter has ≥2 healthy peers to compare
                        const warnings: string[] = [];
                        if (avgCurrent > 0 && str.current < thresholdI)
                            warnings.push(`${str.current.toFixed(2)}A < ${avgCurrent.toFixed(2)}A avg`);
                        if (avgVoltage > 0 && str.voltage < thresholdV)
                            warnings.push(`${str.voltage.toFixed(1)}V < ${avgVoltage.toFixed(1)}V avg`);
                        if (warnings.length > 0) {
                            faultType = 'thresh';
                            faultMessage = `${inverter.name} → ${str.id}: Vượt ngưỡng (${warnings.join(' | ')})`;
                        }
                    }

                    if (faultType) {
                        const faultKey = `${baseKey}-${faultType}`;
                        currentCycleFaultKeys.add(faultKey);
                        // Only start tracking if not already pending
                        if (!pendingFaultsRef.current.has(faultKey)) {
                            pendingFaultsRef.current.set(faultKey, {
                                firstSeenAt: nowMs,
                                faultType,
                                message: faultMessage,
                                inverterId: inverter.id,
                                inverterName: inverter.name,
                                loggerName: loggerName,
                            });
                        }
                    }
                }
            }

            // ── PASS 2: Self-healing – remove faults not seen this cycle ──
            for (const key of pendingFaultsRef.current.keys()) {
                if (!currentCycleFaultKeys.has(key)) {
                    pendingFaultsRef.current.delete(key);
                    confirmedFaultsRef.current.delete(key);
                }
            }

            // ── PASS 3: Evaluate fault age & classify ─────────────────────
            const newlyConfirmed: string[] = [];

            for (const [key, fault] of pendingFaultsRef.current) {
                const ageMs = nowMs - fault.firstSeenAt;

                if (ageMs >= ALERT_CONFIG.debounceMs) {
                    // CONFIRMED fault (persisted ≥ 15 min) → ERROR
                    if (!confirmedFaultsRef.current.has(key)) {
                        newlyConfirmed.push(key);
                        confirmedFaultsRef.current.add(key);
                    }
                    generatedAlerts.push({
                        id: `smart-${key}`,
                        timestamp: fault.firstSeenAt,
                        level: 'error',
                        message: `🔴 NGUY HIỂM [${Math.round(ageMs / 60000)}p] ${fault.loggerName} - ${fault.message}`,
                        source: `${fault.loggerName} - ${fault.inverterName}`,
                        deviceType: 'inverter',
                        deviceId: fault.inverterId,
                    });
                }
                // Faults within debounce window are silently tracked – no UI alert shown
            }

            // Queue sound alarm once per newly confirmed fault
            if (newlyConfirmed.length > 0) {
                newAudioToPlay.push('alarm');
            }
        }

        if (newAudioToPlay.length > 0) {
            // NOTE: Updating state inside useMemo is generally discouraged,
            // but we do it conditionally. A better standard is useEffect.
            // Queue audio in a setTimeout to avoid React warnings or synchronous loops during render
            setTimeout(() => {
                setAudioQueue(prev => [...prev, ...newAudioToPlay]);
            }, 0);
        }

        const enrichedServerAlerts = serverAlerts.map(enrichAlert);
        return [...enrichedServerAlerts, ...generatedAlerts];
    }, [sites, serverAlerts, inverterIndex, enrichAlert]);

    useEffect(() => {
        if (audioQueue.length > 0) {
            const nextAudio = audioQueue[0];
            if (nextAudio === 'happy') playHappyChime();
            else if (nextAudio === 'gentle') playGentleChime();
            else if (nextAudio === 'alarm') playAlarmSound();

            setAudioQueue(prev => prev.slice(1));
        }
    }, [audioQueue]);

    return smartAlerts;
}
