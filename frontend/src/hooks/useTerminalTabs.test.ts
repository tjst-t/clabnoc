import { describe, it, expect } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useTerminalTabs } from './useTerminalTabs'

describe('useTerminalTabs', () => {
  it('starts with empty tabs', () => {
    const { result } = renderHook(() => useTerminalTabs())
    expect(result.current.tabs).toHaveLength(0)
    expect(result.current.activeTabId).toBeNull()
  })

  it('addTab adds a new tab and sets it active', () => {
    const { result } = renderHook(() => useTerminalTabs())
    act(() => {
      result.current.addTab('proj1', 'spine1', 'exec', 'ws://localhost/exec')
    })
    expect(result.current.tabs).toHaveLength(1)
    expect(result.current.tabs[0].node).toBe('spine1')
    expect(result.current.tabs[0].type).toBe('exec')
    expect(result.current.activeTabId).toBe(result.current.tabs[0].id)
  })

  it('removeTab removes the tab', () => {
    const { result } = renderHook(() => useTerminalTabs())
    let tabId: string = ''
    act(() => {
      tabId = result.current.addTab('proj1', 'spine1', 'exec', 'ws://localhost/exec')
    })
    act(() => {
      result.current.removeTab(tabId)
    })
    expect(result.current.tabs).toHaveLength(0)
  })

  it('switchProject preserves tabs per project', () => {
    const { result } = renderHook(() => useTerminalTabs())
    act(() => {
      result.current.switchProject('proj1')
      result.current.addTab('proj1', 'node1', 'exec', 'ws://localhost/exec1')
    })
    act(() => {
      result.current.switchProject('proj2')
    })
    expect(result.current.tabs).toHaveLength(0) // proj2 has no tabs
    act(() => {
      result.current.switchProject('proj1')
    })
    expect(result.current.tabs).toHaveLength(1) // proj1 tabs restored
  })
})
