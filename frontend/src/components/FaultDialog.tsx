import { useState, useEffect } from 'react';
import type { TopologyLink, NetemParams, BPFPreset } from '../types/topology';
import { getBPFPresets } from '../lib/api';

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
  const [bpfFilter, setBpfFilter] = useState(link.netem?.bpf_filter ?? '');
  const [presets, setPresets] = useState<BPFPreset[]>([]);

  useEffect(() => {
    getBPFPresets().then(setPresets).catch(() => {});
  }, []);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const finalParams = { ...params, bpf_filter: bpfFilter || undefined };
    onApply(link.id, finalParams);
    onClose();
  };

  const update = (field: keyof NetemParams, value: string) => {
    const num = parseInt(value, 10);
    if (!isNaN(num) && num >= 0) {
      setParams((p) => ({ ...p, [field]: num }));
    }
  };

  const handlePresetChange = (value: string) => {
    if (value === '') {
      setBpfFilter('');
      return;
    }
    const preset = presets.find((p) => p.name === value);
    if (preset) {
      setBpfFilter(preset.filter);
    }
  };

  const hasBpf = bpfFilter.trim().length > 0;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center" onClick={onClose}>
      <div className="absolute inset-0" style={{ background: 'var(--noc-modal-overlay)' }} />
      <div
        className="relative bg-noc-bg tui-border w-96 animate-fade-in"
        onClick={(e) => e.stopPropagation()}
      >
        {/* ─── Header ─── */}
        <div className="px-3 py-1.5 tui-border-b flex items-center justify-between">
          <div>
            <span className="text-xs text-noc-text-bright font-bold">Network Emulation</span>
            <div className="text-2xs text-noc-text-dim">
              {link.a.node}:{link.a.interface} &lt;-&gt; {link.z.node}:{link.z.interface}
            </div>
          </div>
          <button onClick={onClose} className="tui-btn tui-btn-dim">
            x
          </button>
        </div>

        {/* ─── Form ─── */}
        <form onSubmit={handleSubmit} className="p-3 space-y-3">
          <div className="grid grid-cols-2 gap-2">
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

          {/* ─── BPF Filter ─── */}
          <div className="tui-border p-2 space-y-2">
            <div className="text-2xs text-noc-text-dim">--- BPF Filter (selective) ---</div>
            <div>
              <label htmlFor="bpf-preset" className="block text-2xs text-noc-text-dim mb-0.5">
                Preset
              </label>
              <select
                id="bpf-preset"
                value={presets.find((p) => p.filter === bpfFilter)?.name ?? ''}
                onChange={(e) => handlePresetChange(e.target.value)}
                className="w-full bg-noc-surface tui-border px-2 py-1
                           text-xs text-noc-text-bright
                           focus:outline-none focus:border-noc-cyan"
              >
                <option value="">-- none (all traffic) --</option>
                {presets.map((p) => (
                  <option key={p.name} value={p.name}>
                    {p.name} ({p.filter})
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label htmlFor="bpf-custom" className="block text-2xs text-noc-text-dim mb-0.5">
                Custom Filter
              </label>
              <input
                id="bpf-custom"
                type="text"
                value={bpfFilter}
                onChange={(e) => setBpfFilter(e.target.value)}
                placeholder="e.g. tcp port 179"
                className="w-full bg-noc-surface tui-border px-2 py-1
                           text-xs text-noc-text-bright
                           placeholder:text-noc-text-dim placeholder:opacity-50
                           focus:outline-none focus:border-noc-cyan"
              />
            </div>
            {hasBpf && (
              <div className="text-2xs text-noc-cyan">
                Only matching traffic will be affected by netem
              </div>
            )}
          </div>

          {/* Preview */}
          <div className="tui-border p-2">
            <div className="text-2xs text-noc-text-dim mb-1">--- Preview ---</div>
            {hasBpf ? (
              <div className="space-y-0.5">
                <code className="text-2xs text-noc-amber block">
                  tc qdisc add dev veth root handle 1: prio bands 3
                </code>
                <code className="text-2xs text-noc-amber block">
                  tc qdisc add dev veth parent 1:1 netem
                  {params.delay_ms > 0 && ` delay ${params.delay_ms}ms`}
                  {params.jitter_ms > 0 && ` ${params.jitter_ms}ms`}
                  {params.loss_percent > 0 && ` loss ${params.loss_percent}%`}
                  {params.corrupt_percent > 0 && ` corrupt ${params.corrupt_percent}%`}
                  {params.duplicate_percent > 0 && ` duplicate ${params.duplicate_percent}%`}
                </code>
                <code className="text-2xs text-noc-cyan block">
                  tc filter add dev veth ... match &quot;{bpfFilter}&quot; flowid 1:1
                </code>
              </div>
            ) : (
              <code className="text-2xs text-noc-amber block">
                tc qdisc add dev veth root netem
                {params.delay_ms > 0 && ` delay ${params.delay_ms}ms`}
                {params.jitter_ms > 0 && ` ${params.jitter_ms}ms`}
                {params.loss_percent > 0 && ` loss ${params.loss_percent}%`}
                {params.corrupt_percent > 0 && ` corrupt ${params.corrupt_percent}%`}
                {params.duplicate_percent > 0 && ` duplicate ${params.duplicate_percent}%`}
              </code>
            )}
          </div>

          {/* Buttons */}
          <div className="flex justify-end gap-2">
            <button type="button" onClick={onClose} className="tui-btn tui-btn-dim">
              Cancel
            </button>
            <button type="submit" className="tui-btn tui-btn-amber">
              Apply
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
      <label htmlFor={id} className="block text-2xs text-noc-text-dim mb-0.5">
        {label}
      </label>
      <input
        id={id}
        type="number"
        min={0}
        max={max}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full bg-noc-surface tui-border px-2 py-1
                   text-xs text-noc-text-bright
                   focus:outline-none focus:border-noc-amber
                   [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
      />
    </div>
  );
}
