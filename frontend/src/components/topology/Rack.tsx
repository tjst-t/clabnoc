import type { RackLayout, UMarker } from '../../lib/rack-layout';
import { U_HEIGHT, RACK_TOP_MARGIN } from '../../lib/rack-layout';

interface Props {
  rack: RackLayout;
  uMarkers: UMarker[];
  children?: React.ReactNode;
}

export function Rack({ rack, uMarkers, children }: Props) {
  // Derive U count from rack height: height = units * U_HEIGHT + RACK_TOP_MARGIN + 10
  const rackUnits = Math.round((rack.height - RACK_TOP_MARGIN - 10) / U_HEIGHT);

  return (
    <g>
      {/* Rack cabinet body */}
      <rect
        x={rack.x}
        y={rack.y}
        width={rack.width}
        height={rack.height}
        style={{ fill: 'var(--noc-rack-fill)', stroke: 'var(--noc-rack-stroke)' }}
        strokeWidth={1.5}
        rx={3}
      />

      {/* Rack label */}
      <text
        x={rack.x + rack.width / 2}
        y={rack.y + 20}
        style={{ fill: 'var(--noc-dc-label)' }}
        fontSize={11}
        fontFamily="'JetBrains Mono', monospace"
        fontWeight={600}
        textAnchor="middle"
        letterSpacing="1px"
      >
        {rack.label}
      </text>

      {/* U-position ruler ticks */}
      {uMarkers.map((m) => (
        <text
          key={m.unit}
          x={rack.x + 5}
          y={m.y + U_HEIGHT - 4}
          style={{ fill: 'var(--noc-u-marker)' }}
          fontSize={7}
          fontFamily="monospace"
        >
          {m.unit}
        </text>
      ))}

      {/* Subtle horizontal grid lines for each U */}
      {Array.from({ length: rackUnits }, (_, i) => {
        const lineY = rack.y + RACK_TOP_MARGIN + i * U_HEIGHT;
        return (
          <line
            key={i}
            x1={rack.x + 1}
            y1={lineY}
            x2={rack.x + rack.width - 1}
            y2={lineY}
            style={{ stroke: 'var(--noc-grid-line)' }}
            strokeWidth={0.5}
          />
        );
      })}

      {children}
    </g>
  );
}
