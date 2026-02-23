import { renderHook, act } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { useTerminalTabs } from './useTerminalTabs';

describe('useTerminalTabs', () => {
  it('starts with empty tabs', () => {
    const { result } = renderHook(() => useTerminalTabs('project-a'));
    expect(result.current.tabs).toEqual([]);
    expect(result.current.activeTabId).toBeNull();
  });

  it('adds a tab', () => {
    const { result } = renderHook(() => useTerminalTabs('project-a'));
    act(() => {
      result.current.addTab('spine1', 'exec');
    });
    expect(result.current.tabs).toHaveLength(1);
    expect(result.current.tabs[0]!.node).toBe('spine1');
    expect(result.current.tabs[0]!.type).toBe('exec');
    expect(result.current.activeTabId).toBe(result.current.tabs[0]!.id);
  });

  it('removes a tab', () => {
    const { result } = renderHook(() => useTerminalTabs('project-a'));
    act(() => {
      result.current.addTab('spine1', 'exec');
    });
    const tabId = result.current.tabs[0]!.id;
    act(() => {
      result.current.removeTab(tabId);
    });
    expect(result.current.tabs).toHaveLength(0);
  });

  it('sets active tab when adding', () => {
    const { result } = renderHook(() => useTerminalTabs('project-a'));
    act(() => {
      result.current.addTab('spine1', 'exec');
      result.current.addTab('leaf1', 'ssh');
    });
    expect(result.current.tabs).toHaveLength(2);
    // Active tab should be the last added
    expect(result.current.activeTabId).toBe(result.current.tabs[1]!.id);
  });

  it('generates correct labels', () => {
    const { result } = renderHook(() => useTerminalTabs('project-a'));
    act(() => {
      result.current.addTab('spine1', 'exec');
    });
    expect(result.current.tabs[0]!.label).toBe('spine1 (exec)');
  });
});
