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
