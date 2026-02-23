import type { Project, TopologyData, NodeInfo, LinkInfo, NetemConfig } from '../types/topology';

const BASE = '/api/v1';

export async function fetchProjects(): Promise<Project[]> {
  const res = await fetch(`${BASE}/projects`);
  if (!res.ok) throw new Error(`fetch projects: ${res.status}`);
  return res.json();
}

export async function fetchTopology(project: string): Promise<TopologyData> {
  const res = await fetch(`${BASE}/projects/${encodeURIComponent(project)}/topology`);
  if (!res.ok) throw new Error(`fetch topology: ${res.status}`);
  return res.json();
}

export async function fetchNode(project: string, node: string): Promise<NodeInfo> {
  const res = await fetch(`${BASE}/projects/${encodeURIComponent(project)}/nodes/${encodeURIComponent(node)}`);
  if (!res.ok) throw new Error(`fetch node: ${res.status}`);
  return res.json();
}

export async function nodeAction(project: string, node: string, action: 'start' | 'stop' | 'restart'): Promise<void> {
  const res = await fetch(`${BASE}/projects/${encodeURIComponent(project)}/nodes/${encodeURIComponent(node)}/action`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ action }),
  });
  if (!res.ok) throw new Error(`node action: ${res.status}`);
}

export function getExecWSUrl(project: string, node: string, cmd = '/bin/bash'): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${proto}//${window.location.host}/api/v1/projects/${encodeURIComponent(project)}/nodes/${encodeURIComponent(node)}/exec?cmd=${encodeURIComponent(cmd)}`;
}

export function getSSHWSUrl(project: string, node: string, user = 'admin', port = 22): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${proto}//${window.location.host}/api/v1/projects/${encodeURIComponent(project)}/nodes/${encodeURIComponent(node)}/ssh?user=${encodeURIComponent(user)}&port=${port}`;
}

export async function fetchLinks(project: string): Promise<LinkInfo[]> {
  const res = await fetch(`${BASE}/projects/${encodeURIComponent(project)}/links`);
  if (!res.ok) throw new Error(`fetch links: ${res.status}`);
  return res.json();
}

export async function injectFault(project: string, linkId: string, action: 'down' | 'up' | 'netem' | 'clear_netem', netem?: NetemConfig): Promise<void> {
  const body: { action: string; netem?: NetemConfig } = { action };
  if (netem) body.netem = netem;
  const res = await fetch(
    `${BASE}/projects/${encodeURIComponent(project)}/links/${encodeURIComponent(linkId)}/fault`,
    { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) }
  );
  if (!res.ok) throw new Error(`fault injection: ${res.status}`);
}

export function getEventsWSUrl(project?: string): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const base = `${proto}//${window.location.host}/api/v1/events`;
  return project ? `${base}?project=${encodeURIComponent(project)}` : base;
}
