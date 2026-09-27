'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { Activity, Search, Eye } from 'lucide-react';
import { DashboardLayout } from '../../../components/layout/DashboardLayout';
import { SeverityBadge } from '../../../components/ui/SeverityBadge';
import { LoadingState, ErrorState, EmptyState } from '../../../components/ui/States';
import { eventsApi } from '../../../lib/api/events';

export default function EventsPage() {
  const [searchQuery, setSearchQuery] = useState<string>('');

  const {
    data: events,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['events'],
    queryFn: () => eventsApi.getEvents({ limit: 100 }),
  });

  const filteredEvents = (events || []).filter((evt) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      evt.event_type.toLowerCase().includes(q) ||
      evt.source_ip.toLowerCase().includes(q) ||
      evt.resource.toLowerCase().includes(q) ||
      evt.action.toLowerCase().includes(q) ||
      evt.id.toLowerCase().includes(q)
    );
  });

  return (
    <DashboardLayout onRefresh={refetch}>
      <div className="space-y-6">
        {/* Header Bar */}
        <div className="bg-slate-900 border border-slate-800 rounded-lg p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <Activity className="w-5 h-5 text-cyan-400" />
            <div>
              <h2 className="text-base font-bold text-slate-100 uppercase tracking-wider">
                Security Telemetry Stream
              </h2>
              <p className="text-xs font-mono text-slate-400">
                {filteredEvents.length} Events Ingested
              </p>
            </div>
          </div>

          <div className="relative w-full sm:w-64">
            <Search className="w-4 h-4 text-slate-500 absolute left-3 top-2.5" />
            <input
              type="text"
              placeholder="Search event type, IP, resource..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-9 pr-3 py-1.5 bg-slate-950 border border-slate-800 rounded text-xs font-mono text-slate-200 focus:outline-none focus:border-cyan-500 transition"
            />
          </div>
        </div>

        {/* Events Table */}
        {isLoading ? (
          <LoadingState message="Fetching event telemetry stream from Aegis backend..." />
        ) : isError ? (
          <ErrorState onRetry={refetch} />
        ) : filteredEvents.length === 0 ? (
          <EmptyState
            title="No Security Events Ingested"
            message="No security telemetry events match the current search."
          />
        ) : (
          <div className="bg-slate-900 border border-slate-800 rounded-lg overflow-hidden shadow-xl">
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-slate-800 text-[11px] font-mono uppercase text-slate-400 bg-slate-950/60">
                    <th className="py-3 px-4">Event Type</th>
                    <th className="py-3 px-4">Source</th>
                    <th className="py-3 px-4">Source IP</th>
                    <th className="py-3 px-4">Resource</th>
                    <th className="py-3 px-4">Action</th>
                    <th className="py-3 px-4">Severity</th>
                    <th className="py-3 px-4">Timestamp</th>
                    <th className="py-3 px-4 text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60 text-xs">
                  {filteredEvents.map((evt) => (
                    <tr key={evt.id} className="hover:bg-slate-800/40 transition">
                      <td className="py-3 px-4 font-mono font-semibold text-slate-100">
                        {evt.event_type}
                      </td>
                      <td className="py-3 px-4 text-slate-400 font-mono">{evt.source || '-'}</td>
                      <td className="py-3 px-4 font-mono text-slate-300">{evt.source_ip}</td>
                      <td className="py-3 px-4 text-slate-300 max-w-[150px] truncate">
                        {evt.resource || '-'}
                      </td>
                      <td className="py-3 px-4 font-mono text-slate-400">{evt.action || '-'}</td>
                      <td className="py-3 px-4">
                        <SeverityBadge severity={evt.severity} />
                      </td>
                      <td className="py-3 px-4 font-mono text-slate-500 text-[11px] whitespace-nowrap">
                        {new Date(evt.timestamp).toLocaleString()}
                      </td>
                      <td className="py-3 px-4 text-right whitespace-nowrap">
                        <Link
                          href={`/dashboard/events/${evt.id}`}
                          className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-mono bg-slate-950 hover:bg-slate-800 text-slate-300 rounded border border-slate-700 transition"
                        >
                          <Eye className="w-3.5 h-3.5" />
                          <span>Inspect</span>
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}
