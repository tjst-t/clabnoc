import type { TopologyNode, AccessMethod } from '../types/topology';

interface Props {
  node: TopologyNode;
  onClose: () => void;
  onOpenTerminal: (node: string, type: 'exec' | 'ssh') => void;
  onNodeAction: (node: string, action: 'start' | 'stop' | 'restart') => void;
}

const ROLE_BADGE_COLORS: Record<string, string> = {
  spine: 'border-noc-cyan text-noc-cyan',
  leaf: 'border-noc-accent text-noc-accent',
  router: 'border-noc-cyan text-noc-cyan',
  switch: 'border-noc-accent text-noc-accent',
  server: 'border-noc-amber text-noc-amber',
  bmc: 'border-noc-red text-noc-red',
};

function AccessButton({ method, onOpenTerminal, nodeName }: {
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
        className="flex items-center gap-2 px-3 py-2 bg-noc-surface border border-noc-border
                   rounded text-xs font-mono hover:border-noc-amber hover:text-noc-amber
                   transition-colors cursor-pointer"
      >
        <span className="w-1.5 h-1.5 rounded-full bg-noc-amber" />
        {method.label}
      </a>
    );
  }

  return (
    <button
      onClick={() => onOpenTerminal(nodeName, method.type as 'exec' | 'ssh')}
      className="flex items-center gap-2 px-3 py-2 bg-noc-surface border border-noc-border
                 rounded text-xs font-mono hover:border-noc-accent hover:text-noc-accent
                 transition-colors cursor-pointer text-left w-full"
    >
      <span className="w-1.5 h-1.5 rounded-full bg-noc-accent" />
      {method.label}
    </button>
  );
}

export function NodePanel({ node, onClose, onOpenTerminal, onNodeAction }: Props) {
  const roleClass = ROLE_BADGE_COLORS[node.graph.role] ?? 'border-noc-text-dim text-noc-text-dim';

  return (
    <div className="w-80 h-full bg-noc-panel border-l border-noc-border overflow-y-auto animate-fade-in">
      {/* Header */}
      <div className="sticky top-0 bg-noc-panel border-b border-noc-border px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span
            className={`w-2 h-2 rounded-full ${
              node.status === 'running' ? 'bg-noc-green' : 'bg-noc-red'
            }`}
          />
          <h2 className="font-mono text-sm font-semibold text-noc-text-bright">{node.name}</h2>
        </div>
        <button
          onClick={onClose}
          className="text-noc-text-dim hover:text-noc-text transition-colors p-1"
        >
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path d="M1 1L13 13M13 1L1 13" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
        </button>
      </div>

      <div className="p-4 space-y-4">
        {/* Role & Kind */}
        <div className="flex items-center gap-2">
          <span className={`px-2 py-0.5 border rounded text-2xs font-mono uppercase tracking-wider ${roleClass}`}>
            {node.graph.role || 'unknown'}
          </span>
          <span className="text-2xs font-mono text-noc-text-dim">{node.kind}</span>
        </div>

        {/* Details table */}
        <div className="space-y-1.5">
          <DetailRow label="Image" value={node.image} />
          <DetailRow label="Status" value={node.status} valueClass={
            node.status === 'running' ? 'text-noc-green' : 'text-noc-red'
          } />
          {node.mgmt_ipv4 && <DetailRow label="Mgmt IPv4" value={node.mgmt_ipv4} />}
          {node.mgmt_ipv6 && <DetailRow label="Mgmt IPv6" value={node.mgmt_ipv6} />}
          {node.container_id && (
            <DetailRow label="Container" value={node.container_id.slice(0, 12)} />
          )}
          {node.graph.dc && <DetailRow label="DC" value={node.graph.dc} />}
          {node.graph.rack && <DetailRow label="Rack" value={node.graph.rack} />}
        </div>

        {/* Port bindings */}
        {node.port_bindings && node.port_bindings.length > 0 && (
          <div>
            <SectionLabel>Port Bindings</SectionLabel>
            <div className="space-y-1">
              {node.port_bindings.map((pb, i) => (
                <div key={i} className="font-mono text-2xs text-noc-text-dim">
                  {pb.host_ip}:{pb.host_port} → {pb.port}/{pb.protocol}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Access methods */}
        <div>
          <SectionLabel>Access</SectionLabel>
          <div className="space-y-1.5">
            {node.access_methods.map((m, i) => (
              <AccessButton key={i} method={m} onOpenTerminal={onOpenTerminal} nodeName={node.name} />
            ))}
          </div>
        </div>

        {/* Actions */}
        <div>
          <SectionLabel>Actions</SectionLabel>
          <div className="flex gap-2">
            {node.status === 'running' ? (
              <>
                <ActionBtn label="Stop" color="red" onClick={() => onNodeAction(node.name, 'stop')} />
                <ActionBtn label="Restart" color="amber" onClick={() => onNodeAction(node.name, 'restart')} />
              </>
            ) : (
              <ActionBtn label="Start" color="green" onClick={() => onNodeAction(node.name, 'start')} />
            )}
          </div>
        </div>

        {/* Labels */}
        {Object.keys(node.labels).length > 0 && (
          <div>
            <SectionLabel>Labels</SectionLabel>
            <div className="space-y-0.5">
              {Object.entries(node.labels).map(([k, v]) => (
                <div key={k} className="font-mono text-2xs">
                  <span className="text-noc-text-dim">{k}=</span>
                  <span className="text-noc-text">{v}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function DetailRow({ label, value, valueClass }: { label: string; value: string; valueClass?: string }) {
  return (
    <div className="flex justify-between items-start gap-3">
      <span className="text-2xs font-mono text-noc-text-dim uppercase tracking-wider shrink-0">
        {label}
      </span>
      <span className={`text-xs font-mono text-right break-all ${valueClass ?? 'text-noc-text'}`}>
        {value}
      </span>
    </div>
  );
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div className="text-2xs font-mono uppercase tracking-widest text-noc-text-dim mb-2 border-b border-noc-border pb-1">
      {children}
    </div>
  );
}

function ActionBtn({ label, color, onClick }: { label: string; color: string; onClick: () => void }) {
  const colorMap: Record<string, string> = {
    red: 'border-noc-red text-noc-red hover:bg-noc-red/10',
    amber: 'border-noc-amber text-noc-amber hover:bg-noc-amber/10',
    green: 'border-noc-green text-noc-green hover:bg-noc-green/10',
  };
  return (
    <button
      onClick={onClick}
      className={`px-3 py-1.5 border rounded text-xs font-mono uppercase tracking-wider
                 transition-colors cursor-pointer ${colorMap[color] ?? ''}`}
    >
      {label}
    </button>
  );
}
