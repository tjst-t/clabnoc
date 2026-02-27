import { useState, useMemo } from 'react';
import type { Topology, TopologyNode } from '../types/topology';

interface Props {
  topology: Topology | null;
  onSelectNode: (node: TopologyNode) => void;
  selectedNodeName: string | null;
  searchQuery: string;
}

type SortKey = 'name' | 'kind' | 'status' | 'mgmt_ipv4' | 'dc' | 'rack' | 'unit';
type SortDir = 'asc' | 'desc';

const COLUMNS: { key: SortKey; label: string; width: string }[] = [
  { key: 'name', label: 'NAME', width: 'flex-[2]' },
  { key: 'kind', label: 'KIND', width: 'flex-1' },
  { key: 'status', label: 'STATUS', width: 'w-20' },
  { key: 'mgmt_ipv4', label: 'MGMT IP', width: 'flex-1' },
  { key: 'dc', label: 'DC', width: 'w-16' },
  { key: 'rack', label: 'RACK', width: 'flex-1' },
  { key: 'unit', label: 'U', width: 'w-10' },
];

function getField(node: TopologyNode, key: SortKey): string | number {
  switch (key) {
    case 'name': return node.name;
    case 'kind': return node.kind;
    case 'status': return node.status;
    case 'mgmt_ipv4': return node.mgmt_ipv4 || '';
    case 'dc': return node.graph.dc || '';
    case 'rack': return node.graph.rack || '';
    case 'unit': return node.graph.rack_unit;
  }
}

export function NodeTable({ topology, onSelectNode, selectedNodeName, searchQuery }: Props) {
  const [sortKey, setSortKey] = useState<SortKey>('name');
  const [sortDir, setSortDir] = useState<SortDir>('asc');

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDir('asc');
    }
  };

  const rows = useMemo(() => {
    if (!topology) return [];
    let nodes = [...topology.nodes];

    // Filter by search
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      nodes = nodes.filter(
        (n) =>
          n.name.toLowerCase().includes(q) ||
          n.kind.toLowerCase().includes(q) ||
          n.status.toLowerCase().includes(q)
      );
    }

    // Sort
    nodes.sort((a, b) => {
      const aVal = getField(a, sortKey);
      const bVal = getField(b, sortKey);
      const cmp = typeof aVal === 'number' && typeof bVal === 'number'
        ? aVal - bVal
        : String(aVal).localeCompare(String(bVal));
      return sortDir === 'asc' ? cmp : -cmp;
    });

    return nodes;
  }, [topology, searchQuery, sortKey, sortDir]);

  if (!topology) {
    return (
      <div className="flex-1 flex items-center justify-center" style={{ background: 'var(--noc-topo-bg)' }}>
        <div className="text-center text-xs font-mono" style={{ color: 'var(--noc-empty-text)' }}>
          <div>No project selected</div>
          <div className="text-[10px] mt-1" style={{ color: 'var(--noc-empty-muted)' }}>
            Select a project to view nodes
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col overflow-hidden" style={{ background: 'var(--noc-topo-bg)' }}>
      {/* Header row */}
      <div
        className="flex items-center px-3 h-7 shrink-0 select-none"
        style={{ borderBottom: '1px solid var(--noc-border)' }}
      >
        {COLUMNS.map((col) => (
          <button
            key={col.key}
            onClick={() => handleSort(col.key)}
            className={`${col.width} text-left text-2xs font-bold tracking-wider px-1 py-0.5 cursor-pointer transition-colors hover:text-noc-accent shrink-0`}
            style={{ color: sortKey === col.key ? 'var(--noc-accent)' : 'var(--noc-text-dim)' }}
          >
            {col.label}
            {sortKey === col.key && (
              <span className="ml-0.5">{sortDir === 'asc' ? '\u25B2' : '\u25BC'}</span>
            )}
          </button>
        ))}
      </div>

      {/* Table body */}
      <div className="flex-1 overflow-y-auto">
        {rows.length === 0 ? (
          <div className="text-center text-2xs py-8" style={{ color: 'var(--noc-text-dim)' }}>
            {searchQuery ? `No nodes matching "${searchQuery}"` : 'No nodes'}
          </div>
        ) : (
          rows.map((node) => {
            const isSelected = node.name === selectedNodeName;
            return (
              <div
                key={node.name}
                onClick={() => onSelectNode(node)}
                className="flex items-center px-3 h-6 cursor-pointer transition-colors"
                style={{
                  borderBottom: '1px solid var(--noc-table-divider)',
                  background: isSelected ? 'var(--noc-device-selected-bg)' : undefined,
                }}
                onMouseEnter={(e) => {
                  if (!isSelected) e.currentTarget.style.background = 'var(--noc-surface)';
                }}
                onMouseLeave={(e) => {
                  if (!isSelected) e.currentTarget.style.background = '';
                }}
              >
                {/* Name */}
                <div className="flex-[2] text-2xs px-1 truncate" style={{ color: isSelected ? 'var(--noc-accent)' : 'var(--noc-text-bright)' }}>
                  {isSelected && <span style={{ color: 'var(--noc-accent)' }}>{'\u25B8'} </span>}
                  {node.name}
                </div>
                {/* Kind */}
                <div className="flex-1 text-2xs px-1 truncate" style={{ color: 'var(--noc-text)' }}>
                  {node.kind}
                </div>
                {/* Status */}
                <div className="w-20 text-2xs px-1 shrink-0">
                  <span style={{ color: node.status === 'running' ? 'var(--noc-green)' : 'var(--noc-red)' }}>
                    {node.status === 'running' ? '*' : 'x'} {node.status}
                  </span>
                </div>
                {/* Mgmt IP */}
                <div className="flex-1 text-2xs px-1 truncate" style={{ color: 'var(--noc-text-dim)' }}>
                  {node.mgmt_ipv4 || '--'}
                </div>
                {/* DC */}
                <div className="w-16 text-2xs px-1 truncate shrink-0" style={{ color: 'var(--noc-text-dim)' }}>
                  {node.graph.dc || '--'}
                </div>
                {/* Rack */}
                <div className="flex-1 text-2xs px-1 truncate" style={{ color: 'var(--noc-text-dim)' }}>
                  {node.graph.rack || '--'}
                </div>
                {/* Unit */}
                <div className="w-10 text-2xs px-1 shrink-0" style={{ color: 'var(--noc-text-dim)' }}>
                  {node.graph.rack_unit || '--'}
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Footer status bar */}
      <div
        className="h-5 flex items-center px-3 text-2xs shrink-0"
        style={{ borderTop: '1px solid var(--noc-border)', color: 'var(--noc-text-dim)' }}
      >
        {rows.length} node{rows.length !== 1 ? 's' : ''}
        {searchQuery && ` (filtered from ${topology.nodes.length})`}
      </div>
    </div>
  );
}
