import { EXTERNAL_NODE_COLOR, type ExternalNodeLayout } from '../../lib/rack-layout';

interface Props {
  layout: ExternalNodeLayout;
  selected: boolean;
  dimmed: boolean;
  onClick: () => void;
}

export function ExternalDevice({ layout, selected, dimmed, onClick }: Props) {
  const borderColor = selected ? 'var(--noc-device-name-selected)' : EXTERNAL_NODE_COLOR.border;
  const fillColor = selected ? 'var(--noc-device-selected-bg)' : 'rgba(127,140,141,0.04)';

  const nameY = layout.height >= 36
    ? layout.y + layout.height - 6
    : layout.y + layout.height / 2 + 3;

  return (
    <g
      style={{
        cursor: 'pointer',
        opacity: dimmed ? 0.15 : 0.85,
        transition: 'opacity 0.2s ease',
      }}
      onPointerDown={(e) => e.stopPropagation()}
      onClick={(e) => {
        e.stopPropagation();
        onClick();
      }}
    >
      {/* Dashed border body — ghostly external device */}
      <rect
        x={layout.x}
        y={layout.y}
        width={layout.width}
        height={layout.height}
        fill={fillColor}
        stroke={borderColor}
        strokeWidth={selected ? 1.8 : 1}
        strokeDasharray="4,3"
        rx={2}
      />

      {/* External marker — hollow diamond instead of status LED */}
      <polygon
        points={`${layout.x + 7},${layout.y + layout.height / 2 - 3} ${layout.x + 10},${layout.y + layout.height / 2} ${layout.x + 7},${layout.y + layout.height / 2 + 3} ${layout.x + 4},${layout.y + layout.height / 2}`}
        fill="none"
        stroke={EXTERNAL_NODE_COLOR.fill}
        strokeWidth={0.8}
        opacity={0.6}
      />

      {/* Device label */}
      <text
        x={layout.x + 14}
        y={nameY}
        fill={selected ? 'var(--noc-device-name-selected)' : 'var(--noc-text-dim)'}
        fontSize={8}
        fontFamily="'JetBrains Mono', monospace"
        fontStyle="italic"
      >
        {layout.node.label || layout.node.name}
      </text>
    </g>
  );
}
