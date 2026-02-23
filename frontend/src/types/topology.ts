export interface Project {
  name: string;
  nodes: number;
  status: 'running' | 'partial' | 'stopped';
  labdir: string;
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
  target: string;
}

export interface GraphInfo {
  dc: string;
  rack: string;
  rack_unit: number;
  role: string;
  icon: string;
  hidden: boolean;
}

export interface NodeInfo {
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

export interface Endpoint {
  node: string;
  interface: string;
  mac: string;
}

export interface NetemConfig {
  delay_ms: number;
  jitter_ms: number;
  loss_percent: number;
  corrupt_percent: number;
  duplicate_percent: number;
}

export interface LinkInfo {
  id: string;
  a: Endpoint;
  z: Endpoint;
  state: 'up' | 'down' | 'degraded';
  netem: NetemConfig | null;
}

export interface TopologyGroups {
  dcs: string[];
  racks: Record<string, string[]>;
}

export interface TopologyData {
  name: string;
  nodes: NodeInfo[];
  links: LinkInfo[];
  groups: TopologyGroups;
}
