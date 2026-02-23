import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ProjectSelector } from './ProjectSelector';
import type { ProjectInfo } from '../types/topology';

const mockProjects: ProjectInfo[] = [
  { name: 'dc-fabric', nodes: 5, status: 'running', labdir: '/tmp/clab-dc-fabric' },
  { name: 'wan-lab', nodes: 3, status: 'partial', labdir: '/tmp/clab-wan-lab' },
];

describe('ProjectSelector', () => {
  it('renders project options', () => {
    render(
      <ProjectSelector projects={mockProjects} selected={null} onSelect={() => {}} loading={false} />
    );
    const select = screen.getByRole('combobox');
    expect(select).toBeInTheDocument();
    expect(select).toHaveValue('');
  });

  it('shows selected project info', () => {
    render(
      <ProjectSelector projects={mockProjects} selected="dc-fabric" onSelect={() => {}} loading={false} />
    );
    expect(screen.getByText('5 nodes')).toBeInTheDocument();
    expect(screen.getByText('[running]')).toBeInTheDocument();
  });

  it('calls onSelect when changed', () => {
    const onSelect = vi.fn();
    render(
      <ProjectSelector projects={mockProjects} selected={null} onSelect={onSelect} loading={false} />
    );
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'dc-fabric' } });
    expect(onSelect).toHaveBeenCalledWith('dc-fabric');
  });

  it('disables when loading', () => {
    render(
      <ProjectSelector projects={[]} selected={null} onSelect={() => {}} loading={true} />
    );
    expect(screen.getByRole('combobox')).toBeDisabled();
  });
});
