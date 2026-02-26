import type { DCLayout } from '../../lib/rack-layout';

interface Props {
  dc: DCLayout;
  children: React.ReactNode;
}

export function DataCenter({ dc, children }: Props) {
  return (
    <g>
      {/* DC floor boundary — subtle rectangle */}
      <rect
        x={dc.x}
        y={dc.y}
        width={dc.width}
        height={dc.height}
        style={{ fill: 'var(--noc-rack-fill)', stroke: 'var(--noc-rack-stroke)' }}
        strokeWidth={1}
        rx={3}
        opacity={0.5}
      />
      {/* DC label */}
      <text
        x={dc.x + dc.width / 2}
        y={dc.y + 16}
        style={{ fill: 'var(--noc-dc-label)' }}
        fontSize={11}
        fontFamily="'JetBrains Mono', monospace"
        fontWeight={600}
        textAnchor="middle"
        letterSpacing="2px"
      >
        {dc.label}
      </text>
      {children}
    </g>
  );
}
