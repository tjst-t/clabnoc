import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useWebSocket } from './useWebSocket'

class MockWebSocket {
  static instances: MockWebSocket[] = []
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3
  readyState = 0 // CONNECTING
  onmessage: ((ev: MessageEvent) => void) | null = null
  onerror: ((ev: Event) => void) | null = null
  onclose: (() => void) | null = null
  onopen: (() => void) | null = null
  url: string

  constructor(url: string) {
    this.url = url
    MockWebSocket.instances.push(this)
    setTimeout(() => { this.readyState = MockWebSocket.OPEN }, 0)
  }

  close() { this.readyState = MockWebSocket.CLOSED }
  send() {}

  simulateMessage(data: string) {
    this.onmessage?.({ data } as MessageEvent)
  }
}

describe('useWebSocket', () => {
  beforeEach(() => {
    MockWebSocket.instances = []
    vi.stubGlobal('WebSocket', MockWebSocket)
  })
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('does not connect when url is null', () => {
    renderHook(() => useWebSocket({ url: null, onMessage: vi.fn() }))
    expect(MockWebSocket.instances).toHaveLength(0)
  })

  it('does not connect when disabled', () => {
    renderHook(() => useWebSocket({ url: 'ws://localhost/test', onMessage: vi.fn(), enabled: false }))
    expect(MockWebSocket.instances).toHaveLength(0)
  })

  it('connects when url is provided', () => {
    renderHook(() => useWebSocket({ url: 'ws://localhost/test', onMessage: vi.fn() }))
    expect(MockWebSocket.instances).toHaveLength(1)
  })

  it('calls onMessage with parsed JSON', () => {
    const onMessage = vi.fn()
    renderHook(() => useWebSocket({ url: 'ws://localhost/test', onMessage }))
    act(() => {
      MockWebSocket.instances[0]?.simulateMessage(JSON.stringify({ type: 'test', data: {} }))
    })
    expect(onMessage).toHaveBeenCalledWith({ type: 'test', data: {} })
  })

  it('ignores non-JSON messages without throwing', () => {
    const onMessage = vi.fn()
    expect(() => {
      renderHook(() => useWebSocket({ url: 'ws://localhost/test', onMessage }))
      act(() => { MockWebSocket.instances[0]?.simulateMessage('not-json') })
    }).not.toThrow()
  })
})
