import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { DetailPanel } from './DetailPanel';
import type { TopologyNode, TopologyLink } from '../types/topology';

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
  graph: { dc: 'dc1', rack: 'rack1', rack_unit: 40, rack_unit_size: 1, role: 'spine', icon: 'switch', hidden: false },
};

const mockLink: TopologyLink = {
  id: 'spine1:e1-1__leaf1:e1-49',
  a: { node: 'spine1', interface: 'e1-1', mac: 'aa:c1:ab:04:f3:6e' },
  z: { node: 'leaf1', interface: 'e1-49', mac: 'aa:c1:ab:9b:98:f7' },
  state: 'up',
  netem: null,
};

const defaultProps = {
  project: 'test-project',
  onOpenTerminal: vi.fn(),
  onNodeAction: vi.fn(),
  onFaultAction: vi.fn(),
  onOpenNetemDialog: vi.fn(),
};

describe('DetailPanel', () => {
  it('shows empty state when nothing selected', () => {
    render(<DetailPanel node={null} link={null} {...defaultProps} />);
    expect(screen.getByText(/no selection/i)).toBeInTheDocument();
  });

  it('renders node details when node is selected', () => {
    render(<DetailPanel node={mockNode} link={null} {...defaultProps} />);
    expect(screen.getByText('spine1')).toBeInTheDocument();
    expect(screen.getByText('running')).toBeInTheDocument();
    expect(screen.getByText('nokia_srlinux')).toBeInTheDocument();
  });

  it('renders access methods for node', () => {
    render(<DetailPanel node={mockNode} link={null} {...defaultProps} />);
    expect(screen.getByText('Console (docker exec)')).toBeInTheDocument();
    expect(screen.getByText('SSH')).toBeInTheDocument();
  });

  it('calls onOpenTerminal when clicking exec button', () => {
    const onOpen = vi.fn();
    render(<DetailPanel node={mockNode} link={null} {...defaultProps} onOpenTerminal={onOpen} />);
    fireEvent.click(screen.getByText('Console (docker exec)'));
    expect(onOpen).toHaveBeenCalledWith('spine1', 'exec');
  });

  it('shows stop/restart for running nodes', () => {
    render(<DetailPanel node={mockNode} link={null} {...defaultProps} />);
    expect(screen.getByText('Stop')).toBeInTheDocument();
    expect(screen.getByText('Restart')).toBeInTheDocument();
  });

  it('shows start for stopped nodes', () => {
    const stoppedNode = { ...mockNode, status: 'stopped' };
    render(<DetailPanel node={stoppedNode} link={null} {...defaultProps} />);
    expect(screen.getByText('Start')).toBeInTheDocument();
  });

  it('displays node labels', () => {
    render(<DetailPanel node={mockNode} link={null} {...defaultProps} />);
    expect(screen.getByText('graph-dc=')).toBeInTheDocument();
    expect(screen.getAllByText('dc1').length).toBeGreaterThanOrEqual(1);
  });

  it('renders link endpoints when link is selected', () => {
    render(<DetailPanel node={null} link={mockLink} {...defaultProps} />);
    expect(screen.getByText('spine1')).toBeInTheDocument();
    expect(screen.getByText('leaf1')).toBeInTheDocument();
    expect(screen.getByText('e1-1')).toBeInTheDocument();
    expect(screen.getByText('e1-49')).toBeInTheDocument();
  });

  it('shows link down button when link state is up', () => {
    render(<DetailPanel node={null} link={mockLink} {...defaultProps} />);
    expect(screen.getByText('Link Down')).toBeInTheDocument();
    expect(screen.getByText('Apply Netem')).toBeInTheDocument();
  });

  it('shows link up button when link state is down', () => {
    const downLink = { ...mockLink, state: 'down' as const };
    render(<DetailPanel node={null} link={downLink} {...defaultProps} />);
    expect(screen.getByText('Link Up')).toBeInTheDocument();
  });

  it('shows clear netem when link is degraded', () => {
    const degradedLink = {
      ...mockLink,
      state: 'degraded' as const,
      netem: { delay_ms: 100, jitter_ms: 10, loss_percent: 5, corrupt_percent: 0, duplicate_percent: 0 },
    };
    render(<DetailPanel node={null} link={degradedLink} {...defaultProps} />);
    expect(screen.getByText('Clear Netem')).toBeInTheDocument();
    expect(screen.getByText('100ms')).toBeInTheDocument();
  });

  it('calls onFaultAction when clicking link down', () => {
    const onFault = vi.fn();
    render(<DetailPanel node={null} link={mockLink} {...defaultProps} onFaultAction={onFault} />);
    fireEvent.click(screen.getByText('Link Down'));
    expect(onFault).toHaveBeenCalledWith('spine1:e1-1__leaf1:e1-49', 'down');
  });

  it('prefers node over link when both provided', () => {
    render(<DetailPanel node={mockNode} link={mockLink} {...defaultProps} />);
    // Should show node content (spine1 as header), not link content
    expect(screen.getByText('nokia_srlinux')).toBeInTheDocument();
  });

  it('renders VNC access as proxy link', () => {
    const bmcNode: TopologyNode = {
      ...mockNode,
      name: 'bmc1',
      access_methods: [
        { type: 'exec', label: 'Console (docker exec)' },
        { type: 'vnc', label: 'noVNC (BMC)', target: 'https://172.20.20.6/novnc/vnc.html' },
      ],
    };
    render(<DetailPanel node={bmcNode} link={null} {...defaultProps} />);
    const vncLink = screen.getByText('noVNC (BMC)');
    expect(vncLink.tagName).toBe('A');
    expect(vncLink).toHaveAttribute('target', '_blank');
    const href = vncLink.getAttribute('href')!;
    expect(href).toContain('/proxy/test-project/bmc1/novnc/vnc.html');
    expect(href).toContain('path=');
    expect(href).toContain('websockify');
    expect(href).not.toContain('172.20.20.6');
  });
});
