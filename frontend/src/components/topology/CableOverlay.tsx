import type { CableLayout, RackLayout } from '../../lib/rack-layout';
import type { TopologyLink } from '../../types/topology';
import { buildCablePath, LINK_STATE_COLORS } from '../../lib/rack-layout';

interface Props {
  cables: CableLayout[];
  rackMap: Map<string, RackLayout>;
  totalWidth?: number;
  totalHeight?: number;
  selectedLinkId: string | null;
  onSelectLink: (link: TopologyLink) => void;
  onContextMenuLink: (link: TopologyLink, x: number, y: number) => void;
}

const HIT_WIDTH = 12;

function getCableColor(link: TopologyLink): string {
  if (link.state === 'down') return LINK_STATE_COLORS.down!;
  if (link.state === 'degraded') return LINK_STATE_COLORS.degraded!;
  return LINK_STATE_COLORS.up!;
}

export function CableOverlay({
  cables,
  rackMap,
  selectedLinkId,
  onSelectLink,
  onContextMenuLink,
}: Props) {
  return (
    <g>
      {cables.map((cable, idx) => {
        const d = buildCablePath(cable, idx, rackMap);
        const selected = cable.link.id === selectedLinkId;
        const color = getCableColor(cable.link);

        return (
          <g key={cable.link.id}>
            {/* Glow effect */}
            <path
              d={d}
              fill="none"
              stroke={color}
              strokeWidth={5}
              opacity={0.12}
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            {/* Cable line */}
            <path
              d={d}
              fill="none"
              stroke={color}
              strokeWidth={selected ? 2.5 : 1.8}
              opacity={selected ? 1 : 0.85}
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeDasharray={
                cable.link.state === 'down' ? '6,4' :
                cable.link.state === 'degraded' ? '3,3' :
                undefined
              }
            />
            {/* Hit area */}
            <path
              d={d}
              fill="none"
              stroke="transparent"
              strokeWidth={HIT_WIDTH}
              style={{ pointerEvents: 'stroke', cursor: 'pointer' }}
              onClick={(e) => {
                e.stopPropagation();
                onSelectLink(cable.link);
              }}
              onContextMenu={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onContextMenuLink(cable.link, e.clientX, e.clientY);
              }}
            />
          </g>
        );
      })}
    </g>
  );
}
