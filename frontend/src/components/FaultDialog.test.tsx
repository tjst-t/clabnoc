import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { FaultDialog } from './FaultDialog';
import type { TopologyLink } from '../types/topology';

const mockLink: TopologyLink = {
  id: 'spine1:e1-1__leaf1:e1-49',
  a: { node: 'spine1', interface: 'e1-1' },
  z: { node: 'leaf1', interface: 'e1-49' },
  state: 'up',
  netem: null,
};

describe('FaultDialog', () => {
  it('renders netem parameter inputs', () => {
    render(<FaultDialog link={mockLink} onApply={() => {}} onClose={() => {}} />);
    expect(screen.getByLabelText(/delay/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/jitter/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/loss/i)).toBeInTheDocument();
  });

  it('renders endpoint info', () => {
    render(<FaultDialog link={mockLink} onApply={() => {}} onClose={() => {}} />);
    expect(screen.getByText(/spine1:e1-1/)).toBeInTheDocument();
    expect(screen.getByText(/leaf1:e1-49/)).toBeInTheDocument();
  });

  it('calls onApply with params when submitted', () => {
    const onApply = vi.fn();
    render(<FaultDialog link={mockLink} onApply={onApply} onClose={() => {}} />);

    const delayInput = screen.getByLabelText(/delay/i);
    fireEvent.change(delayInput, { target: { value: '100' } });

    const lossInput = screen.getByLabelText(/loss/i);
    fireEvent.change(lossInput, { target: { value: '30' } });

    fireEvent.click(screen.getByText('Apply Netem'));
    expect(onApply).toHaveBeenCalledWith('spine1:e1-1__leaf1:e1-49', expect.objectContaining({
      delay_ms: 100,
      loss_percent: 30,
    }));
  });

  it('calls onClose when clicking cancel', () => {
    const onClose = vi.fn();
    render(<FaultDialog link={mockLink} onApply={() => {}} onClose={onClose} />);
    fireEvent.click(screen.getByText('Cancel'));
    expect(onClose).toHaveBeenCalled();
  });

  it('pre-fills existing netem values', () => {
    const linkWithNetem = {
      ...mockLink,
      netem: { delay_ms: 50, jitter_ms: 5, loss_percent: 10, corrupt_percent: 0, duplicate_percent: 0 },
    };
    render(<FaultDialog link={linkWithNetem} onApply={() => {}} onClose={() => {}} />);
    expect(screen.getByLabelText(/delay/i)).toHaveValue(50);
    expect(screen.getByLabelText(/loss/i)).toHaveValue(10);
  });

  it('shows preview command', () => {
    render(<FaultDialog link={mockLink} onApply={() => {}} onClose={() => {}} />);
    expect(screen.getByText(/tc qdisc add/)).toBeInTheDocument();
  });
});
