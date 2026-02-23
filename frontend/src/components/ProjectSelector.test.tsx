import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ProjectSelector } from './ProjectSelector'
import type { Project } from '../types/topology'

const mockProjects: Project[] = [
  { name: 'dc-fabric', nodes: 12, status: 'running', labdir: '/tmp/clab-dc-fabric' },
  { name: 'test-lab', nodes: 3, status: 'stopped', labdir: '/tmp/clab-test-lab' },
]

describe('ProjectSelector', () => {
  it('renders project list', () => {
    render(<ProjectSelector projects={mockProjects} selected={null} onSelect={vi.fn()} loading={false} error={null} onRefresh={vi.fn()} />)
    expect(screen.getByText('dc-fabric')).toBeInTheDocument()
    expect(screen.getByText('test-lab')).toBeInTheDocument()
  })

  it('calls onSelect when project clicked', () => {
    const onSelect = vi.fn()
    render(<ProjectSelector projects={mockProjects} selected={null} onSelect={onSelect} loading={false} error={null} onRefresh={vi.fn()} />)
    fireEvent.click(screen.getByText('dc-fabric'))
    expect(onSelect).toHaveBeenCalledWith('dc-fabric')
  })

  it('shows loading state', () => {
    render(<ProjectSelector projects={[]} selected={null} onSelect={vi.fn()} loading={true} error={null} onRefresh={vi.fn()} />)
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('shows error state', () => {
    render(<ProjectSelector projects={[]} selected={null} onSelect={vi.fn()} loading={false} error="Network error" onRefresh={vi.fn()} />)
    expect(screen.getByText('Network error')).toBeInTheDocument()
  })

  it('marks selected project as active', () => {
    const { container } = render(<ProjectSelector projects={mockProjects} selected="dc-fabric" onSelect={vi.fn()} loading={false} error={null} onRefresh={vi.fn()} />)
    const activeItem = container.querySelector('.project-item.active')
    expect(activeItem).toBeInTheDocument()
    expect(activeItem?.textContent).toContain('dc-fabric')
  })

  it('shows empty state when no projects', () => {
    render(<ProjectSelector projects={[]} selected={null} onSelect={vi.fn()} loading={false} error={null} onRefresh={vi.fn()} />)
    expect(screen.getByText(/No clab projects/)).toBeInTheDocument()
  })
})
