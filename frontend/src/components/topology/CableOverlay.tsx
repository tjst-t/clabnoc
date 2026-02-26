import { memo } from 'react';
import type { CableLayout, RackLayout } from '../../lib/rack-layout';
import type { TopologyLink } from '../../types/topology';
import { buildCablePath, buildDirectCablePath, LINK_STATE_COLORS } from '../../lib/rack-layout';

interface Props {
  allCables: CableLayout[];
  highlightedCableIds: Set<string>;
  faultedCableIds: Set<string>;
  rackMap: Map<string, RackLayout>;
  totalWidth?: number;
  totalHeight?: number;
  selectedLinkId: string | null;
  onSelectLink: (link: TopologyLink) => void;
  onContextMenuLink: (link: TopologyLink, x: number, y: number) => void;
}

const BG_HIT_WIDTH = 6;
const FG_HIT_WIDTH = 12;

function getCableColor(link: TopologyLink): string {
  if (link.state === 'down') return LINK_STATE_COLORS.down!;
  if (link.state === 'degraded') return LINK_STATE_COLORS.degraded!;
  return LINK_STATE_COLORS.up!;
}

/** Background cables — always visible, thin bezier curves */
const BackgroundCables = memo(function BackgroundCables({
  cables,
  highlightedIds,
  faultedIds,
  onSelectLink,
  onContextMenuLink,
}: {
  cables: CableLayout[];
  highlightedIds: Set<string>;
  faultedIds: Set<string>;
  onSelectLink: (link: TopologyLink) => void;
  onContextMenuLink: (link: TopologyLink, x: number, y: number) => void;
}) {
  return (
    <g>
      {cables.map((cable) => {
        // Skip cables that will be drawn in foreground tiers
        if (highlightedIds.has(cable.link.id) || faultedIds.has(cable.link.id)) return null;

        const d = buildDirectCablePath(cable);
        return (
          <g key={`bg-${cable.link.id}`}>
            {/* Background cable line */}
            <path
              d={d}
              fill="none"
              stroke="#3498db"
              strokeWidth={0.6}
              opacity={0.18}
              strokeLinecap="round"
            />
            {/* Hit area for interaction */}
            <path
              d={d}
              fill="none"
              stroke="transparent"
              strokeWidth={BG_HIT_WIDTH}
              style={{ pointerEvents: 'stroke', cursor: 'pointer' }}
              onPointerDown={(e) => e.stopPropagation()}
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
});

/** Foreground cable (faulted or highlighted) with glow + orthogonal routing */
function ForegroundCable({
  cable,
  cableIndex,
  rackMap,
  isSelected,
  onSelectLink,
  onContextMenuLink,
}: {
  cable: CableLayout;
  cableIndex: number;
  rackMap: Map<string, RackLayout>;
  isSelected: boolean;
  onSelectLink: (link: TopologyLink) => void;
  onContextMenuLink: (link: TopologyLink, x: number, y: number) => void;
}) {
  const d = buildCablePath(cable, cableIndex, rackMap);
  const color = getCableColor(cable.link);

  return (
    <g>
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
        strokeWidth={isSelected ? 2.5 : 1.8}
        opacity={isSelected ? 1 : 0.85}
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
        strokeWidth={FG_HIT_WIDTH}
        style={{ pointerEvents: 'stroke', cursor: 'pointer' }}
        onPointerDown={(e) => e.stopPropagation()}
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
}

export function CableOverlay({
  allCables,
  highlightedCableIds,
  faultedCableIds,
  rackMap,
  selectedLinkId,
  onSelectLink,
  onContextMenuLink,
}: Props) {
  // Separate faulted and highlighted cables (preserving order for stagger index)
  const faultedCables: { cable: CableLayout; idx: number }[] = [];
  const highlightedCables: { cable: CableLayout; idx: number }[] = [];

  allCables.forEach((cable, idx) => {
    if (faultedCableIds.has(cable.link.id)) {
      faultedCables.push({ cable, idx });
    } else if (highlightedCableIds.has(cable.link.id)) {
      highlightedCables.push({ cable, idx });
    }
  });

  return (
    <g>
      {/* Tier 1: Background — always visible, bezier curves */}
      <BackgroundCables
        cables={allCables}
        highlightedIds={highlightedCableIds}
        faultedIds={faultedCableIds}
        onSelectLink={onSelectLink}
        onContextMenuLink={onContextMenuLink}
      />

      {/* Tier 2: Faulted cables — always prominent */}
      <g>
        {faultedCables.map(({ cable, idx }) => (
          <ForegroundCable
            key={`fault-${cable.link.id}`}
            cable={cable}
            cableIndex={idx}
            rackMap={rackMap}
            isSelected={cable.link.id === selectedLinkId}
            onSelectLink={onSelectLink}
            onContextMenuLink={onContextMenuLink}
          />
        ))}
      </g>

      {/* Tier 3: Highlighted cables — selection-driven */}
      <g>
        {highlightedCables.map(({ cable, idx }) => (
          <ForegroundCable
            key={`hl-${cable.link.id}`}
            cable={cable}
            cableIndex={idx}
            rackMap={rackMap}
            isSelected={cable.link.id === selectedLinkId}
            onSelectLink={onSelectLink}
            onContextMenuLink={onContextMenuLink}
          />
        ))}
      </g>
    </g>
  );
}
