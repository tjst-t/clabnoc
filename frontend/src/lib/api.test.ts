import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getProjects, getTopology, getNodes, nodeAction, injectFault } from './api';

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

beforeEach(() => {
  mockFetch.mockReset();
});

describe('API client', () => {
  it('getProjects calls correct endpoint', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve([{ name: 'test', nodes: 3, status: 'running' }]),
    });

    const projects = await getProjects();
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/projects', undefined);
    expect(projects).toHaveLength(1);
    expect(projects[0]!.name).toBe('test');
  });

  it('getTopology calls correct endpoint', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ name: 'test', nodes: [], links: [], groups: { dcs: [], racks: {} } }),
    });

    const topo = await getTopology('test');
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/projects/test/topology', undefined);
    expect(topo.name).toBe('test');
  });

  it('getNodes calls correct endpoint', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve([{ name: 'spine1' }]),
    });

    const nodes = await getNodes('test');
    expect(mockFetch).toHaveBeenCalledWith('/api/v1/projects/test/nodes', undefined);
    expect(nodes).toHaveLength(1);
  });

  it('nodeAction sends POST with body', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ status: 'ok' }),
    });

    await nodeAction('test', 'spine1', { action: 'stop' });
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/projects/test/nodes/spine1/action',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ action: 'stop' }),
      })
    );
  });

  it('injectFault sends POST with body', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ status: 'ok' }),
    });

    await injectFault('test', 'link1', { action: 'down' });
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/v1/projects/test/links/link1/fault',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ action: 'down' }),
      })
    );
  });

  it('throws on non-ok response', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      text: () => Promise.resolve('internal error'),
    });

    await expect(getProjects()).rejects.toThrow('API error 500');
  });
});
