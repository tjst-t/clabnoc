import { useState, useEffect, useCallback } from 'react';
import { fetchTopology } from '../lib/api';
import type { TopologyData } from '../types/topology';

export function useTopology(project: string | null) {
  const [topology, setTopology] = useState<TopologyData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!project) { setTopology(null); return; }
    setLoading(true);
    setError(null);
    try {
      const data = await fetchTopology(project);
      setTopology(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load topology');
    } finally {
      setLoading(false);
    }
  }, [project]);

  useEffect(() => { load(); }, [load]);

  return { topology, loading, error, refresh: load };
}
