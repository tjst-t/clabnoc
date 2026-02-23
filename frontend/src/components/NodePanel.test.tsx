import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { NodePanel } from './NodePanel';
import type { TopologyNode } from '../types/topology';

const mockNode: TopologyNode = {
  name: 'spine1',
  kind: 'nokia_srlinux',
  image: 'ghcr.io/nokia/srlinux:24.10.1',
  status: 'running',
  mgmt_ipv4: '172.20.20.2',
  mgmt_ipv6: '3fff:172:20:20::2',
  container_id: 'abc123def456',
  labels: { 'graph-dc': 'dc1', 'graph-rack': 'rack1' },
  port_bindings: [],
  access_methods: [
    { type: 'exec', label: 'Console (docker exec)' },
    { type: 'ssh', label: 'SSH', target: '172.20.20.2:22' },
  ],
  graph: { dc: 'dc1', rack: 'rack1', rack_unit: 40, role: 'spine', icon: 'switch', hidden: false },
};

describe('NodePanel', () => {
  it('renders node name and status', () => {
    render(
      <NodePanel node={mockNode} onClose={() => {}} onOpenTerminal={() => {}} onNodeAction={() => {}} />
    );
    expect(screen.getByText('spine1')).toBeInTheDocument();
    expect(screen.getByText('running')).toBeInTheDocument();
  });

  it('renders access methods', () => {
    render(
      <NodePanel node={mockNode} onClose={() => {}} onOpenTerminal={() => {}} onNodeAction={() => {}} />
    );
    expect(screen.getByText('Console (docker exec)')).toBeInTheDocument();
    expect(screen.getByText('SSH')).toBeInTheDocument();
  });

  it('calls onOpenTerminal when clicking exec button', () => {
    const onOpen = vi.fn();
    render(
      <NodePanel node={mockNode} onClose={() => {}} onOpenTerminal={onOpen} onNodeAction={() => {}} />
    );
    fireEvent.click(screen.getByText('Console (docker exec)'));
    expect(onOpen).toHaveBeenCalledWith('spine1', 'exec');
  });

  it('calls onClose when clicking close button', () => {
    const onClose = vi.fn();
    render(
      <NodePanel node={mockNode} onClose={onClose} onOpenTerminal={() => {}} onNodeAction={() => {}} />
    );
    // Find close button (the X)
    const buttons = screen.getAllByRole('button');
    const closeBtn = buttons.find(b => b.querySelector('svg'));
    if (closeBtn) fireEvent.click(closeBtn);
    expect(onClose).toHaveBeenCalled();
  });

  it('shows stop/restart for running nodes', () => {
    render(
      <NodePanel node={mockNode} onClose={() => {}} onOpenTerminal={() => {}} onNodeAction={() => {}} />
    );
    expect(screen.getByText('Stop')).toBeInTheDocument();
    expect(screen.getByText('Restart')).toBeInTheDocument();
  });

  it('shows start for stopped nodes', () => {
    const stoppedNode = { ...mockNode, status: 'stopped' };
    render(
      <NodePanel node={stoppedNode} onClose={() => {}} onOpenTerminal={() => {}} onNodeAction={() => {}} />
    );
    expect(screen.getByText('Start')).toBeInTheDocument();
  });

  it('displays node labels', () => {
    render(
      <NodePanel node={mockNode} onClose={() => {}} onOpenTerminal={() => {}} onNodeAction={() => {}} />
    );
    expect(screen.getByText('graph-dc=')).toBeInTheDocument();
    // dc1 appears in both detail rows and labels
    expect(screen.getAllByText('dc1').length).toBeGreaterThanOrEqual(1);
  });
});
