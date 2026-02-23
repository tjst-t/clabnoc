import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { FaultDialog } from './FaultDialog'
import type { LinkInfo } from '../types/topology'

const mockLink: LinkInfo = {
  id: 'spine1:e1-1__leaf1:e1-49',
  a: { node: 'spine1', interface: 'e1-1', mac: '' },
  z: { node: 'leaf1', interface: 'e1-49', mac: '' },
  state: 'up',
  netem: null,
}

describe('FaultDialog', () => {
  it('renders dialog with default values', () => {
    render(<FaultDialog link={mockLink} onApply={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.getByText('Apply netem')).toBeInTheDocument()
    expect(screen.getByText(/spine1:e1-1/)).toBeInTheDocument()
  })

  it('calls onCancel when Cancel clicked', () => {
    const onCancel = vi.fn()
    render(<FaultDialog link={mockLink} onApply={vi.fn()} onCancel={onCancel} />)
    fireEvent.click(screen.getByText('Cancel'))
    expect(onCancel).toHaveBeenCalled()
  })

  it('calls onApply with params when Apply clicked', () => {
    const onApply = vi.fn()
    render(<FaultDialog link={mockLink} onApply={onApply} onCancel={vi.fn()} />)
    fireEvent.click(screen.getByText('Apply'))
    expect(onApply).toHaveBeenCalledWith(mockLink.id, expect.objectContaining({
      delay_ms: expect.any(Number),
      loss_percent: expect.any(Number),
    }))
  })

  it('returns null when no link', () => {
    const { container } = render(<FaultDialog link={null} onApply={vi.fn()} onCancel={vi.fn()} />)
    expect(container.firstChild).toBeNull()
  })
})
