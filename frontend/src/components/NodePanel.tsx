import type { TopologyNode, AccessMethod } from '../types/topology';

interface Props {
  node: TopologyNode;
  onClose: () => void;
  onOpenTerminal: (node: string, type: 'exec' | 'ssh') => void;
  onNodeAction: (node: string, action: 'start' | 'stop' | 'restart') => void;
}

function AccessButton({
  method,
  onOpenTerminal,
  nodeName,
}: {
  method: AccessMethod;
  onOpenTerminal: (node: string, type: 'exec' | 'ssh') => void;
  nodeName: string;
}) {
  if (method.type === 'vnc') {
    return (
      <a
        href={method.target}
        target="_blank"
        rel="noopener noreferrer"
        className="tui-btn tui-btn-amber"
      >
        {method.label}
      </a>
    );
  }

  return (
    <button
      onClick={() => onOpenTerminal(nodeName, method.type as 'exec' | 'ssh')}
      className="tui-btn tui-btn-cyan"
    >
      {method.label}
    </button>
  );
}

export function NodePanel({ node, onClose, onOpenTerminal, onNodeAction }: Props) {
  return (
    <div className="w-72 shrink-0 h-full bg-noc-surface tui-border-l overflow-y-auto animate-fade-in flex flex-col">
      {/* ─── Header ─── */}
      <div className="tui-border-b px-3 py-1.5 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-2 text-xs">
          <span
            className={
              node.status === 'running' ? 'text-noc-green' : 'text-noc-red'
            }
          >
            {node.status === 'running' ? '*' : 'x'}
          </span>
          <span className="text-noc-text-bright font-bold">{node.name}</span>
        </div>
        <button
          onClick={onClose}
          className="tui-btn tui-btn-dim"
        >
          x
        </button>
      </div>

      <div className="flex-1 overflow-y-auto">
        <div className="px-3 py-2 space-y-3">
          {/* ─── Info ─── */}
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
              <DetailRow
                label="Role"
                value={node.graph.role}
                valueClass="text-noc-cyan"
              />
            )}
          </TuiSection>

          {/* ─── Port Bindings ─── */}
          {node.port_bindings && node.port_bindings.length > 0 && (
            <TuiSection title="Ports">
              {node.port_bindings.map((pb, i) => (
                <div key={i} className="text-2xs text-noc-text-dim">
                  {pb.host_ip}:{pb.host_port} -&gt; {pb.port}/{pb.protocol}
                </div>
              ))}
            </TuiSection>
          )}

          {/* ─── Access ─── */}
          <TuiSection title="Access">
            <div className="flex flex-wrap gap-2">
              {node.access_methods.map((m, i) => (
                <AccessButton
                  key={i}
                  method={m}
                  onOpenTerminal={onOpenTerminal}
                  nodeName={node.name}
                />
              ))}
            </div>
          </TuiSection>

          {/* ─── Actions ─── */}
          <TuiSection title="Actions">
            <div className="flex gap-2">
              {node.status === 'running' ? (
                <>
                  <button
                    onClick={() => onNodeAction(node.name, 'stop')}
                    className="tui-btn tui-btn-red"
                  >
                    Stop
                  </button>
                  <button
                    onClick={() => onNodeAction(node.name, 'restart')}
                    className="tui-btn tui-btn-amber"
                  >
                    Restart
                  </button>
                </>
              ) : (
                <button
                  onClick={() => onNodeAction(node.name, 'start')}
                  className="tui-btn tui-btn-green"
                >
                  Start
                </button>
              )}
            </div>
          </TuiSection>

          {/* ─── Labels ─── */}
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

function DetailRow({
  label,
  value,
  valueClass,
}: {
  label: string;
  value: string;
  valueClass?: string;
}) {
  return (
    <div className="flex justify-between items-start gap-2 text-2xs">
      <span className="text-noc-text-dim shrink-0">{label}:</span>
      <span className={`text-right break-all ${valueClass ?? 'text-noc-text'}`}>{value}</span>
    </div>
  );
}
