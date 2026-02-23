import React from 'react';
import type { Project } from '../types/topology';

interface Props {
  projects: Project[];
  selected: string | null;
  onSelect: (name: string) => void;
  loading: boolean;
  error: string | null;
  onRefresh: () => void;
}

export function ProjectSelector({ projects, selected, onSelect, loading, error, onRefresh }: Props) {
  return (
    <div className="project-selector">
      <div className="project-selector-header">
        <span>Projects</span>
        <button onClick={onRefresh} disabled={loading} title="Refresh">&#x21BB;</button>
      </div>
      {error && <div className="error">{error}</div>}
      {loading ? (
        <div className="loading">Loading...</div>
      ) : (
        <ul className="project-list">
          {projects.map(p => (
            <li
              key={p.name}
              className={`project-item ${selected === p.name ? 'active' : ''} status-${p.status}`}
              onClick={() => onSelect(p.name)}
            >
              <span className="project-name">{p.name}</span>
              <span className="project-meta">{p.nodes} nodes</span>
              <span className={`status-badge ${p.status}`}>{p.status}</span>
            </li>
          ))}
          {projects.length === 0 && <li className="empty">No clab projects running</li>}
        </ul>
      )}
    </div>
  );
}
