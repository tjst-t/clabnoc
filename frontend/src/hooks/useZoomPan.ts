import { useState, useCallback, useEffect, useRef } from 'react';

interface ZoomPanState {
  scale: number;
  tx: number;
  ty: number;
}

interface UseZoomPanOptions {
  minScale?: number;
  maxScale?: number;
  wheelSensitivity?: number;
}

export function useZoomPan(options: UseZoomPanOptions = {}) {
  const { minScale = 0.2, maxScale = 4, wheelSensitivity = 0.002 } = options;

  const [state, setState] = useState<ZoomPanState>({ scale: 1, tx: 0, ty: 0 });
  const containerRef = useRef<HTMLDivElement>(null);
  const dragging = useRef(false);
  const dragStart = useRef({ x: 0, y: 0, tx: 0, ty: 0 });

  // Wheel zoom at cursor position
  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;

    const onWheel = (e: WheelEvent) => {
      e.preventDefault();
      const rect = el.getBoundingClientRect();
      const cursorX = e.clientX - rect.left;
      const cursorY = e.clientY - rect.top;

      setState((prev) => {
        const factor = 1 - e.deltaY * wheelSensitivity;
        const newScale = Math.min(maxScale, Math.max(minScale, prev.scale * factor));
        const ratio = newScale / prev.scale;

        // Zoom toward cursor: adjust translation so cursor position stays fixed
        const newTx = cursorX - ratio * (cursorX - prev.tx);
        const newTy = cursorY - ratio * (cursorY - prev.ty);

        return { scale: newScale, tx: newTx, ty: newTy };
      });
    };

    el.addEventListener('wheel', onWheel, { passive: false });
    return () => el.removeEventListener('wheel', onWheel);
  }, [minScale, maxScale, wheelSensitivity]);

  // Pointer drag for pan (background only)
  const onPointerDown = useCallback(
    (e: React.PointerEvent) => {
      // Only pan when clicking the background container itself
      if (e.target !== e.currentTarget) return;
      if (e.button !== 0) return; // left button only
      dragging.current = true;
      dragStart.current = { x: e.clientX, y: e.clientY, tx: state.tx, ty: state.ty };
      (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    },
    [state.tx, state.ty]
  );

  const onPointerMove = useCallback((e: React.PointerEvent) => {
    if (!dragging.current) return;
    const dx = e.clientX - dragStart.current.x;
    const dy = e.clientY - dragStart.current.y;
    setState((prev) => ({
      ...prev,
      tx: dragStart.current.tx + dx,
      ty: dragStart.current.ty + dy,
    }));
  }, []);

  const onPointerUp = useCallback(() => {
    dragging.current = false;
  }, []);

  // Fit content into visible area
  const fitContent = useCallback(
    (contentWidth: number, contentHeight: number) => {
      const el = containerRef.current;
      if (!el) return;
      const rect = el.getBoundingClientRect();
      const padding = 40;
      const scaleX = (rect.width - padding * 2) / contentWidth;
      const scaleY = (rect.height - padding * 2) / contentHeight;
      const scale = Math.min(Math.max(Math.min(scaleX, scaleY), minScale), maxScale);
      const tx = (rect.width - contentWidth * scale) / 2;
      const ty = (rect.height - contentHeight * scale) / 2;
      setState({ scale, tx, ty });
    },
    [minScale, maxScale]
  );

  const transform = `translate(${state.tx}px, ${state.ty}px) scale(${state.scale})`;

  return {
    containerRef,
    transform,
    scale: state.scale,
    onPointerDown,
    onPointerMove,
    onPointerUp,
    fitContent,
  };
}
