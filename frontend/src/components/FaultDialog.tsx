import React, { useState } from 'react';
import type { LinkInfo, NetemConfig } from '../types/topology';

interface Props {
  link: LinkInfo | null;
  onApply: (linkId: string, params: NetemConfig) => void;
  onCancel: () => void;
}

export function FaultDialog({ link, onApply, onCancel }: Props) {
  const [delayMs, setDelayMs] = useState(link?.netem?.delay_ms ?? 100);
  const [jitterMs, setJitterMs] = useState(link?.netem?.jitter_ms ?? 10);
  const [lossPercent, setLossPercent] = useState(link?.netem?.loss_percent ?? 0);
  const [corruptPercent, setCorruptPercent] = useState(link?.netem?.corrupt_percent ?? 0);
  const [duplicatePercent, setDuplicatePercent] = useState(link?.netem?.duplicate_percent ?? 0);

  if (!link) return null;

  const handleApply = () => {
    // Validate: loss/corrupt/duplicate must be 0-100
    if (lossPercent < 0 || lossPercent > 100) return;
    if (corruptPercent < 0 || corruptPercent > 100) return;
    if (duplicatePercent < 0 || duplicatePercent > 100) return;
    onApply(link.id, {
      delay_ms: delayMs,
      jitter_ms: jitterMs,
      loss_percent: lossPercent,
      corrupt_percent: corruptPercent,
      duplicate_percent: duplicatePercent,
    });
  };

  return (
    <div className="dialog-overlay" onClick={onCancel}>
      <div className="dialog" onClick={e => e.stopPropagation()}>
        <div className="dialog-header">
          <h3>Apply netem</h3>
          <button onClick={onCancel}>&#x2715;</button>
        </div>
        <div className="dialog-body">
          <p className="dialog-link">{link.a.node}:{link.a.interface} &#x2194; {link.z.node}:{link.z.interface}</p>
          <label>
            Delay (ms)
            <input type="number" min="0" value={delayMs} onChange={e => setDelayMs(Number(e.target.value))} />
          </label>
          <label>
            Jitter (ms)
            <input type="number" min="0" value={jitterMs} onChange={e => setJitterMs(Number(e.target.value))} />
          </label>
          <label>
            Packet Loss (%)
            <input type="number" min="0" max="100" step="0.1" value={lossPercent} onChange={e => setLossPercent(Number(e.target.value))} />
          </label>
          <label>
            Corruption (%)
            <input type="number" min="0" max="100" step="0.1" value={corruptPercent} onChange={e => setCorruptPercent(Number(e.target.value))} />
          </label>
          <label>
            Duplication (%)
            <input type="number" min="0" max="100" step="0.1" value={duplicatePercent} onChange={e => setDuplicatePercent(Number(e.target.value))} />
          </label>
        </div>
        <div className="dialog-footer">
          <button onClick={onCancel}>Cancel</button>
          <button onClick={handleApply} className="btn-primary">Apply</button>
        </div>
      </div>
    </div>
  );
}
