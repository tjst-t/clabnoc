import { useState, useCallback, useRef } from 'react';

interface UseResizableOptions {
  direction: 'horizontal' | 'vertical';
  initialSize: number;
  minSize: number;
  maxSize: number;
}

export function useResizable({ direction, initialSize, minSize, maxSize }: UseResizableOptions) {
  const [size, setSize] = useState(initialSize);
  const dragging = useRef(false);

  const handleMouseDown = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      dragging.current = true;

      const startPos = direction === 'horizontal' ? e.clientX : e.clientY;
      const startSize = size;
      const handle = e.currentTarget as HTMLElement;
      handle.classList.add('dragging');

      // Set cursor on body during drag to prevent flickering
      const prevCursor = document.body.style.cursor;
      document.body.style.cursor = direction === 'horizontal' ? 'col-resize' : 'row-resize';

      // Prevent text selection during drag
      const prevSelect = document.body.style.userSelect;
      document.body.style.userSelect = 'none';

      const onMouseMove = (ev: MouseEvent) => {
        const currentPos = direction === 'horizontal' ? ev.clientX : ev.clientY;
        // For horizontal (right panel): dragging left increases size, dragging right decreases
        // For vertical (bottom panel): dragging up increases size, dragging down decreases
        const delta = startPos - currentPos;
        const newSize = Math.min(maxSize, Math.max(minSize, startSize + delta));
        setSize(newSize);
      };

      const onMouseUp = () => {
        dragging.current = false;
        handle.classList.remove('dragging');
        document.body.style.cursor = prevCursor;
        document.body.style.userSelect = prevSelect;
        document.removeEventListener('mousemove', onMouseMove);
        document.removeEventListener('mouseup', onMouseUp);
      };

      document.addEventListener('mousemove', onMouseMove);
      document.addEventListener('mouseup', onMouseUp);
    },
    [direction, size, minSize, maxSize]
  );

  return { size, handleMouseDown };
}
