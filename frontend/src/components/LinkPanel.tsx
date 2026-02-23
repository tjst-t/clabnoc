import type { TopologyLink } from '../types/topology';

interface Props {
  link: TopologyLink;
  onClose: () => void;
  onFaultAction: (linkId: string, action: 'up' | 'down' | 'clear_netem') => void;
  onOpenNetemDialog: (link: TopologyLink) => void;
}

const STATE_DISPLAY: Record<string, { color: string; label: string }> = {
  up: { color: 'text-noc-green', label: 'UP' },
  down: { color: 'text-noc-red', label: 'DOWN' },
  degraded: { color: 'text-noc-amber', label: 'DEGRADED' },
};

export function LinkPanel({ link, onClose, onFaultAction, onOpenNetemDialog }: Props) {
  const state = STATE_DISPLAY[link.state] ?? STATE_DISPLAY.up!;

  return (
    <div className="w-72 h-full bg-noc-bg tui-border-l overflow-y-auto animate-fade-in flex flex-col">
      {/* ─── Header ─── */}
      <div className="tui-border-b px-3 py-1.5 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-2 text-xs">
          <span className={state.color}>{state.label === 'UP' ? '*' : state.label === 'DOWN' ? 'x' : '~'}</span>
          <span className="text-noc-text-bright font-bold">Link</span>
        </div>
        <button onClick={onClose} className="tui-btn tui-btn-dim">
          x
        </button>
      </div>

      <div className="flex-1 overflow-y-auto">
        <div className="px-3 py-2 space-y-3">
          {/* ─── State ─── */}
          <div className="flex items-center gap-2 text-xs">
            <span className="text-noc-text-dim">State:</span>
            <span className={state.color}>[{state.label}]</span>
          </div>

          {/* ─── Endpoints ─── */}
          <TuiSection title="Endpoints">
            <EndpointInfo label="A" node={link.a.node} iface={link.a.interface} mac={link.a.mac} />
            <div className="text-2xs text-noc-text-dim text-center py-0.5">
              {'<'}--{'>'}
            </div>
            <EndpointInfo label="Z" node={link.z.node} iface={link.z.interface} mac={link.z.mac} />
          </TuiSection>

          {/* ─── Host veths ─── */}
          {(link.host_veth_a || link.host_veth_z) && (
            <TuiSection title="Host Interfaces">
              {link.host_veth_a && (
                <div className="text-2xs text-noc-text-dim">A: {link.host_veth_a}</div>
              )}
              {link.host_veth_z && (
                <div className="text-2xs text-noc-text-dim">Z: {link.host_veth_z}</div>
              )}
            </TuiSection>
          )}

          {/* ─── Netem info ─── */}
          {link.netem && (
            <TuiSection title="Active Netem">
              <div className="tui-border p-2 space-y-0.5">
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
            </TuiSection>
          )}

          {/* ─── Fault Injection ─── */}
          <TuiSection title="Fault Injection">
            <div className="flex flex-wrap gap-2">
              {link.state === 'up' ? (
                <>
                  <button
                    onClick={() => onFaultAction(link.id, 'down')}
                    className="tui-btn tui-btn-red"
                  >
                    Link Down
                  </button>
                  <button
                    onClick={() => onOpenNetemDialog(link)}
                    className="tui-btn tui-btn-amber"
                  >
                    Apply Netem
                  </button>
                </>
              ) : link.state === 'down' ? (
                <button
                  onClick={() => onFaultAction(link.id, 'up')}
                  className="tui-btn tui-btn-green"
                >
                  Link Up
                </button>
              ) : (
                <>
                  <button
                    onClick={() => onFaultAction(link.id, 'clear_netem')}
                    className="tui-btn tui-btn-green"
                  >
                    Clear Netem
                  </button>
                  <button
                    onClick={() => onOpenNetemDialog(link)}
                    className="tui-btn tui-btn-amber"
                  >
                    Update Netem
                  </button>
                </>
              )}
            </div>
          </TuiSection>
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

function EndpointInfo({
  label,
  node,
  iface,
  mac,
}: {
  label: string;
  node: string;
  iface: string;
  mac?: string;
}) {
  return (
    <div className="tui-border p-2">
      <div className="text-2xs text-noc-text-dim">{label}-side</div>
      <div className="text-xs text-noc-text-bright">{node}</div>
      <div className="text-2xs text-noc-cyan">{iface}</div>
      {mac && <div className="text-2xs text-noc-text-dim">{mac}</div>}
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
