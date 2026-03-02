import { useRef, useEffect, useState } from 'react';
import type { PacketInfo } from '../types/topology';
import { useCaptureStream } from '../hooks/useCaptureStream';

interface Props {
  project: string;
  linkId: string;
  linkLabel: string;
  onClose: () => void;
}

const PROTOCOL_COLORS: Record<string, string> = {
  TCP: 'text-noc-cyan',
  UDP: 'text-noc-green',
  ICMP: 'text-noc-amber',
  ICMPv6: 'text-noc-amber',
  ARP: 'text-noc-red',
  IP: 'text-noc-text',
};

export function CapturePanel({ project, linkId, linkLabel, onClose }: Props) {
  const { packets, connected, paused, pause, resume, clear } = useCaptureStream(project, linkId);
  const [autoScroll, setAutoScroll] = useState(true);
  const [displayFilter, setDisplayFilter] = useState('');
  const tableRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new packets arrive
  useEffect(() => {
    if (autoScroll && tableRef.current) {
      tableRef.current.scrollTop = tableRef.current.scrollHeight;
    }
  }, [packets.length, autoScroll]);

  const filteredPackets = displayFilter
    ? packets.filter((p) => matchesFilter(p, displayFilter))
    : packets;

  return (
    <div className="flex flex-col h-full bg-noc-bg">
      {/* ─── Header ─── */}
      <div className="tui-border-b px-3 py-1 flex items-center gap-3 shrink-0">
        <div className="flex items-center gap-2 flex-1">
          <span className={`text-2xs ${connected ? 'text-noc-green' : 'text-noc-red'}`}>
            {connected ? '*' : 'x'}
          </span>
          <span className="text-xs text-noc-text-bright font-bold">Capture</span>
          <span className="text-2xs text-noc-text-dim">{linkLabel}</span>
          {connected && !paused && (
            <span className="text-2xs text-noc-red animate-pulse-slow">LIVE</span>
          )}
          {paused && (
            <span className="text-2xs text-noc-amber">PAUSED</span>
          )}
          <span className="text-2xs text-noc-text-dim">
            {filteredPackets.length}
            {displayFilter ? `/${packets.length}` : ''} pkts
          </span>
        </div>
        <div className="flex items-center gap-1">
          {paused ? (
            <button onClick={resume} className="tui-btn tui-btn-green">resume</button>
          ) : (
            <button onClick={pause} className="tui-btn tui-btn-amber">pause</button>
          )}
          <button
            onClick={() => setAutoScroll((v) => !v)}
            className={`tui-btn ${autoScroll ? 'tui-btn-cyan' : 'tui-btn-dim'}`}
          >
            scroll
          </button>
          <button onClick={clear} className="tui-btn tui-btn-dim">clear</button>
          <button onClick={onClose} className="tui-btn tui-btn-dim">x</button>
        </div>
      </div>

      {/* ─── Display Filter ─── */}
      <div className="px-3 py-1 tui-border-b shrink-0">
        <input
          type="text"
          value={displayFilter}
          onChange={(e) => setDisplayFilter(e.target.value)}
          placeholder="display filter... (e.g. TCP, 10.0.0.1, BGP)"
          className="w-full bg-transparent text-2xs text-noc-text-bright
                     placeholder:text-noc-text-dim placeholder:opacity-50
                     focus:outline-none"
        />
      </div>

      {/* ─── Packet Table ─── */}
      <div ref={tableRef} className="flex-1 overflow-auto">
        <table className="w-full text-2xs">
          <thead className="sticky top-0 bg-noc-surface z-10">
            <tr className="tui-border-b">
              <th className="px-2 py-0.5 text-left text-noc-text-dim font-normal w-12">No</th>
              <th className="px-2 py-0.5 text-left text-noc-text-dim font-normal w-28">Time</th>
              <th className="px-2 py-0.5 text-left text-noc-text-dim font-normal">Source</th>
              <th className="px-2 py-0.5 text-left text-noc-text-dim font-normal">Destination</th>
              <th className="px-2 py-0.5 text-left text-noc-text-dim font-normal w-16">Proto</th>
              <th className="px-2 py-0.5 text-right text-noc-text-dim font-normal w-14">Len</th>
              <th className="px-2 py-0.5 text-left text-noc-text-dim font-normal">Info</th>
            </tr>
          </thead>
          <tbody>
            {filteredPackets.map((pkt) => (
              <PacketRow key={pkt.no} packet={pkt} />
            ))}
          </tbody>
        </table>

        {filteredPackets.length === 0 && (
          <div className="flex items-center justify-center h-20 text-2xs text-noc-text-dim">
            {connected
              ? displayFilter
                ? 'No packets match filter'
                : 'Waiting for packets...'
              : 'Disconnected'}
          </div>
        )}
      </div>
    </div>
  );
}

function PacketRow({ packet }: { packet: PacketInfo }) {
  const protoColor = PROTOCOL_COLORS[packet.protocol] ?? 'text-noc-text';

  return (
    <tr className="hover:bg-noc-surface transition-colors border-b"
        style={{ borderColor: 'var(--noc-border)', borderBottomWidth: '0.5px' }}>
      <td className="px-2 py-px text-noc-text-dim">{packet.no}</td>
      <td className="px-2 py-px text-noc-text-dim">{packet.time}</td>
      <td className="px-2 py-px text-noc-text">{packet.source}</td>
      <td className="px-2 py-px text-noc-text">{packet.destination}</td>
      <td className={`px-2 py-px font-bold ${protoColor}`}>{packet.protocol}</td>
      <td className="px-2 py-px text-noc-text-dim text-right">{packet.length}</td>
      <td className="px-2 py-px text-noc-text-dim truncate max-w-xs">{packet.info}</td>
    </tr>
  );
}

function matchesFilter(pkt: PacketInfo, filter: string): boolean {
  const lower = filter.toLowerCase();
  return (
    pkt.source.toLowerCase().includes(lower) ||
    pkt.destination.toLowerCase().includes(lower) ||
    pkt.protocol.toLowerCase().includes(lower) ||
    pkt.info.toLowerCase().includes(lower)
  );
}
