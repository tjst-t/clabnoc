import { useState, useEffect } from 'react';
import type { SSHCredentials } from '../types/topology';
import { getSSHCredentials } from '../lib/api';

interface Props {
  node: { name: string; kind: string };
  project: string;
  onConnect: (credentials: SSHCredentials) => void;
  onClose: () => void;
}

export function SSHDialog({ node, project, onConnect, onClose }: Props) {
  const [credentials, setCredentials] = useState<SSHCredentials>({
    username: 'admin',
    password: '',
    port: 22,
  });
  const [loading, setLoading] = useState(true);
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getSSHCredentials(project, node.name)
      .then((creds) => {
        if (!cancelled) {
          setCredentials(creds);
          setLoading(false);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err.message);
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [project, node.name]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onConnect(credentials);
    onClose();
  };

  const updateField = <K extends keyof SSHCredentials>(field: K, value: SSHCredentials[K]) => {
    setCredentials((prev) => ({ ...prev, [field]: value }));
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center" onClick={onClose}>
      <div className="absolute inset-0 bg-black/70" />
      <div
        className="relative bg-noc-bg tui-border w-96 animate-fade-in"
        onClick={(e) => e.stopPropagation()}
      >
        {/* ─── Header ─── */}
        <div className="px-3 py-1.5 tui-border-b flex items-center justify-between">
          <div>
            <span className="text-xs text-noc-text-bright font-bold">SSH Connect</span>
            <div className="text-2xs text-noc-text-dim">
              {node.name}
              {node.kind && (
                <span className="text-noc-cyan ml-1">({node.kind})</span>
              )}
            </div>
          </div>
          <button onClick={onClose} className="tui-btn tui-btn-dim">
            x
          </button>
        </div>

        {/* ─── Form ─── */}
        {loading ? (
          <div className="p-3 text-center">
            <div className="text-2xs text-noc-text-dim animate-pulse-slow">
              Loading credentials...
            </div>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="p-3 space-y-3">
            {error && (
              <div className="text-2xs text-noc-amber">
                Could not fetch defaults: {error}
              </div>
            )}

            {/* Username */}
            <div>
              <label htmlFor="ssh-username" className="block text-2xs text-noc-text-dim mb-0.5">
                Username
              </label>
              <input
                id="ssh-username"
                type="text"
                value={credentials.username}
                onChange={(e) => updateField('username', e.target.value)}
                autoFocus
                className="w-full bg-noc-surface tui-border px-2 py-1
                           text-xs text-noc-text-bright
                           focus:outline-none focus:border-noc-cyan"
              />
            </div>

            {/* Password */}
            <div>
              <label htmlFor="ssh-password" className="block text-2xs text-noc-text-dim mb-0.5">
                Password
              </label>
              <div className="relative">
                <input
                  id="ssh-password"
                  type={showPassword ? 'text' : 'password'}
                  value={credentials.password}
                  onChange={(e) => updateField('password', e.target.value)}
                  className="w-full bg-noc-surface tui-border px-2 py-1 pr-12
                             text-xs text-noc-text-bright
                             focus:outline-none focus:border-noc-cyan"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  className="absolute right-1 top-1/2 -translate-y-1/2 tui-btn tui-btn-dim text-2xs"
                >
                  {showPassword ? 'hide' : 'show'}
                </button>
              </div>
            </div>

            {/* Port */}
            <div>
              <label htmlFor="ssh-port" className="block text-2xs text-noc-text-dim mb-0.5">
                Port
              </label>
              <input
                id="ssh-port"
                type="number"
                min={1}
                max={65535}
                value={credentials.port}
                onChange={(e) => {
                  const num = parseInt(e.target.value, 10);
                  if (!isNaN(num) && num >= 1 && num <= 65535) {
                    updateField('port', num);
                  }
                }}
                className="w-full bg-noc-surface tui-border px-2 py-1
                           text-xs text-noc-text-bright
                           focus:outline-none focus:border-noc-cyan
                           [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
              />
            </div>

            {/* Preview */}
            <div className="tui-border p-2">
              <div className="text-2xs text-noc-text-dim mb-1">--- Preview ---</div>
              <code className="text-2xs text-noc-cyan block">
                ssh {credentials.username}@{node.name}
                {credentials.port !== 22 && ` -p ${credentials.port}`}
              </code>
            </div>

            {/* Buttons */}
            <div className="flex justify-end gap-2">
              <button type="button" onClick={onClose} className="tui-btn tui-btn-dim">
                Cancel
              </button>
              <button type="submit" className="tui-btn tui-btn-cyan">
                Connect
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
