import type { TopologyNode, TopologyLink, AccessMethod, ContainerStats } from '../types/topology';
import { formatBytes } from '../lib/format';

interface Props {
  project: string | null;
  node: TopologyNode | null;
  link: TopologyLink | null;
  onOpenTerminal: (node: string, type: 'exec' | 'ssh') => void;
  onNodeAction: (node: string, action: 'start' | 'stop' | 'restart') => void;
  onFaultAction: (linkId: string, action: 'up' | 'down' | 'clear_netem') => void;
  onOpenNetemDialog: (link: TopologyLink) => void;
  style?: React.CSSProperties;
  mobile?: boolean;
  containerStats?: Map<string, ContainerStats>;
}

export function DetailPanel({
  project,
  node,
  link,
  onOpenTerminal,
  onNodeAction,
  onFaultAction,
  onOpenNetemDialog,
  style,
  mobile,
  containerStats,
}: Props) {
  return (
    <div
      className={`bg-noc-surface flex flex-col ${
        mobile
          ? 'tui-border-t'
          : 'shrink-0 tui-border-l overflow-hidden'
      }`}
      style={style}
    >
      {node ? (
        <NodeContent
          project={project}
          node={node}
          onOpenTerminal={onOpenTerminal}
          onNodeAction={onNodeAction}
          stats={containerStats?.get(node.name)}
        />
      ) : link ? (
        <LinkContent
          link={link}
          onFaultAction={onFaultAction}
          onOpenNetemDialog={onOpenNetemDialog}
        />
      ) : (
        <EmptyContent />
      )}
    </div>
  );
}

function EmptyContent() {
  return (
    <div className="flex-1 flex items-center justify-center">
      <div className="text-center text-2xs text-noc-text-dim px-4">
        <div>-- No Selection --</div>
        <div className="mt-1">Click a node or link</div>
        <div>to view details</div>
      </div>
    </div>
  );
}

/* ─── Node Content ─── */

function NodeContent({
  project,
  node,
  onOpenTerminal,
  onNodeAction,
  stats,
}: {
  project: string | null;
  node: TopologyNode;
  onOpenTerminal: (node: string, type: 'exec' | 'ssh') => void;
  onNodeAction: (node: string, action: 'start' | 'stop' | 'restart') => void;
  stats?: ContainerStats;
}) {
  return (
    <>
      <div className="tui-border-b px-3 py-1.5 flex items-center gap-2 text-xs shrink-0">
        <span className={node.status === 'running' ? 'text-noc-green' : 'text-noc-red'}>
          {node.status === 'running' ? '*' : 'x'}
        </span>
        <span className="text-noc-text-bright font-bold">{node.name}</span>
      </div>

      <div className="flex-1 overflow-y-auto">
        <div className="px-3 py-2 space-y-3">
          <TuiSection title="Node">
            <DetailRow label="Kind" value={node.kind} />
            <DetailRow label="Image" value={node.image} />
            <DetailRow
              label="Status"
              value={node.status}
              valueClass={node.status === 'running' ? 'text-noc-green' : 'text-noc-red'}
            />
            {node.mgmt_ipv4 && <DetailRow label="Mgmt" value={node.mgmt_ipv4} />}
            {node.mgmt_ipv6 && <DetailRow label="IPv6" value={node.mgmt_ipv6} />}
            {node.container_id && (
              <DetailRow label="Container" value={node.container_id.slice(0, 12)} />
            )}
            {node.graph.dc && <DetailRow label="DC" value={node.graph.dc} />}
            {node.graph.rack && <DetailRow label="Rack" value={node.graph.rack} />}
            {node.graph.role && (
              <DetailRow label="Role" value={node.graph.role} valueClass="text-noc-cyan" />
            )}
          </TuiSection>

          {stats && (
            <TuiSection title="Resources">
              <DetailRow label="CPU" value={`${stats.cpu_percent.toFixed(1)}%`} valueClass="text-noc-cyan" />
              <DetailRow
                label="Memory"
                value={`${formatBytes(stats.memory_bytes)} / ${formatBytes(stats.memory_limit)}`}
                valueClass="text-noc-cyan"
              />
              <div className="mt-1 h-1.5 w-full rounded-sm overflow-hidden" style={{ background: 'var(--noc-border)' }}>
                <div
                  className="h-full transition-all duration-500"
                  style={{
                    width: `${stats.memory_limit > 0 ? Math.min((stats.memory_bytes / stats.memory_limit) * 100, 100) : 0}%`,
                    background: 'var(--noc-cyan)',
                  }}
                />
              </div>
            </TuiSection>
          )}

          {node.port_bindings && node.port_bindings.length > 0 && (
            <TuiSection title="Ports">
              {node.port_bindings.map((pb, i) => (
                <div key={i} className="text-2xs text-noc-text-dim">
                  {pb.host_ip}:{pb.host_port} -&gt; {pb.port}/{pb.protocol}
                </div>
              ))}
            </TuiSection>
          )}

          <TuiSection title="Access">
            <div className="flex flex-wrap gap-2">
              {node.access_methods.map((m, i) => (
                <AccessButton key={i} method={m} onOpenTerminal={onOpenTerminal} nodeName={node.name} project={project} />
              ))}
            </div>
          </TuiSection>

          <TuiSection title="Actions">
            <div className="flex gap-2">
              {node.status === 'running' ? (
                <>
                  <button onClick={() => onNodeAction(node.name, 'stop')} className="tui-btn tui-btn-red">
                    Stop
                  </button>
                  <button onClick={() => onNodeAction(node.name, 'restart')} className="tui-btn tui-btn-amber">
                    Restart
                  </button>
                </>
              ) : (
                <button onClick={() => onNodeAction(node.name, 'start')} className="tui-btn tui-btn-green">
                  Start
                </button>
              )}
            </div>
          </TuiSection>

          {Object.keys(node.labels).length > 0 && (
            <TuiSection title="Labels">
              {Object.entries(node.labels).map(([k, v]) => (
                <div key={k} className="text-2xs">
                  <span className="text-noc-text-dim">{k}=</span>
                  <span className="text-noc-text">{v}</span>
                </div>
              ))}
            </TuiSection>
          )}
        </div>
      </div>
    </>
  );
}

/* ─── Link Content ─── */

function LinkContent({
  link,
  onFaultAction,
  onOpenNetemDialog,
}: {
  link: TopologyLink;
  onFaultAction: (linkId: string, action: 'up' | 'down' | 'clear_netem') => void;
  onOpenNetemDialog: (link: TopologyLink) => void;
}) {
  const stateColor =
    link.state === 'up' ? 'text-noc-green' : link.state === 'down' ? 'text-noc-red' : 'text-noc-amber';
  const stateChar = link.state === 'up' ? '*' : link.state === 'down' ? 'x' : '~';
  const stateLabel = link.state.toUpperCase();

  return (
    <>
      <div className="tui-border-b px-3 py-1.5 flex items-center gap-2 text-xs shrink-0">
        <span className={stateColor}>{stateChar}</span>
        <span className="text-noc-text-bright font-bold">Link</span>
        <span className={`text-2xs ${stateColor}`}>[{stateLabel}]</span>
      </div>

      <div className="flex-1 overflow-y-auto">
        <div className="px-3 py-2 space-y-3">
          <TuiSection title="Endpoints">
            <EndpointInfo label="A" node={link.a.node} iface={link.a.interface} mac={link.a.mac} />
            <div className="text-2xs text-noc-text-dim text-center py-0.5">{'<'}--{'>'}</div>
            <EndpointInfo label="Z" node={link.z.node} iface={link.z.interface} mac={link.z.mac} />
          </TuiSection>

          {link.netem && (
            <TuiSection title="Active Netem">
              <div className="tui-border p-2 space-y-0.5">
                {link.netem.delay_ms > 0 && <NetemRow label="Delay" value={`${link.netem.delay_ms}ms`} />}
                {link.netem.jitter_ms > 0 && <NetemRow label="Jitter" value={`${link.netem.jitter_ms}ms`} />}
                {link.netem.loss_percent > 0 && <NetemRow label="Loss" value={`${link.netem.loss_percent}%`} />}
                {link.netem.corrupt_percent > 0 && <NetemRow label="Corrupt" value={`${link.netem.corrupt_percent}%`} />}
                {link.netem.duplicate_percent > 0 && <NetemRow label="Duplicate" value={`${link.netem.duplicate_percent}%`} />}
              </div>
            </TuiSection>
          )}

          <TuiSection title="Fault Injection">
            <div className="flex flex-wrap gap-2">
              {link.state === 'up' ? (
                <>
                  <button onClick={() => onFaultAction(link.id, 'down')} className="tui-btn tui-btn-red">
                    Link Down
                  </button>
                  <button onClick={() => onOpenNetemDialog(link)} className="tui-btn tui-btn-amber">
                    Apply Netem
                  </button>
                </>
              ) : link.state === 'down' ? (
                <button onClick={() => onFaultAction(link.id, 'up')} className="tui-btn tui-btn-green">
                  Link Up
                </button>
              ) : (
                <>
                  <button onClick={() => onFaultAction(link.id, 'clear_netem')} className="tui-btn tui-btn-green">
                    Clear Netem
                  </button>
                  <button onClick={() => onOpenNetemDialog(link)} className="tui-btn tui-btn-amber">
                    Update Netem
                  </button>
                </>
              )}
            </div>
          </TuiSection>
        </div>
      </div>
    </>
  );
}

/* ─── Shared components ─── */

function AccessButton({
  method,
  onOpenTerminal,
  nodeName,
  project,
}: {
  method: AccessMethod;
  onOpenTerminal: (node: string, type: 'exec' | 'ssh') => void;
  nodeName: string;
  project: string | null;
}) {
  if (method.type === 'vnc' && project) {
    const proxyBase = `/proxy/${encodeURIComponent(project)}/${encodeURIComponent(nodeName)}`;
    const href = `${proxyBase}/novnc/vnc.html?path=${encodeURIComponent(proxyBase.slice(1) + '/websockify')}&autoconnect=true`;
    return (
      <a href={href} target="_blank" rel="noopener noreferrer" className="tui-btn tui-btn-amber">
        {method.label}
      </a>
    );
  }
  return (
    <button onClick={() => onOpenTerminal(nodeName, method.type as 'exec' | 'ssh')} className="tui-btn tui-btn-cyan">
      {method.label}
    </button>
  );
}

function EndpointInfo({ label, node, iface, mac }: { label: string; node: string; iface: string; mac?: string }) {
  return (
    <div className="tui-border p-2">
      <div className="text-2xs text-noc-text-dim">{label}-side</div>
      <div className="text-xs text-noc-text-bright">{node}</div>
      <div className="text-2xs text-noc-cyan">{iface}</div>
      {mac && <div className="text-2xs text-noc-text-dim">{mac}</div>}
    </div>
  );
}

function TuiSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <div className="text-2xs text-noc-text-dim mb-1">
        {'--- '}
        {title}
        {' ---'}
      </div>
      <div className="space-y-0.5">{children}</div>
    </div>
  );
}

function DetailRow({ label, value, valueClass }: { label: string; value: string; valueClass?: string }) {
  return (
    <div className="flex justify-between items-start gap-2 text-2xs">
      <span className="text-noc-text-dim shrink-0">{label}:</span>
      <span className={`text-right break-all ${valueClass ?? 'text-noc-text'}`}>{value}</span>
    </div>
  );
}

function NetemRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between text-2xs">
      <span className="text-noc-amber">{label}</span>
      <span className="text-noc-text-bright">{value}</span>
    </div>
  );
}

