import type { ProjectInfo } from '../types/topology';

interface Props {
  projects: ProjectInfo[];
  selected: string | null;
  onSelect: (name: string) => void;
  loading: boolean;
}

const STATUS_DOT: Record<string, string> = {
  running: 'bg-noc-green',
  partial: 'bg-noc-amber',
  stopped: 'bg-noc-red',
};

export function ProjectSelector({ projects, selected, onSelect, loading }: Props) {
  return (
    <div className="flex items-center gap-3">
      <span className="text-2xs font-mono uppercase tracking-widest text-noc-text-dim">
        Project
      </span>
      <div className="relative">
        <select
          value={selected ?? ''}
          onChange={(e) => onSelect(e.target.value)}
          disabled={loading}
          className="appearance-none bg-noc-surface border border-noc-border rounded px-3 py-1.5 pr-8
                     font-mono text-sm text-noc-text-bright
                     focus:outline-none focus:border-noc-accent
                     disabled:opacity-50 cursor-pointer
                     hover:border-noc-border-active transition-colors"
        >
          <option value="">— select —</option>
          {projects.map((p) => (
            <option key={p.name} value={p.name}>
              {p.name}
            </option>
          ))}
        </select>
        <div className="absolute right-2 top-1/2 -translate-y-1/2 pointer-events-none">
          <svg width="10" height="6" viewBox="0 0 10 6" fill="none">
            <path d="M1 1L5 5L9 1" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
        </div>
      </div>

      {selected && (
        <div className="flex items-center gap-2 animate-fade-in">
          {projects
            .filter((p) => p.name === selected)
            .map((p) => (
              <div key={p.name} className="flex items-center gap-2 font-mono text-2xs">
                <span className={`w-1.5 h-1.5 rounded-full ${STATUS_DOT[p.status] ?? 'bg-noc-text-dim'}`} />
                <span className="text-noc-text-dim">{p.nodes} nodes</span>
                <span className="text-noc-text-dim">·</span>
                <span className={
                  p.status === 'running' ? 'text-noc-green' :
                  p.status === 'partial' ? 'text-noc-amber' : 'text-noc-red'
                }>
                  {p.status}
                </span>
              </div>
            ))}
        </div>
      )}
    </div>
  );
}
