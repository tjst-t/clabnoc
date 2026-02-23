import type { ProjectInfo } from '../types/topology';

interface Props {
  projects: ProjectInfo[];
  selected: string | null;
  onSelect: (name: string) => void;
  loading: boolean;
}

export function ProjectSelector({ projects, selected, onSelect, loading }: Props) {
  return (
    <div className="flex items-center gap-2 text-xs">
      <span className="text-noc-text-dim">Project:</span>
      <select
        value={selected ?? ''}
        onChange={(e) => onSelect(e.target.value)}
        disabled={loading}
        className="appearance-none bg-transparent border-none
                   text-noc-text-bright text-xs
                   focus:outline-none
                   disabled:opacity-50 cursor-pointer
                   pr-4"
      >
        <option value="">-- select --</option>
        {projects.map((p) => (
          <option key={p.name} value={p.name}>
            {p.name}
          </option>
        ))}
      </select>
      <span className="text-noc-text-dim">▼</span>

      {selected && (
        <div className="flex items-center gap-2 animate-fade-in">
          {projects
            .filter((p) => p.name === selected)
            .map((p) => (
              <div key={p.name} className="flex items-center gap-2 text-2xs">
                <span className="text-noc-text-dim">{p.nodes} nodes</span>
                <span
                  className={
                    p.status === 'running'
                      ? 'text-noc-green'
                      : p.status === 'partial'
                        ? 'text-noc-amber'
                        : 'text-noc-red'
                  }
                >
                  [{p.status}]
                </span>
              </div>
            ))}
        </div>
      )}
    </div>
  );
}
