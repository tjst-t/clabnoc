import type { DCLayout } from '../../lib/rack-layout';

interface Props {
  dc: DCLayout;
  children: React.ReactNode;
}

export function DataCenter({ dc, children }: Props) {
  return (
    <g>
      {/* DC floor boundary — subtle dark rectangle */}
      <rect
        x={dc.x}
        y={dc.y}
        width={dc.width}
        height={dc.height}
        fill="#0d1219"
        stroke="#1a2740"
        strokeWidth={1}
        rx={3}
        opacity={0.5}
      />
      {/* DC label */}
      <text
        x={dc.x + dc.width / 2}
        y={dc.y + 16}
        fill="#4a6a8a"
        fontSize={11}
        fontFamily="'JetBrains Mono', monospace"
        fontWeight={600}
        textAnchor="middle"
        letterSpacing="2px"
        style={{ textTransform: 'uppercase' }}
      >
        {dc.label}
      </text>
      {children}
    </g>
  );
}
