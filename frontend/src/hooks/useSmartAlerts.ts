import { useRef, useMemo, useCallback, useEffect, useState } from 'react';
import type { Site, DeviceAlert, Inverter } from '../types';
import { playAlarmSound, playGentleChime, playHappyChime } from '../utils/audio';
import { SOLAR_START_HOUR, SOLAR_END_HOUR } from '../config/constants';

// ─── Types ──────────────────────────────────────────────────────────
interface InverterEntry {
    inverter: Inverter;
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

    // Track previously seen alert IDs from the server to play audio ONLY FOR NEW ONES
    const prevServerAlertIdsRef = useRef<Set<string>>(new Set());

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
        }
        prevHourRef.current = currentHour;

        // Alarm audio logic: check if serverAlerts contains brand NEW error-level alerts
        const currentServerAlertIds = new Set(serverAlerts.map(a => a.id));
        let hasNewServerAlarm = false;
        
        for (const a of serverAlerts) {
            if (a.level === 'error' && !prevServerAlertIdsRef.current.has(a.id)) {
                hasNewServerAlarm = true;
                break;
            }
        }
        
        prevServerAlertIdsRef.current = currentServerAlertIds;

        if (hasNewServerAlarm) {
            newAudioToPlay.push('alarm');
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
