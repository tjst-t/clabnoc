import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { NodePanel } from './NodePanel'
import type { NodeInfo } from '../types/topology'

const mockNode: NodeInfo = {
  name: 'spine1',
  kind: 'nokia_srlinux',
  image: 'ghcr.io/nokia/srlinux:24.10.1',
  status: 'running',
  mgmt_ipv4: '172.20.20.2',
  mgmt_ipv6: '',
  container_id: 'abc123def456',
  labels: { 'graph-dc': 'dc1', 'graph-role': 'spine' },
  port_bindings: [],
  access_methods: [
    { type: 'exec', label: 'Console (docker exec)', target: '' },
    { type: 'ssh', label: 'SSH', target: '172.20.20.2:22' },
  ],
  graph: { dc: 'dc1', rack: 'rack1', rack_unit: 40, role: 'spine', icon: 'switch', hidden: false },
}

describe('NodePanel', () => {
  it('renders node details', () => {
    render(<NodePanel node={mockNode} project="test" onOpenTerminal={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByText('spine1')).toBeInTheDocument()
    expect(screen.getByText('nokia_srlinux')).toBeInTheDocument()
    expect(screen.getByText('running')).toBeInTheDocument()
    expect(screen.getByText('172.20.20.2')).toBeInTheDocument()
  })

  it('calls onOpenTerminal with exec type', () => {
    const onOpen = vi.fn()
    render(<NodePanel node={mockNode} project="test" onOpenTerminal={onOpen} onClose={vi.fn()} />)
    fireEvent.click(screen.getByText('Console (docker exec)'))
    expect(onOpen).toHaveBeenCalledWith('spine1', 'exec')
  })

  it('calls onOpenTerminal with ssh type', () => {
    const onOpen = vi.fn()
    render(<NodePanel node={mockNode} project="test" onOpenTerminal={onOpen} onClose={vi.fn()} />)
    fireEvent.click(screen.getByText('SSH'))
    expect(onOpen).toHaveBeenCalledWith('spine1', 'ssh')
  })

  it('calls onClose when close button clicked', () => {
    const onClose = vi.fn()
    render(<NodePanel node={mockNode} project="test" onOpenTerminal={vi.fn()} onClose={onClose} />)
    fireEvent.click(screen.getByText('✕'))
    expect(onClose).toHaveBeenCalled()
  })

  it('returns null when no node selected', () => {
    const { container } = render(<NodePanel node={null} project="test" onOpenTerminal={vi.fn()} onClose={vi.fn()} />)
    expect(container.firstChild).toBeNull()
  })

  it('calls onAction with start when Start clicked', () => {
    const onAction = vi.fn()
    const stoppedNode = { ...mockNode, status: 'exited' }
    render(<NodePanel node={stoppedNode} project="test" onOpenTerminal={vi.fn()} onClose={vi.fn()} onAction={onAction} />)
    fireEvent.click(screen.getByText('Start'))
    expect(onAction).toHaveBeenCalledWith('spine1', 'start')
  })

  it('calls onAction with stop when Stop clicked', () => {
    const onAction = vi.fn()
    render(<NodePanel node={mockNode} project="test" onOpenTerminal={vi.fn()} onClose={vi.fn()} onAction={onAction} />)
    fireEvent.click(screen.getByText('Stop'))
    expect(onAction).toHaveBeenCalledWith('spine1', 'stop')
  })

  it('disables Start when node is running', () => {
    render(<NodePanel node={mockNode} project="test" onOpenTerminal={vi.fn()} onClose={vi.fn()} onAction={vi.fn()} />)
    expect(screen.getByText('Start')).toBeDisabled()
  })
})
