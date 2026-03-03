import { EXTERNAL_NODE_COLOR, type ExternalNodeLayout } from '../../lib/rack-layout';

interface Props {
  layout: ExternalNodeLayout;
  selected: boolean;
  dimmed: boolean;
  onClick: () => void;
}

export function ExternalDevice({ layout, selected, dimmed, onClick }: Props) {
  const borderColor = selected ? 'var(--noc-device-name-selected)' : EXTERNAL_NODE_COLOR.border;
  const fillColor = selected ? 'var(--noc-device-selected-bg)' : EXTERNAL_NODE_COLOR.bg;

  const nameY = layout.height >= 36
    ? layout.y + layout.height - 6
    : layout.y + layout.height / 2 + 3;

  return (
    <g
      style={{
        cursor: 'pointer',
        opacity: dimmed ? 0.15 : 1,
        transition: 'opacity 0.2s ease',
      }}
      onPointerDown={(e) => e.stopPropagation()}
      onClick={(e) => {
        e.stopPropagation();
        onClick();
      }}
    >
      {/* Dashed border body */}
      <rect
        x={layout.x}
        y={layout.y}
        width={layout.width}
        height={layout.height}
        fill={fillColor}
        stroke={borderColor}
        strokeWidth={selected ? 2 : 1.2}
        strokeDasharray="4,3"
        rx={2}
      />

      {/* External marker — hollow diamond */}
      <polygon
        points={`${layout.x + 7},${layout.y + layout.height / 2 - 3} ${layout.x + 10},${layout.y + layout.height / 2} ${layout.x + 7},${layout.y + layout.height / 2 + 3} ${layout.x + 4},${layout.y + layout.height / 2}`}
        fill="none"
        stroke={borderColor}
        strokeWidth={1}
        opacity={0.85}
      />

      {/* Device label */}
      <text
        x={layout.x + 14}
        y={nameY}
        fill={selected ? 'var(--noc-device-name-selected)' : 'var(--noc-device-name)'}
        fontSize={8}
        fontFamily="'JetBrains Mono', monospace"
        fontStyle="italic"
      >
        {layout.node.label || layout.node.name}
      </text>
    </g>
  );
}
