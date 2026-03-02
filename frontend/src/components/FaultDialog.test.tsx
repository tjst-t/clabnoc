import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { FaultDialog } from './FaultDialog';
import type { TopologyLink } from '../types/topology';

vi.mock('../lib/api', () => ({
  getBPFPresets: vi.fn().mockResolvedValue([
    { name: 'DNS', filter: 'udp port 53', description: 'DNS traffic' },
    { name: 'BGP', filter: 'tcp port 179', description: 'BGP sessions' },
    { name: 'ICMP', filter: 'icmp', description: 'ICMP packets' },
  ]),
}));

const mockLink: TopologyLink = {
  id: 'spine1:e1-1__leaf1:e1-49',
  a: { node: 'spine1', interface: 'e1-1' },
  z: { node: 'leaf1', interface: 'e1-49' },
  state: 'up',
  netem: null,
};

describe('FaultDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

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

    fireEvent.click(screen.getByText('Apply'));
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

  // BPF filter tests

  it('renders BPF filter section', () => {
    render(<FaultDialog link={mockLink} onApply={() => {}} onClose={() => {}} />);
    expect(screen.getByLabelText(/preset/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/custom filter/i)).toBeInTheDocument();
  });

  it('loads BPF presets into dropdown', async () => {
    render(<FaultDialog link={mockLink} onApply={() => {}} onClose={() => {}} />);
    await waitFor(() => {
      expect(screen.getByText(/DNS \(udp port 53\)/)).toBeInTheDocument();
      expect(screen.getByText(/BGP \(tcp port 179\)/)).toBeInTheDocument();
    });
  });

  it('sets custom filter when preset is selected', async () => {
    render(<FaultDialog link={mockLink} onApply={() => {}} onClose={() => {}} />);
    await waitFor(() => {
      expect(screen.getByText(/BGP/)).toBeInTheDocument();
    });

    const select = screen.getByLabelText(/preset/i);
    fireEvent.change(select, { target: { value: 'BGP' } });

    const customInput = screen.getByLabelText(/custom filter/i);
    expect(customInput).toHaveValue('tcp port 179');
  });

  it('includes bpf_filter in params when submitted with filter', async () => {
    const onApply = vi.fn();
    render(<FaultDialog link={mockLink} onApply={onApply} onClose={() => {}} />);

    const customInput = screen.getByLabelText(/custom filter/i);
    fireEvent.change(customInput, { target: { value: 'tcp port 80' } });

    const delayInput = screen.getByLabelText(/delay/i);
    fireEvent.change(delayInput, { target: { value: '100' } });

    fireEvent.click(screen.getByText('Apply'));
    expect(onApply).toHaveBeenCalledWith('spine1:e1-1__leaf1:e1-49', expect.objectContaining({
      delay_ms: 100,
      bpf_filter: 'tcp port 80',
    }));
  });

  it('does not include bpf_filter when empty', () => {
    const onApply = vi.fn();
    render(<FaultDialog link={mockLink} onApply={onApply} onClose={() => {}} />);

    fireEvent.click(screen.getByText('Apply'));
    const calledParams = onApply.mock.calls[0]![1];
    expect(calledParams.bpf_filter).toBeUndefined();
  });

  it('shows prio+filter preview when BPF filter is set', () => {
    render(<FaultDialog link={mockLink} onApply={() => {}} onClose={() => {}} />);

    const customInput = screen.getByLabelText(/custom filter/i);
    fireEvent.change(customInput, { target: { value: 'icmp' } });

    expect(screen.getByText(/prio bands 3/)).toBeInTheDocument();
    expect(screen.getByText(/parent 1:1 netem/)).toBeInTheDocument();
    expect(screen.getByText(/flowid 1:1/)).toBeInTheDocument();
  });

  it('pre-fills existing bpf_filter value', () => {
    const linkWithBpf = {
      ...mockLink,
      netem: {
        delay_ms: 100,
        jitter_ms: 0,
        loss_percent: 5,
        corrupt_percent: 0,
        duplicate_percent: 0,
        bpf_filter: 'tcp port 179',
      },
    };
    render(<FaultDialog link={linkWithBpf} onApply={() => {}} onClose={() => {}} />);
    expect(screen.getByLabelText(/custom filter/i)).toHaveValue('tcp port 179');
  });
});
