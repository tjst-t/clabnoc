import type { TopologyLink } from '../types/topology';

interface Props {
  link: TopologyLink;
  onClose: () => void;
  onFaultAction: (linkId: string, action: 'up' | 'down' | 'clear_netem') => void;
  onOpenNetemDialog: (link: TopologyLink) => void;
}

const STATE_DISPLAY: Record<string, { color: string; label: string; dot: string }> = {
  up: { color: 'text-noc-green', label: 'UP', dot: 'bg-noc-green' },
  down: { color: 'text-noc-red', label: 'DOWN', dot: 'bg-noc-red' },
  degraded: { color: 'text-noc-amber', label: 'DEGRADED', dot: 'bg-noc-amber' },
};

export function LinkPanel({ link, onClose, onFaultAction, onOpenNetemDialog }: Props) {
  const state = STATE_DISPLAY[link.state] ?? STATE_DISPLAY.up!;

  return (
    <div className="w-80 h-full bg-noc-panel border-l border-noc-border overflow-y-auto animate-fade-in">
      {/* Header */}
      <div className="sticky top-0 bg-noc-panel border-b border-noc-border px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${state.dot}`} />
          <h2 className="font-mono text-sm font-semibold text-noc-text-bright">Link</h2>
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
        {/* State */}
        <div className="flex items-center gap-2">
          <span className={`px-2 py-0.5 border rounded text-2xs font-mono uppercase tracking-wider
            ${link.state === 'up' ? 'border-noc-green text-noc-green' :
              link.state === 'down' ? 'border-noc-red text-noc-red' :
              'border-noc-amber text-noc-amber'}`}>
            {state.label}
          </span>
        </div>

        {/* Endpoints */}
        <div className="space-y-3">
          <EndpointInfo label="A-side" node={link.a.node} iface={link.a.interface} mac={link.a.mac} />
          <div className="flex items-center gap-2 px-2">
            <div className="flex-1 border-t border-noc-border border-dashed" />
            <span className="text-2xs text-noc-text-dim font-mono">↔</span>
            <div className="flex-1 border-t border-noc-border border-dashed" />
          </div>
          <EndpointInfo label="Z-side" node={link.z.node} iface={link.z.interface} mac={link.z.mac} />
        </div>

        {/* Host veths */}
        {(link.host_veth_a || link.host_veth_z) && (
          <div>
            <SectionLabel>Host Interfaces</SectionLabel>
            {link.host_veth_a && <div className="font-mono text-2xs text-noc-text-dim">A: {link.host_veth_a}</div>}
            {link.host_veth_z && <div className="font-mono text-2xs text-noc-text-dim">Z: {link.host_veth_z}</div>}
          </div>
        )}

        {/* Netem info */}
        {link.netem && (
          <div>
            <SectionLabel>Active Netem</SectionLabel>
            <div className="bg-noc-surface border border-noc-amber/30 rounded p-2 space-y-0.5">
              {link.netem.delay_ms > 0 && (
                <NetemRow label="Delay" value={`${link.netem.delay_ms}ms`} />
              )}
              {link.netem.jitter_ms > 0 && (
                <NetemRow label="Jitter" value={`${link.netem.jitter_ms}ms`} />
              )}
              {link.netem.loss_percent > 0 && (
                <NetemRow label="Loss" value={`${link.netem.loss_percent}%`} />
              )}
              {link.netem.corrupt_percent > 0 && (
                <NetemRow label="Corrupt" value={`${link.netem.corrupt_percent}%`} />
              )}
              {link.netem.duplicate_percent > 0 && (
                <NetemRow label="Duplicate" value={`${link.netem.duplicate_percent}%`} />
              )}
            </div>
          </div>
        )}

        {/* Fault injection actions */}
        <div>
          <SectionLabel>Fault Injection</SectionLabel>
          <div className="space-y-2">
            {link.state === 'up' ? (
              <>
                <button
                  onClick={() => onFaultAction(link.id, 'down')}
                  className="w-full px-3 py-2 border border-noc-red text-noc-red rounded text-xs font-mono
                             uppercase tracking-wider hover:bg-noc-red/10 transition-colors cursor-pointer"
                >
                  Link Down
                </button>
                <button
                  onClick={() => onOpenNetemDialog(link)}
                  className="w-full px-3 py-2 border border-noc-amber text-noc-amber rounded text-xs font-mono
                             uppercase tracking-wider hover:bg-noc-amber/10 transition-colors cursor-pointer"
                >
                  Apply Netem
                </button>
              </>
            ) : link.state === 'down' ? (
              <button
                onClick={() => onFaultAction(link.id, 'up')}
                className="w-full px-3 py-2 border border-noc-green text-noc-green rounded text-xs font-mono
                           uppercase tracking-wider hover:bg-noc-green/10 transition-colors cursor-pointer"
              >
                Link Up (Restore)
              </button>
            ) : (
              <>
                <button
                  onClick={() => onFaultAction(link.id, 'clear_netem')}
                  className="w-full px-3 py-2 border border-noc-green text-noc-green rounded text-xs font-mono
                             uppercase tracking-wider hover:bg-noc-green/10 transition-colors cursor-pointer"
                >
                  Clear Netem
                </button>
                <button
                  onClick={() => onOpenNetemDialog(link)}
                  className="w-full px-3 py-2 border border-noc-amber text-noc-amber rounded text-xs font-mono
                             uppercase tracking-wider hover:bg-noc-amber/10 transition-colors cursor-pointer"
                >
                  Update Netem
                </button>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function EndpointInfo({ label, node, iface, mac }: {
  label: string; node: string; iface: string; mac?: string;
}) {
  return (
    <div className="bg-noc-surface border border-noc-border rounded p-2.5">
      <div className="text-2xs font-mono text-noc-text-dim uppercase tracking-wider mb-1">{label}</div>
      <div className="font-mono text-sm text-noc-text-bright">{node}</div>
      <div className="font-mono text-xs text-noc-cyan mt-0.5">{iface}</div>
      {mac && <div className="font-mono text-2xs text-noc-text-dim mt-0.5">{mac}</div>}
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

function NetemRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between font-mono text-2xs">
      <span className="text-noc-amber">{label}</span>
      <span className="text-noc-text-bright">{value}</span>
    </div>
  );
}
