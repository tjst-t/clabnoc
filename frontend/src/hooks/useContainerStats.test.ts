import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { useContainerStats } from './useContainerStats';

// Mock WebSocket
let mockWsInstances: MockWebSocket[] = [];

class MockWebSocket {
  url: string;
  onmessage: ((event: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  readyState = 1;
  closed = false;

  constructor(url: string) {
    this.url = url;
    mockWsInstances.push(this);
  }

  close() {
    this.closed = true;
  }

  simulateMessage(data: unknown) {
    if (this.onmessage) {
      this.onmessage({ data: JSON.stringify(data) });
    }
  }
}

beforeEach(() => {
  mockWsInstances = [];
  vi.stubGlobal('WebSocket', MockWebSocket);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('useContainerStats', () => {
  it('returns empty map when project is null', () => {
    const { result } = renderHook(() => useContainerStats(null));
    expect(result.current.size).toBe(0);
    expect(mockWsInstances).toHaveLength(0);
  });

  it('creates WebSocket when project is set', () => {
    renderHook(() => useContainerStats('test-project'));
    expect(mockWsInstances).toHaveLength(1);
    expect(mockWsInstances[0]!.url).toContain('/stats');
    expect(mockWsInstances[0]!.url).toContain('test-project');
  });

  it('parses valid stats messages', () => {
    const { result } = renderHook(() => useContainerStats('test-project'));
    const ws = mockWsInstances[0]!;

    act(() => {
      ws.simulateMessage({
        type: 'stats',
        stats: {
          spine1: { cpu_percent: 5.2, memory_bytes: 104857600, memory_limit: 8589934592 },
          leaf1: { cpu_percent: 1.1, memory_bytes: 52428800, memory_limit: 4294967296 },
        },
      });
    });

    expect(result.current.size).toBe(2);
    expect(result.current.get('spine1')?.cpu_percent).toBe(5.2);
    expect(result.current.get('leaf1')?.memory_bytes).toBe(52428800);
  });

  it('rejects stats with missing fields', () => {
    const { result } = renderHook(() => useContainerStats('test-project'));
    const ws = mockWsInstances[0]!;

    act(() => {
      ws.simulateMessage({
        type: 'stats',
        stats: {
          valid: { cpu_percent: 1.0, memory_bytes: 1024, memory_limit: 2048 },
          invalid_missing_limit: { cpu_percent: 1.0, memory_bytes: 1024 },
          invalid_string: 'not an object',
          invalid_null: null,
        },
      });
    });

    expect(result.current.size).toBe(1);
    expect(result.current.has('valid')).toBe(true);
    expect(result.current.has('invalid_missing_limit')).toBe(false);
  });

  it('ignores non-stats messages', () => {
    const { result } = renderHook(() => useContainerStats('test-project'));
    const ws = mockWsInstances[0]!;

    act(() => {
      ws.simulateMessage({ type: 'connected', data: { status: 'ok' } });
    });

    expect(result.current.size).toBe(0);
  });

  it('ignores malformed JSON', () => {
    const { result } = renderHook(() => useContainerStats('test-project'));
    const ws = mockWsInstances[0]!;

    act(() => {
      if (ws.onmessage) {
        ws.onmessage({ data: 'not json{{{' });
      }
    });

    expect(result.current.size).toBe(0);
  });

  it('closes WebSocket on unmount', () => {
    const { unmount } = renderHook(() => useContainerStats('test-project'));
    const ws = mockWsInstances[0]!;

    expect(ws.closed).toBe(false);
    unmount();
    expect(ws.closed).toBe(true);
  });

  it('closes old WebSocket when project changes', () => {
    const { rerender } = renderHook(
      ({ project }) => useContainerStats(project),
      { initialProps: { project: 'project-a' as string | null } }
    );
    const ws1 = mockWsInstances[0]!;

    rerender({ project: 'project-b' });
    expect(ws1.closed).toBe(true);
    expect(mockWsInstances.length).toBeGreaterThan(1);
  });

  it('clears stats when project changes to null', () => {
    const { result, rerender } = renderHook(
      ({ project }) => useContainerStats(project),
      { initialProps: { project: 'test' as string | null } }
    );
    const ws = mockWsInstances[0]!;

    act(() => {
      ws.simulateMessage({
        type: 'stats',
        stats: { spine1: { cpu_percent: 1, memory_bytes: 100, memory_limit: 200 } },
      });
    });
    expect(result.current.size).toBe(1);

    rerender({ project: null });
    expect(result.current.size).toBe(0);
  });
});
