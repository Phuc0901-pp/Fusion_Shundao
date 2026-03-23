import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { CHART_MIN_VISIBLE_POINTS } from '../config/constants';

interface UseChartZoomOptions {
  dataLength: number;
}

export function useChartZoom({ dataLength }: UseChartZoomOptions) {
  const [zoomStart, setZoomStart] = useState(0);
  const [zoomEnd, setZoomEnd] = useState(dataLength);
  const containerRef = useRef<HTMLDivElement>(null);

  // Sync when data length changes (new data arrives)
  useEffect(() => {
    if (dataLength > 0) {
      setZoomStart(0);
      setZoomEnd(dataLength);
    }
  }, [dataLength]);

  const handleWheel = useCallback(
    (e: WheelEvent) => {
      if (!e.ctrlKey || dataLength === 0) return;
      e.preventDefault();
      e.stopPropagation();

      const container = containerRef.current;
      if (!container) return;

      const rect = container.getBoundingClientRect();
      const chartLeft = 45; // Approx Y-Axis width
      const chartRight = 15;
      const chartWidth = rect.width - chartLeft - chartRight;
      const mouseX = e.clientX - rect.left - chartLeft;
      const ratio = Math.max(0, Math.min(1, mouseX / chartWidth));

      const currentRange = zoomEnd - zoomStart;
      const zoomFactor = e.deltaY > 0 ? 1.15 : 0.85;
      let newRange = Math.round(currentRange * zoomFactor);

      newRange = Math.max(CHART_MIN_VISIBLE_POINTS, Math.min(dataLength, newRange));

      const centerIndex = zoomStart + ratio * currentRange;
      let newStart = Math.round(centerIndex - ratio * newRange);
      let newEnd = newStart + newRange;

      if (newStart < 0) { newStart = 0; newEnd = newRange; }
      if (newEnd > dataLength) { newEnd = dataLength; newStart = Math.max(0, newEnd - newRange); }

      setZoomStart(newStart);
      setZoomEnd(newEnd);
    },
    [dataLength, zoomStart, zoomEnd],
  );

  // Attach event listener
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;
    container.addEventListener('wheel', handleWheel, { passive: false });
    return () => container.removeEventListener('wheel', handleWheel);
  }, [handleWheel]);

  const visibleRange = useMemo(() => ({ start: zoomStart, end: zoomEnd }), [zoomStart, zoomEnd]);

  return { containerRef, visibleRange };
}
