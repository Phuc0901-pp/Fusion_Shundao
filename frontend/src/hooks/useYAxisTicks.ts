import { useMemo } from 'react';

/**
 * Generates nicely-rounded Y-axis tick values for power charts.
 * @param dataMax - The maximum value observed in the visible data slice.
 * @param tickCount - Number of ticks to generate (default: 6).
 */
export function useYAxisTicks(dataMax: number, tickCount = 6): number[] {
  return useMemo(() => {
    if (dataMax === 0) return [0, 2000, 4000, 6000, 8000, 10000];
    const step = Math.ceil(dataMax / (tickCount - 1) / 100) * 100;
    const ticks: number[] = [];
    for (let i = 0; i < tickCount; i++) {
      ticks.push(i * step);
    }
    return ticks;
  }, [dataMax, tickCount]);
}
