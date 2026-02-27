import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { NodeTable } from './NodeTable';
import type { Topology, TopologyNode } from '../types/topology';

function makeNode(overrides: Partial<TopologyNode> & { name: string }): TopologyNode {
  return {
    kind: 'linux',
    image: 'alpine:latest',
    status: 'running',
    mgmt_ipv4: '172.20.0.2',
    mgmt_ipv6: '',
    container_id: 'abc123',
    labels: {},
    port_bindings: [],
    access_methods: [],
    graph: { dc: 'dc1', rack: 'rack-a', rack_unit: 10, rack_unit_size: 1, role: '', icon: '', hidden: false },
    ...overrides,
  };
}

const topology: Topology = {
  name: 'test',
  nodes: [
    makeNode({ name: 'spine1', kind: 'ceos', graph: { dc: 'dc1', rack: 'rack-a', rack_unit: 42, rack_unit_size: 1, role: '', icon: '', hidden: false } }),
    makeNode({ name: 'leaf1', kind: 'ceos', status: 'stopped', mgmt_ipv4: '172.20.0.3', graph: { dc: 'dc1', rack: 'rack-b', rack_unit: 20, rack_unit_size: 1, role: '', icon: '', hidden: false } }),
    makeNode({ name: 'server1', kind: 'linux', graph: { dc: 'dc2', rack: 'rack-c', rack_unit: 5, rack_unit_size: 1, role: '', icon: '', hidden: false } }),
  ],
  links: [],
  groups: { dcs: ['dc1', 'dc2'], racks: {} },
};

describe('NodeTable', () => {
  it('renders all nodes', () => {
    render(<NodeTable topology={topology} onSelectNode={() => {}} selectedNodeName={null} searchQuery="" />);
    expect(screen.getByText('spine1')).toBeTruthy();
    expect(screen.getByText('leaf1')).toBeTruthy();
    expect(screen.getByText('server1')).toBeTruthy();
  });

  it('shows empty state when no topology', () => {
    render(<NodeTable topology={null} onSelectNode={() => {}} selectedNodeName={null} searchQuery="" />);
    expect(screen.getByText('No project selected')).toBeTruthy();
  });

  it('filters by search query', () => {
    render(<NodeTable topology={topology} onSelectNode={() => {}} selectedNodeName={null} searchQuery="spine" />);
    expect(screen.getByText('spine1')).toBeTruthy();
    expect(screen.queryByText('leaf1')).toBeNull();
    expect(screen.queryByText('server1')).toBeNull();
  });

  it('filters by kind', () => {
    render(<NodeTable topology={topology} onSelectNode={() => {}} selectedNodeName={null} searchQuery="linux" />);
    expect(screen.getByText('server1')).toBeTruthy();
    expect(screen.queryByText('spine1')).toBeNull();
  });

  it('filters by status', () => {
    render(<NodeTable topology={topology} onSelectNode={() => {}} selectedNodeName={null} searchQuery="stopped" />);
    expect(screen.getByText('leaf1')).toBeTruthy();
    expect(screen.queryByText('spine1')).toBeNull();
  });

  it('calls onSelectNode when row clicked', () => {
    const onSelect = vi.fn();
    render(<NodeTable topology={topology} onSelectNode={onSelect} selectedNodeName={null} searchQuery="" />);
    fireEvent.click(screen.getByText('spine1'));
    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onSelect.mock.calls[0][0].name).toBe('spine1');
  });

  it('sorts by column header click', () => {
    const { container } = render(
      <NodeTable topology={topology} onSelectNode={() => {}} selectedNodeName={null} searchQuery="" />
    );
    // Default sort is name asc
    const rows = container.querySelectorAll('[class*="cursor-pointer"][class*="h-6"]');
    expect(rows.length).toBe(3);

    // Click Unit header to sort by unit
    fireEvent.click(screen.getByText('U'));
    const rowsAfter = container.querySelectorAll('[class*="cursor-pointer"][class*="h-6"]');
    // First row should be server1 (unit 5)
    expect(rowsAfter[0].textContent).toContain('server1');
  });

  it('shows footer with node count', () => {
    render(<NodeTable topology={topology} onSelectNode={() => {}} selectedNodeName={null} searchQuery="" />);
    expect(screen.getByText('3 nodes')).toBeTruthy();
  });

  it('shows filtered count in footer', () => {
    render(<NodeTable topology={topology} onSelectNode={() => {}} selectedNodeName={null} searchQuery="spine" />);
    expect(screen.getByText(/1 node.*filtered from 3/)).toBeTruthy();
  });
});
