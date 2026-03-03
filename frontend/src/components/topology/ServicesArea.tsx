interface Props {
  x: number;
  y: number;
  width: number;
  height: number;
  children?: React.ReactNode;
}

/**
 * Transparent container with subtle "Services" label for DC-level external nodes.
 * A thin dashed boundary that groups non-rack external infrastructure.
 */
export function ServicesArea({ x, y, width, height, children }: Props) {
  return (
    <g>
      {/* Boundary — subtle dashed line */}
      <rect
        x={x - 4}
        y={y - 14}
        width={width + 8}
        height={height + 18}
        fill="none"
        stroke="var(--noc-text-dim)"
        strokeWidth={0.6}
        strokeDasharray="3,4"
        rx={2}
        opacity={0.5}
      />

      {/* Section label */}
      <text
        x={x - 2}
        y={y - 4}
        fill="var(--noc-text-dim)"
        fontSize={7}
        fontFamily="'JetBrains Mono', monospace"
        letterSpacing="1.5px"
        opacity={0.7}
      >
        SERVICES
      </text>

      {children}
    </g>
  );
}
