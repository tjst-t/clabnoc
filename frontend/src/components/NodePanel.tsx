import React from 'react';
import type { NodeInfo, AccessMethod } from '../types/topology';

interface Props {
  node: NodeInfo | null;
  project: string;
  onOpenTerminal: (node: string, type: 'exec' | 'ssh') => void;
  onClose: () => void;
  onAction?: (node: string, action: 'start' | 'stop' | 'restart') => void;
}

export function NodePanel({ node, project: _project, onOpenTerminal, onClose, onAction }: Props) {
  if (!node) return null;

  const handleAccess = (method: AccessMethod) => {
    if (method.type === 'exec') {
      onOpenTerminal(node.name, 'exec');
    } else if (method.type === 'ssh') {
      onOpenTerminal(node.name, 'ssh');
    } else if (method.type === 'vnc') {
      window.open(method.target, '_blank');
    }
  };

  return (
    <div className="node-panel">
      <div className="node-panel-header">
        <h3>{node.name}</h3>
        <button onClick={onClose}>&#x2715;</button>
      </div>
      <div className="node-panel-body">
        <dl>
          <dt>Kind</dt><dd>{node.kind}</dd>
          <dt>Image</dt><dd>{node.image}</dd>
          <dt>Status</dt><dd className={`status-${node.status}`}>{node.status}</dd>
          <dt>Mgmt IPv4</dt><dd>{node.mgmt_ipv4 || '-'}</dd>
          <dt>Mgmt IPv6</dt><dd>{node.mgmt_ipv6 || '-'}</dd>
          <dt>Container ID</dt><dd>{node.container_id ? node.container_id.slice(0, 12) : '-'}</dd>
        </dl>
        {node.port_bindings.length > 0 && (
          <div className="port-bindings">
            <h4>Port Bindings</h4>
            {node.port_bindings.map((pb, i) => (
              <div key={i}>{pb.host_port}:{pb.port}/{pb.protocol}</div>
            ))}
          </div>
        )}
        <div className="access-methods">
          <h4>Access</h4>
          {node.access_methods.map((m, i) => (
            <button key={i} onClick={() => handleAccess(m)} className={`access-btn ${m.type}`}>
              {m.label}
            </button>
          ))}
        </div>
        {onAction && (
          <div className="node-actions">
            <h4>Actions</h4>
            <button onClick={() => onAction(node.name, 'start')} disabled={node.status === 'running'} className="action-btn start">Start</button>
            <button onClick={() => onAction(node.name, 'stop')} disabled={node.status !== 'running'} className="action-btn stop">Stop</button>
            <button onClick={() => onAction(node.name, 'restart')} className="action-btn restart">Restart</button>
          </div>
        )}
        <div className="node-labels">
          <h4>Labels</h4>
          {Object.entries(node.labels || {}).filter(([k]) => k.startsWith('graph-')).map(([k, v]) => (
            <div key={k}><code>{k}</code>: {v}</div>
          ))}
        </div>
      </div>
    </div>
  );
}
