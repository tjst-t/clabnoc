import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { LinkPanel } from './LinkPanel'
import type { LinkInfo } from '../types/topology'

const mockLink: LinkInfo = {
  id: 'spine1:e1-1__leaf1:e1-49',
  a: { node: 'spine1', interface: 'e1-1', mac: 'aa:bb:cc:dd:ee:01' },
  z: { node: 'leaf1', interface: 'e1-49', mac: 'aa:bb:cc:dd:ee:02' },
  state: 'up',
  netem: null,
}

describe('LinkPanel', () => {
  it('renders link endpoints', () => {
    render(<LinkPanel link={mockLink} project="test" onFaultDown={vi.fn()} onFaultUp={vi.fn()} onFaultNetem={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByText(/spine1:e1-1/)).toBeInTheDocument()
    expect(screen.getByText(/leaf1:e1-49/)).toBeInTheDocument()
  })

  it('shows link state', () => {
    render(<LinkPanel link={mockLink} project="test" onFaultDown={vi.fn()} onFaultUp={vi.fn()} onFaultNetem={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByText(/up/)).toBeInTheDocument()
  })

  it('calls onFaultDown when Link Down clicked', () => {
    const onDown = vi.fn()
    render(<LinkPanel link={mockLink} project="test" onFaultDown={onDown} onFaultUp={vi.fn()} onFaultNetem={vi.fn()} onClose={vi.fn()} />)
    fireEvent.click(screen.getByText('Link Down'))
    expect(onDown).toHaveBeenCalledWith(mockLink.id)
  })

  it('disables Link Down when already down', () => {
    const downLink = { ...mockLink, state: 'down' as const }
    render(<LinkPanel link={downLink} project="test" onFaultDown={vi.fn()} onFaultUp={vi.fn()} onFaultNetem={vi.fn()} onClose={vi.fn()} />)
    expect(screen.getByText('Link Down')).toBeDisabled()
  })

  it('returns null when no link', () => {
    const { container } = render(<LinkPanel link={null} project="test" onFaultDown={vi.fn()} onFaultUp={vi.fn()} onFaultNetem={vi.fn()} onClose={vi.fn()} />)
    expect(container.firstChild).toBeNull()
  })
})
