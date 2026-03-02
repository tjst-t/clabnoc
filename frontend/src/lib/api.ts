import type {
  ProjectInfo,
  Topology,
  TopologyNode,
  TopologyLink,
  FaultRequest,
  NodeActionRequest,
  SSHCredentials,
  BPFPreset,
  CaptureSession,
  CaptureRequest,
} from '../types/topology';

const BASE_URL = '/api/v1';

async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const resp = await fetch(url, init);
  if (!resp.ok) {
    const text = await resp.text();
    throw new Error(`API error ${resp.status}: ${text}`);
  }
  return resp.json();
}

export async function getProjects(): Promise<ProjectInfo[]> {
  return fetchJSON<ProjectInfo[]>(`${BASE_URL}/projects`);
}

export async function getTopology(project: string): Promise<Topology> {
  return fetchJSON<Topology>(`${BASE_URL}/projects/${encodeURIComponent(project)}/topology`);
}

export async function getNodes(project: string): Promise<TopologyNode[]> {
  return fetchJSON<TopologyNode[]>(`${BASE_URL}/projects/${encodeURIComponent(project)}/nodes`);
}

export async function getNode(project: string, node: string): Promise<TopologyNode> {
  return fetchJSON<TopologyNode>(
    `${BASE_URL}/projects/${encodeURIComponent(project)}/nodes/${encodeURIComponent(node)}`
  );
}

export async function nodeAction(project: string, node: string, req: NodeActionRequest): Promise<void> {
  await fetchJSON(`${BASE_URL}/projects/${encodeURIComponent(project)}/nodes/${encodeURIComponent(node)}/action`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
}

export async function getLinks(project: string): Promise<TopologyLink[]> {
  return fetchJSON<TopologyLink[]>(`${BASE_URL}/projects/${encodeURIComponent(project)}/links`);
}

export async function injectFault(project: string, linkId: string, req: FaultRequest): Promise<void> {
  await fetchJSON(
    `${BASE_URL}/projects/${encodeURIComponent(project)}/links/${encodeURIComponent(linkId)}/fault`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    }
  );
}

export async function getBPFPresets(): Promise<BPFPreset[]> {
  return fetchJSON<BPFPreset[]>(`${BASE_URL}/bpf-presets`);
}

export async function captureAction(project: string, linkId: string, req: CaptureRequest): Promise<CaptureSession> {
  return fetchJSON<CaptureSession>(
    `${BASE_URL}/projects/${encodeURIComponent(project)}/links/${encodeURIComponent(linkId)}/capture`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    }
  );
}

export function getCaptureDownloadUrl(project: string, linkId: string): string {
  return `${BASE_URL}/projects/${encodeURIComponent(project)}/links/${encodeURIComponent(linkId)}/capture/download`;
}

export function createExecWebSocket(project: string, node: string, cmd = '/bin/bash'): WebSocket {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  return new WebSocket(
    `${proto}//${host}${BASE_URL}/projects/${encodeURIComponent(project)}/nodes/${encodeURIComponent(node)}/exec?cmd=${encodeURIComponent(cmd)}`
  );
}

export async function getSSHCredentials(project: string, node: string): Promise<SSHCredentials> {
  return fetchJSON<SSHCredentials>(
    `${BASE_URL}/projects/${encodeURIComponent(project)}/nodes/${encodeURIComponent(node)}/ssh-credentials`
  );
}

export function createSSHWebSocket(project: string, node: string): WebSocket {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  return new WebSocket(
    `${proto}//${host}${BASE_URL}/projects/${encodeURIComponent(project)}/nodes/${encodeURIComponent(node)}/ssh`
  );
}

export function createStatsWebSocket(project: string): WebSocket {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  return new WebSocket(
    `${proto}//${host}${BASE_URL}/projects/${encodeURIComponent(project)}/stats`
  );
}

export function createEventsWebSocket(project?: string): WebSocket {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  const params = project ? `?project=${encodeURIComponent(project)}` : '';
  return new WebSocket(`${proto}//${host}${BASE_URL}/events${params}`);
}
