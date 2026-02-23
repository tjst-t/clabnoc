import React from 'react';
import type { LinkInfo } from '../types/topology';

interface Props {
  link: LinkInfo | null;
  project: string;
  onFaultDown: (linkId: string) => void;
  onFaultUp: (linkId: string) => void;
  onFaultNetem: (link: LinkInfo) => void;
  onClose: () => void;
}

export function LinkPanel({ link, project: _project, onFaultDown, onFaultUp, onFaultNetem, onClose }: Props) {
  if (!link) return null;

  const stateColor = { up: '#4CAF50', down: '#F44336', degraded: '#FFC107' }[link.state] || '#999';

  return (
    <div className="link-panel">
      <div className="link-panel-header">
        <h3>Link</h3>
        <button onClick={onClose}>&#x2715;</button>
      </div>
      <div className="link-panel-body">
        <div className="link-endpoints">
          <span className="endpoint">{link.a.node}:{link.a.interface}</span>
          <span className="link-arrow">&#x2194;</span>
          <span className="endpoint">{link.z.node}:{link.z.interface}</span>
        </div>
        <div className="link-state" style={{ color: stateColor }}>
          State: <strong>{link.state}</strong>
        </div>
        {link.netem && (
          <div className="netem-info">
            <h4>netem</h4>
            {link.netem.delay_ms > 0 && <div>Delay: {link.netem.delay_ms}ms &#177;{link.netem.jitter_ms}ms</div>}
            {link.netem.loss_percent > 0 && <div>Loss: {link.netem.loss_percent}%</div>}
            {link.netem.corrupt_percent > 0 && <div>Corrupt: {link.netem.corrupt_percent}%</div>}
          </div>
        )}
        <div className="fault-actions">
          <h4>Fault Injection</h4>
          <button onClick={() => onFaultDown(link.id)} disabled={link.state === 'down'} className="fault-btn down">
            Link Down
          </button>
          <button onClick={() => onFaultUp(link.id)} disabled={link.state === 'up'} className="fault-btn up">
            Link Up (Recover)
          </button>
          <button onClick={() => onFaultNetem(link)} className="fault-btn netem">
            Apply netem...
          </button>
          {link.state === 'degraded' && (
            <button onClick={() => onFaultUp(link.id)} className="fault-btn clear">
              Clear netem
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
