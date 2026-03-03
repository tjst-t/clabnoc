import type { ExternalNetworkLayout } from '../../lib/rack-layout';

interface Props {
  layout: ExternalNetworkLayout;
  selected: boolean;
  dimmed: boolean;
  onClick: () => void;
}

/**
 * Cloud SVG element for top-positioned external networks (Internet, WAN).
 * Rendered as a rounded cloud outline with label text.
 */
export function ExternalNetworkCloud({ layout, selected, dimmed, onClick }: Props) {
  const { x, y, width, height } = layout;
  const cx = x + width / 2;
  const cy = y + height / 2;

  const borderColor = selected ? 'var(--noc-device-name-selected)' : '#7a8a9e';
  const fillColor = selected ? 'rgba(0,212,170,0.06)' : 'rgba(90,106,126,0.05)';

  // Cloud path: three overlapping arcs forming a cloud silhouette
  const hw = width * 0.48;
  const hh = height * 0.38;

  const cloudPath = [
    `M${cx - hw},${cy + hh * 0.3}`,
    `Q${cx - hw},${cy - hh * 0.6} ${cx - hw * 0.45},${cy - hh * 0.8}`,
    `Q${cx - hw * 0.15},${cy - hh * 1.2} ${cx + hw * 0.1},${cy - hh * 0.7}`,
    `Q${cx + hw * 0.4},${cy - hh * 1.3} ${cx + hw * 0.6},${cy - hh * 0.5}`,
    `Q${cx + hw},${cy - hh * 0.5} ${cx + hw},${cy + hh * 0.3}`,
    `Q${cx + hw},${cy + hh * 1.0} ${cx + hw * 0.3},${cy + hh * 1.0}`,
    `Q${cx},${cy + hh * 1.2} ${cx - hw * 0.4},${cy + hh * 1.0}`,
    `Q${cx - hw},${cy + hh * 1.0} ${cx - hw},${cy + hh * 0.3}`,
    'Z',
  ].join(' ');

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
      <path
        d={cloudPath}
        fill={fillColor}
        stroke={borderColor}
        strokeWidth={selected ? 2 : 1.2}
        strokeDasharray="6,3"
      />

      {/* Network label */}
      <text
        x={cx}
        y={cy + 2}
        fill={selected ? 'var(--noc-device-name-selected)' : 'var(--noc-device-name)'}
        fontSize={9}
        fontFamily="'JetBrains Mono', monospace"
        textAnchor="middle"
        dominantBaseline="central"
        letterSpacing="0.5px"
      >
        {layout.network.label}
      </text>
    </g>
  );
}
