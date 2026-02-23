import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { LinkPanel } from './LinkPanel';
import type { TopologyLink } from '../types/topology';

const mockLink: TopologyLink = {
  id: 'spine1:e1-1__leaf1:e1-49',
  a: { node: 'spine1', interface: 'e1-1', mac: 'aa:c1:ab:04:f3:6e' },
  z: { node: 'leaf1', interface: 'e1-49', mac: 'aa:c1:ab:9b:98:f7' },
  state: 'up',
  netem: null,
};

describe('LinkPanel', () => {
  it('renders link endpoints', () => {
    render(
      <LinkPanel link={mockLink} onClose={() => {}} onFaultAction={() => {}} onOpenNetemDialog={() => {}} />
    );
    expect(screen.getByText('spine1')).toBeInTheDocument();
    expect(screen.getByText('leaf1')).toBeInTheDocument();
    expect(screen.getByText('e1-1')).toBeInTheDocument();
    expect(screen.getByText('e1-49')).toBeInTheDocument();
  });

  it('shows link down button when state is up', () => {
    render(
      <LinkPanel link={mockLink} onClose={() => {}} onFaultAction={() => {}} onOpenNetemDialog={() => {}} />
    );
    expect(screen.getByText('Link Down')).toBeInTheDocument();
    expect(screen.getByText('Apply Netem')).toBeInTheDocument();
  });

  it('shows link up button when state is down', () => {
    const downLink = { ...mockLink, state: 'down' as const };
    render(
      <LinkPanel link={downLink} onClose={() => {}} onFaultAction={() => {}} onOpenNetemDialog={() => {}} />
    );
    expect(screen.getByText('Link Up')).toBeInTheDocument();
  });

  it('shows clear netem button when degraded', () => {
    const degradedLink = {
      ...mockLink,
      state: 'degraded' as const,
      netem: { delay_ms: 100, jitter_ms: 10, loss_percent: 5, corrupt_percent: 0, duplicate_percent: 0 },
    };
    render(
      <LinkPanel link={degradedLink} onClose={() => {}} onFaultAction={() => {}} onOpenNetemDialog={() => {}} />
    );
    expect(screen.getByText('Clear Netem')).toBeInTheDocument();
    expect(screen.getByText('100ms')).toBeInTheDocument();
  });

  it('calls onFaultAction when clicking link down', () => {
    const onFault = vi.fn();
    render(
      <LinkPanel link={mockLink} onClose={() => {}} onFaultAction={onFault} onOpenNetemDialog={() => {}} />
    );
    fireEvent.click(screen.getByText('Link Down'));
    expect(onFault).toHaveBeenCalledWith('spine1:e1-1__leaf1:e1-49', 'down');
  });
});
