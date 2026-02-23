import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { TerminalPanel } from './TerminalPanel';
import type { TerminalTab } from '../types/topology';

// Mock TerminalTab component since it uses xterm.js
vi.mock('./TerminalTab', () => ({
  TerminalTab: ({ tab }: { tab: TerminalTab }) => (
    <div data-testid={`terminal-${tab.id}`}>Terminal: {tab.label}</div>
  ),
}));

const mockTabs: TerminalTab[] = [
  { id: 'tab1', node: 'spine1', type: 'exec', label: 'spine1 (exec)' },
  { id: 'tab2', node: 'leaf1', type: 'ssh', label: 'leaf1 (ssh)' },
];

describe('TerminalPanel', () => {
  it('renders tab labels', () => {
    render(
      <TerminalPanel
        project="test"
        tabs={mockTabs}
        activeTabId="tab1"
        onSelectTab={() => {}}
        onCloseTab={() => {}}
        collapsed={false}
        onToggle={() => {}}
      />
    );
    expect(screen.getByText('spine1 (exec)')).toBeInTheDocument();
    expect(screen.getByText('leaf1 (ssh)')).toBeInTheDocument();
  });

  it('calls onSelectTab when clicking a tab', () => {
    const onSelect = vi.fn();
    render(
      <TerminalPanel
        project="test"
        tabs={mockTabs}
        activeTabId="tab1"
        onSelectTab={onSelect}
        onCloseTab={() => {}}
        collapsed={false}
        onToggle={() => {}}
      />
    );
    fireEvent.click(screen.getByText('leaf1 (ssh)'));
    expect(onSelect).toHaveBeenCalledWith('tab2');
  });

  it('shows empty message when no tabs', () => {
    render(
      <TerminalPanel
        project="test"
        tabs={[]}
        activeTabId={null}
        onSelectTab={() => {}}
        onCloseTab={() => {}}
        collapsed={false}
        onToggle={() => {}}
      />
    );
    expect(screen.getByText(/select a node/i)).toBeInTheDocument();
  });

  it('shows tab count', () => {
    render(
      <TerminalPanel
        project="test"
        tabs={mockTabs}
        activeTabId="tab1"
        onSelectTab={() => {}}
        onCloseTab={() => {}}
        collapsed={false}
        onToggle={() => {}}
      />
    );
    expect(screen.getByText('[2]')).toBeInTheDocument();
  });

  it('calls onToggle when clicking header', () => {
    const onToggle = vi.fn();
    render(
      <TerminalPanel
        project="test"
        tabs={mockTabs}
        activeTabId="tab1"
        onSelectTab={() => {}}
        onCloseTab={() => {}}
        collapsed={true}
        onToggle={onToggle}
      />
    );
    fireEvent.click(screen.getByText('Terminal'));
    expect(onToggle).toHaveBeenCalled();
  });
});
