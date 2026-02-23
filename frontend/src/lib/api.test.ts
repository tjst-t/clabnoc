import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { fetchProjects, fetchTopology, getExecWSUrl, getEventsWSUrl } from './api'

describe('api', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('fetchProjects returns project list', async () => {
    const mockData = [{ name: 'test', nodes: 1, status: 'running', labdir: '/tmp/x' }]
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => mockData,
    } as Response)
    const result = await fetchProjects()
    expect(result).toEqual(mockData)
    expect(fetch).toHaveBeenCalledWith('/api/v1/projects')
  })

  it('fetchProjects throws on error', async () => {
    vi.mocked(fetch).mockResolvedValue({ ok: false, status: 500 } as Response)
    await expect(fetchProjects()).rejects.toThrow('500')
  })

  it('getExecWSUrl builds correct URL', () => {
    const url = getExecWSUrl('myproject', 'spine1')
    expect(url).toContain('/api/v1/projects/myproject/nodes/spine1/exec')
    expect(url).toContain('cmd=')
  })

  it('fetchTopology calls correct endpoint', async () => {
    const mockData = { name: 'test', nodes: [], links: [], groups: { dcs: [], racks: {} } }
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => mockData,
    } as Response)
    const result = await fetchTopology('myproject')
    expect(result).toEqual(mockData)
    expect(fetch).toHaveBeenCalledWith('/api/v1/projects/myproject/topology')
  })

  it('getEventsWSUrl builds URL without project', () => {
    const url = getEventsWSUrl()
    expect(url).toContain('/api/v1/events')
  })

  it('getEventsWSUrl builds URL with project', () => {
    const url = getEventsWSUrl('myproject')
    expect(url).toContain('project=myproject')
  })
})
