import type { ExternalNetworkLayout } from '../../lib/rack-layout';

interface Props {
  layout: ExternalNetworkLayout;
  selected: boolean;
  dimmed: boolean;
  onClick: () => void;
}

/**
 * Horizontal bar for bottom-positioned networks and collapsed mgmt networks.
 * Shows a dashed rectangle with label, and a "xN" badge when collapsed.
 */
export function ExternalNetworkBar({ layout, selected, dimmed, onClick }: Props) {
  const { x, y, width, height, network } = layout;

  const borderColor = selected ? 'var(--noc-device-name-selected)' : '#5a6a7e';
  const fillColor = selected ? 'rgba(0,212,170,0.04)' : 'rgba(90,106,126,0.03)';
  const isMgmt = network.name.startsWith('mgmt:');
  const dashArray = isMgmt ? '8,4' : '5,3';

  return (
    <g
      style={{
        cursor: 'pointer',
        opacity: dimmed ? 0.12 : 0.8,
        transition: 'opacity 0.2s ease',
      }}
      onPointerDown={(e) => e.stopPropagation()}
      onClick={(e) => {
        e.stopPropagation();
        onClick();
      }}
    >
      {/* Bar body */}
      <rect
        x={x}
        y={y}
        width={width}
        height={height}
        fill={fillColor}
        stroke={borderColor}
        strokeWidth={selected ? 1.8 : isMgmt ? 1.5 : 1}
        strokeDasharray={dashArray}
        rx={3}
      />

      {/* Network label */}
      <text
        x={x + 10}
        y={y + height / 2 + 1}
        fill={selected ? 'var(--noc-device-name-selected)' : 'var(--noc-text-dim)'}
        fontSize={8}
        fontFamily="'JetBrains Mono', monospace"
        dominantBaseline="central"
      >
        {network.label}
      </text>

      {/* Collapsed badge: xN */}
      {network.collapsed && network.link_count != null && network.link_count > 0 && (
        <g>
          <rect
            x={x + width - 38}
            y={y + (height - 14) / 2}
            width={28}
            height={14}
            fill="rgba(90,106,126,0.15)"
            stroke={borderColor}
            strokeWidth={0.6}
            rx={2}
          />
          <text
            x={x + width - 24}
            y={y + height / 2 + 1}
            fill="var(--noc-cyan)"
            fontSize={8}
            fontFamily="'JetBrains Mono', monospace"
            fontWeight={600}
            textAnchor="middle"
            dominantBaseline="central"
          >
            x{network.link_count}
          </text>
        </g>
      )}
    </g>
  );
}
