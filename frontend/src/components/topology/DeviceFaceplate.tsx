import { ROLE_COLORS, STATUS_COLORS, type DeviceLayout, type PortLayout } from '../../lib/rack-layout';

interface Props {
  device: DeviceLayout;
  ports: PortLayout[];
  selected: boolean;
  dimmed: boolean;
  highlightedPorts: Set<string>;
  cableColorMap: Map<string, string>;
  faultedPortColors: Map<string, string>;
  onClick: () => void;
  onPortClick: (port: PortLayout) => void;
}

function getDeviceColor(device: DeviceLayout): string {
  const status = device.node.status;
  if (status === 'stopped' || status === 'error') return '#e74c3c';
  if (status === 'warning') return '#f39c12';
  const colors = ROLE_COLORS[device.role];
  return colors ? colors.border : '#27ae60';
}

function getDeviceFill(device: DeviceLayout): string {
  const status = device.node.status;
  if (status === 'stopped' || status === 'error') return 'rgba(231,76,60,0.12)';
  if (status === 'warning') return 'rgba(243,156,18,0.08)';
  const colors = ROLE_COLORS[device.role];
  return colors ? colors.bg : 'rgba(39,174,96,0.06)';
}

function getPortColor(
  port: PortLayout,
  cableColorMap: Map<string, string>,
  faultedPortColors: Map<string, string>,
): string {
  // Selection highlight takes priority
  const cableColor = cableColorMap.get(port.key);
  if (cableColor) return cableColor;
  // Faulted port color as fallback (always visible)
  const faultColor = faultedPortColors.get(port.key);
  if (faultColor) return faultColor;
  return '#2a3a4a'; // unconnected
}

export function DeviceFaceplate({
  device,
  ports,
  selected,
  dimmed,
  highlightedPorts,
  cableColorMap,
  faultedPortColors,
  onClick,
  onPortClick,
}: Props) {
  const borderColor = selected ? '#00d4aa' : getDeviceColor(device);
  const fillColor = selected ? 'rgba(0,212,170,0.08)' : getDeviceFill(device);
  const statusColor = STATUS_COLORS[device.node.status] ?? '#2ecc71';
  const isError = device.node.status === 'stopped' || device.node.status === 'error';

  const nameY = device.height >= 36
    ? device.y + device.height - 6
    : device.y + device.height / 2 + 3;

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
      {/* Device body */}
      <rect
        x={device.x}
        y={device.y}
        width={device.width}
        height={device.height}
        fill={fillColor}
        stroke={borderColor}
        strokeWidth={selected ? 2 : 1.2}
        rx={2}
      />

      {/* Status LED */}
      <circle
        cx={device.x + 7}
        cy={device.y + device.height / 2}
        r={2.5}
        fill={statusColor}
      />

      {/* Error pulse animation */}
      {isError && (
        <circle
          cx={device.x + 7}
          cy={device.y + device.height / 2}
          r={2.5}
          fill="none"
          stroke="#e74c3c"
          strokeWidth={1}
        >
          <animate attributeName="r" from="2.5" to="9" dur="1.5s" repeatCount="indefinite" />
          <animate attributeName="opacity" from="1" to="0" dur="1.5s" repeatCount="indefinite" />
        </circle>
      )}

      {/* Device name */}
      <text
        x={device.x + 14}
        y={nameY}
        fill={selected ? '#00d4aa' : '#8899aa'}
        fontSize={8}
        fontFamily="'JetBrains Mono', monospace"
      >
        {device.node.name}
      </text>

      {/* Ports inside the faceplate */}
      {ports.map((port) => {
        const highlighted = highlightedPorts.has(port.key);
        const portColor = getPortColor(port, cableColorMap, faultedPortColors);

        return (
          <g
            key={port.key}
            style={{ cursor: 'pointer' }}
            onClick={(e) => {
              e.stopPropagation();
              onPortClick(port);
            }}
          >
            <rect
              x={port.x}
              y={port.y}
              width={port.w}
              height={port.h}
              fill={portColor}
              stroke="#0a0e17"
              strokeWidth={0.5}
              rx={1}
              opacity={0.9}
            />
            {/* Port highlight ring when cable is visible */}
            {highlighted && (
              <rect
                x={port.x - 1}
                y={port.y - 1}
                width={port.w + 2}
                height={port.h + 2}
                fill="none"
                stroke={portColor}
                strokeWidth={1.5}
                rx={2}
                opacity={0.9}
              />
            )}
            <title>{port.iface}</title>
          </g>
        );
      })}
    </g>
  );
}
