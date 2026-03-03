export interface ProjectInfo {
  name: string;
  nodes: number;
  status: 'running' | 'partial' | 'stopped';
  labdir: string;
}

export interface Topology {
  name: string;
  nodes: TopologyNode[];
  links: TopologyLink[];
  groups: Groups;
  warnings?: string[];
  external_nodes?: ExternalNode[];
  external_networks?: ExternalNetwork[];
  external_links?: ExternalLink[];
}

export interface TopologyNode {
  name: string;
  kind: string;
  image: string;
  status: string;
  mgmt_ipv4: string;
  mgmt_ipv6: string;
  container_id: string;
  labels: Record<string, string>;
  port_bindings: PortBinding[];
  access_methods: AccessMethod[];
  graph: GraphInfo;
}

export interface PortBinding {
  host_ip: string;
  host_port: number;
  port: number;
  protocol: string;
}

export interface AccessMethod {
  type: 'exec' | 'ssh' | 'vnc';
  label: string;
  target?: string;
}

export interface GraphInfo {
  dc: string;
  rack: string;
  rack_unit: number;
  rack_unit_size: number; // Height in U (default 1)
  role: string;
  icon: string;
  hidden: boolean;
}

export interface TopologyLink {
  id: string;
  a: LinkEndpoint;
  z: LinkEndpoint;
  state: 'up' | 'down' | 'degraded';
  netem: NetemParams | null;
}

export interface LinkEndpoint {
  node: string;
  interface: string;
  mac?: string;
}

export interface NetemParams {
  delay_ms: number;
  jitter_ms: number;
  loss_percent: number;
  corrupt_percent: number;
  duplicate_percent: number;
  bpf_filter?: string;
}

export interface BPFPreset {
  name: string;
  filter: string;
  description: string;
}

export interface Groups {
  dcs: string[];
  racks: Record<string, string[]>;
  rack_units?: Record<string, number>; // rack name → total U count
}

export interface NodeActionRequest {
  action: 'start' | 'stop' | 'restart';
}

export interface FaultRequest {
  action: 'down' | 'up' | 'netem' | 'clear_netem';
  netem?: NetemParams;
}

export interface ApiEvent {
  type: string;
  project?: string;
  data: Record<string, string>;
}

export interface SSHCredentials {
  username: string;
  password: string;
  port: number;
}

export interface ContainerStats {
  cpu_percent: number;
  memory_bytes: number;
  memory_limit: number;
}

export interface TerminalTab {
  id: string;
  node: string;
  type: 'exec' | 'ssh';
  label: string;
  sshCredentials?: SSHCredentials;
}

export interface CaptureSession {
  id: string;
  link_id: string;
  interface: string;
  start_time: string;
  file_path: string;
  bpf_filter?: string;
  active: boolean;
}

export interface CaptureRequest {
  action: 'start' | 'stop';
  bpf_filter?: string;
}

export interface PacketInfo {
  no: number;
  time: string;
  source: string;
  destination: string;
  protocol: string;
  length: number;
  info: string;
}

// ── External entities ──

export interface ExternalNode {
  name: string;
  label: string;
  description?: string;
  icon: string;
  interfaces?: string[];
  graph: GraphInfo;
  external: true;
}

export interface ExternalNetwork {
  name: string;
  label: string;
  position: 'top' | 'bottom';
  dc?: string;
  collapsed?: boolean;
  link_count?: number;
}

export interface ExternalLink {
  id: string;
  a: ExternalEndpoint;
  z: ExternalEndpoint;
}

export interface ExternalEndpoint {
  node?: string;
  external?: string;
  network?: string;
  interface?: string;
}
