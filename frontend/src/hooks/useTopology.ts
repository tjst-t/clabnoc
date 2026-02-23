import { useState, useEffect, useCallback } from 'react';
import type { Topology } from '../types/topology';
import { getTopology } from '../lib/api';

export function useTopology(project: string | null) {
  const [topology, setTopology] = useState<Topology | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    if (!project) {
      setTopology(null);
      return;
    }
    try {
      setLoading(true);
      const data = await getTopology(project);
      setTopology(data);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load topology');
    } finally {
      setLoading(false);
    }
  }, [project]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { topology, loading, error, refresh };
}
