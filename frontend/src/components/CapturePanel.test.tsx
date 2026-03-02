import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { CapturePanel } from './CapturePanel';
import type { PacketInfo } from '../types/topology';

// Mock the useCaptureStream hook
const mockPause = vi.fn();
const mockResume = vi.fn();
const mockClear = vi.fn();

let mockPackets: PacketInfo[] = [];
let mockConnected = true;
let mockPaused = false;

vi.mock('../hooks/useCaptureStream', () => ({
  useCaptureStream: () => ({
    packets: mockPackets,
    connected: mockConnected,
    paused: mockPaused,
    pause: mockPause,
    resume: mockResume,
    clear: mockClear,
  }),
}));

const defaultProps = {
  project: 'test-project',
  linkId: 'link1',
  linkLabel: 'spine1 <-> leaf1',
  onClose: vi.fn(),
};

describe('CapturePanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPackets = [];
    mockConnected = true;
    mockPaused = false;
  });

  it('renders header with link label', () => {
    render(<CapturePanel {...defaultProps} />);
    expect(screen.getByText('Capture')).toBeInTheDocument();
    expect(screen.getByText('spine1 <-> leaf1')).toBeInTheDocument();
  });

  it('shows LIVE indicator when connected and not paused', () => {
    render(<CapturePanel {...defaultProps} />);
    expect(screen.getByText('LIVE')).toBeInTheDocument();
  });

  it('shows PAUSED indicator when paused', () => {
    mockPaused = true;
    render(<CapturePanel {...defaultProps} />);
    expect(screen.getByText('PAUSED')).toBeInTheDocument();
  });

  it('shows waiting message when no packets', () => {
    render(<CapturePanel {...defaultProps} />);
    expect(screen.getByText('Waiting for packets...')).toBeInTheDocument();
  });

  it('shows disconnected message when not connected', () => {
    mockConnected = false;
    render(<CapturePanel {...defaultProps} />);
    expect(screen.getByText('Disconnected')).toBeInTheDocument();
  });

  it('renders packet rows', () => {
    mockPackets = [
      { no: 1, time: '12:34:56.789', source: '10.0.0.1', destination: '10.0.0.2', protocol: 'TCP', length: 64, info: 'Flags [S]' },
      { no: 2, time: '12:34:56.790', source: '10.0.0.2', destination: '10.0.0.1', protocol: 'TCP', length: 60, info: 'Flags [S.]' },
    ];
    render(<CapturePanel {...defaultProps} />);
    expect(screen.getAllByText('10.0.0.1')).toHaveLength(2);
    expect(screen.getAllByText('10.0.0.2')).toHaveLength(2);
    expect(screen.getAllByText('TCP')).toHaveLength(2);
  });

  it('calls pause when clicking pause button', () => {
    render(<CapturePanel {...defaultProps} />);
    fireEvent.click(screen.getByText('pause'));
    expect(mockPause).toHaveBeenCalled();
  });

  it('calls resume when paused and clicking resume', () => {
    mockPaused = true;
    render(<CapturePanel {...defaultProps} />);
    fireEvent.click(screen.getByText('resume'));
    expect(mockResume).toHaveBeenCalled();
  });

  it('calls clear when clicking clear button', () => {
    render(<CapturePanel {...defaultProps} />);
    fireEvent.click(screen.getByText('clear'));
    expect(mockClear).toHaveBeenCalled();
  });

  it('calls onClose when clicking close button', () => {
    const onClose = vi.fn();
    render(<CapturePanel {...defaultProps} onClose={onClose} />);
    fireEvent.click(screen.getByText('x'));
    expect(onClose).toHaveBeenCalled();
  });

  it('filters packets by display filter', () => {
    mockPackets = [
      { no: 1, time: '12:34:56.789', source: '10.0.0.1', destination: '10.0.0.2', protocol: 'TCP', length: 64, info: 'Flags [S]' },
      { no: 2, time: '12:34:56.790', source: '10.0.0.1', destination: '10.0.0.2', protocol: 'ICMP', length: 64, info: 'echo request' },
    ];
    render(<CapturePanel {...defaultProps} />);

    const filterInput = screen.getByPlaceholderText(/display filter/i);
    fireEvent.change(filterInput, { target: { value: 'ICMP' } });

    // Should only show ICMP row
    expect(screen.getByText('ICMP')).toBeInTheDocument();
    expect(screen.queryByText('Flags [S]')).not.toBeInTheDocument();
  });

  it('shows packet count', () => {
    mockPackets = [
      { no: 1, time: '12:34:56.789', source: '10.0.0.1', destination: '10.0.0.2', protocol: 'TCP', length: 64, info: 'test' },
    ];
    render(<CapturePanel {...defaultProps} />);
    expect(screen.getByText('1 pkts')).toBeInTheDocument();
  });

  it('renders table headers', () => {
    render(<CapturePanel {...defaultProps} />);
    expect(screen.getByText('No')).toBeInTheDocument();
    expect(screen.getByText('Time')).toBeInTheDocument();
    expect(screen.getByText('Source')).toBeInTheDocument();
    expect(screen.getByText('Destination')).toBeInTheDocument();
    expect(screen.getByText('Proto')).toBeInTheDocument();
    expect(screen.getByText('Len')).toBeInTheDocument();
    expect(screen.getByText('Info')).toBeInTheDocument();
  });
});
