import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { useCaptureStream } from './useCaptureStream';

// Mock WebSocket
class MockWebSocket {
  static instances: MockWebSocket[] = [];
  readyState = WebSocket.OPEN;
  onopen: (() => void) | null = null;
  onmessage: ((e: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  sentMessages: string[] = [];

  constructor(public url: string) {
    MockWebSocket.instances.push(this);
    // Simulate connection in next tick
    setTimeout(() => this.onopen?.(), 0);
  }

  send(data: string) {
    this.sentMessages.push(data);
  }

  close() {
    this.readyState = WebSocket.CLOSED;
    this.onclose?.();
  }

  simulateMessage(data: string) {
    this.onmessage?.({ data });
  }
}

describe('useCaptureStream', () => {
  beforeEach(() => {
    MockWebSocket.instances = [];
    vi.stubGlobal('WebSocket', MockWebSocket);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('returns empty packets initially', () => {
    const { result } = renderHook(() => useCaptureStream(null, null));
    expect(result.current.packets).toEqual([]);
    expect(result.current.connected).toBe(false);
    expect(result.current.paused).toBe(false);
  });

  it('connects when project and linkId are provided', async () => {
    renderHook(() => useCaptureStream('proj', 'link1'));
    expect(MockWebSocket.instances).toHaveLength(1);
    expect(MockWebSocket.instances[0]!.url).toContain('link1');
  });

  it('does not connect when project is null', () => {
    renderHook(() => useCaptureStream(null, 'link1'));
    expect(MockWebSocket.instances).toHaveLength(0);
  });

  it('does not connect when linkId is null', () => {
    renderHook(() => useCaptureStream('proj', null));
    expect(MockWebSocket.instances).toHaveLength(0);
  });

  it('sends pause message', async () => {
    const { result } = renderHook(() => useCaptureStream('proj', 'link1'));

    await act(async () => {
      await new Promise((r) => setTimeout(r, 10));
    });

    act(() => {
      result.current.pause();
    });

    expect(result.current.paused).toBe(true);
    const ws = MockWebSocket.instances[0]!;
    expect(ws.sentMessages).toContain('{"type":"pause"}');
  });

  it('sends resume message', async () => {
    const { result } = renderHook(() => useCaptureStream('proj', 'link1'));

    await act(async () => {
      await new Promise((r) => setTimeout(r, 10));
    });

    act(() => {
      result.current.pause();
    });

    act(() => {
      result.current.resume();
    });

    expect(result.current.paused).toBe(false);
    const ws = MockWebSocket.instances[0]!;
    expect(ws.sentMessages).toContain('{"type":"resume"}');
  });

  it('clears packets', async () => {
    const { result } = renderHook(() => useCaptureStream('proj', 'link1'));

    const ws = MockWebSocket.instances[0]!;
    act(() => {
      ws.simulateMessage(JSON.stringify({ no: 1, time: '12:00:00', source: '1.1.1.1', destination: '2.2.2.2', protocol: 'TCP', length: 64, info: 'test' }));
    });

    expect(result.current.packets).toHaveLength(1);

    act(() => {
      result.current.clear();
    });

    expect(result.current.packets).toHaveLength(0);
  });

  it('closes WebSocket on unmount', async () => {
    const { unmount } = renderHook(() => useCaptureStream('proj', 'link1'));
    const ws = MockWebSocket.instances[0]!;
    unmount();
    expect(ws.readyState).toBe(WebSocket.CLOSED);
  });
});
