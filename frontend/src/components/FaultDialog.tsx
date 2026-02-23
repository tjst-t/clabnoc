import { useState } from 'react';
import type { TopologyLink, NetemParams } from '../types/topology';

interface Props {
  link: TopologyLink;
  onApply: (linkId: string, params: NetemParams) => void;
  onClose: () => void;
}

export function FaultDialog({ link, onApply, onClose }: Props) {
  const [params, setParams] = useState<NetemParams>(
    link.netem ?? {
      delay_ms: 0,
      jitter_ms: 0,
      loss_percent: 0,
      corrupt_percent: 0,
      duplicate_percent: 0,
    }
  );

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onApply(link.id, params);
    onClose();
  };

  const update = (field: keyof NetemParams, value: string) => {
    const num = parseInt(value, 10);
    if (!isNaN(num) && num >= 0) {
      setParams((p) => ({ ...p, [field]: num }));
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center" onClick={onClose}>
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />
      <div
        className="relative bg-noc-panel border border-noc-border rounded-lg shadow-2xl w-96 animate-fade-in"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="px-5 py-4 border-b border-noc-border flex items-center justify-between">
          <div>
            <h3 className="font-display text-sm font-semibold text-noc-text-bright">
              Network Emulation
            </h3>
            <p className="text-2xs font-mono text-noc-text-dim mt-0.5">
              {link.a.node}:{link.a.interface} ↔ {link.z.node}:{link.z.interface}
            </p>
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

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <ParamInput
              label="Delay (ms)"
              value={params.delay_ms}
              onChange={(v) => update('delay_ms', v)}
              max={10000}
            />
            <ParamInput
              label="Jitter (ms)"
              value={params.jitter_ms}
              onChange={(v) => update('jitter_ms', v)}
              max={5000}
            />
            <ParamInput
              label="Loss (%)"
              value={params.loss_percent}
              onChange={(v) => update('loss_percent', v)}
              max={100}
            />
            <ParamInput
              label="Corrupt (%)"
              value={params.corrupt_percent}
              onChange={(v) => update('corrupt_percent', v)}
              max={100}
            />
            <ParamInput
              label="Duplicate (%)"
              value={params.duplicate_percent}
              onChange={(v) => update('duplicate_percent', v)}
              max={100}
            />
          </div>

          {/* Preview */}
          <div className="bg-noc-bg border border-noc-border rounded p-3">
            <div className="text-2xs font-mono text-noc-text-dim uppercase tracking-wider mb-1.5">
              Preview
            </div>
            <code className="text-xs font-mono text-noc-amber block">
              tc qdisc add dev veth root netem
              {params.delay_ms > 0 && ` delay ${params.delay_ms}ms`}
              {params.jitter_ms > 0 && ` ${params.jitter_ms}ms`}
              {params.loss_percent > 0 && ` loss ${params.loss_percent}%`}
              {params.corrupt_percent > 0 && ` corrupt ${params.corrupt_percent}%`}
              {params.duplicate_percent > 0 && ` duplicate ${params.duplicate_percent}%`}
            </code>
          </div>

          {/* Buttons */}
          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 border border-noc-border text-noc-text-dim rounded text-xs font-mono
                         uppercase tracking-wider hover:border-noc-text-dim transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-noc-amber/20 border border-noc-amber text-noc-amber rounded text-xs font-mono
                         uppercase tracking-wider hover:bg-noc-amber/30 transition-colors cursor-pointer"
            >
              Apply Netem
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function ParamInput({
  label,
  value,
  onChange,
  max,
}: {
  label: string;
  value: number;
  onChange: (v: string) => void;
  max: number;
}) {
  const id = `netem-${label.toLowerCase().replace(/[^a-z0-9]/g, '-')}`;
  return (
    <div>
      <label htmlFor={id} className="block text-2xs font-mono text-noc-text-dim uppercase tracking-wider mb-1">
        {label}
      </label>
      <input
        id={id}
        type="number"
        min={0}
        max={max}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full bg-noc-bg border border-noc-border rounded px-2.5 py-1.5
                   font-mono text-sm text-noc-text-bright
                   focus:outline-none focus:border-noc-amber
                   [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
      />
    </div>
  );
}
