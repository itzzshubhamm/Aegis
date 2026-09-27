'use client';

import React from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, Activity, FileJson } from 'lucide-react';
import { DashboardLayout } from '../../../../components/layout/DashboardLayout';
import { SeverityBadge } from '../../../../components/ui/SeverityBadge';
import { LoadingState, ErrorState } from '../../../../components/ui/States';
import { eventsApi } from '../../../../lib/api/events';

export default function EventDetailPage() {
  const params = useParams();
  const router = useRouter();
  const eventId = params.id as string;

  const {
    data: event,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['event', eventId],
    queryFn: () => eventsApi.getEventById(eventId),
    enabled: !!eventId,
  });

  if (isLoading) {
    return (
      <DashboardLayout>
        <LoadingState message="Fetching security event details..." />
      </DashboardLayout>
    );
  }

  if (isError || !event) {
    return (
      <DashboardLayout>
        <ErrorState
          message="Security event not found or access denied due to tenant isolation"
          onRetry={refetch}
        />
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout onRefresh={refetch}>
      <div className="space-y-6 max-w-4xl mx-auto">
        <div className="flex items-center justify-between">
          <button
            onClick={() => router.back()}
            className="inline-flex items-center gap-2 px-3 py-1.5 text-xs font-mono text-slate-400 hover:text-slate-200 bg-slate-900 hover:bg-slate-800 rounded border border-slate-800 transition"
          >
            <ArrowLeft className="w-3.5 h-3.5" />
            <span>Back to Events Stream</span>
          </button>

          <SeverityBadge severity={event.severity} />
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-lg p-6 space-y-6 shadow-xl">
          <div className="flex items-center gap-3 border-b border-slate-800 pb-4">
            <div className="p-2.5 rounded bg-cyan-950 text-cyan-400 border border-cyan-800">
              <Activity className="w-5 h-5" />
            </div>
            <div>
              <span className="text-[10px] font-mono text-cyan-500 uppercase tracking-widest block">
                Normalized Security Event
              </span>
              <h1 className="text-lg font-bold text-slate-100 font-mono">{event.event_type}</h1>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 p-4 rounded-lg bg-slate-950/60 border border-slate-800/80 font-mono text-xs">
            <div>
              <span className="text-slate-500 block uppercase text-[10px]">Event ID</span>
              <span className="text-slate-300 font-medium truncate block">{event.id}</span>
            </div>
            <div>
              <span className="text-slate-500 block uppercase text-[10px]">Source IP</span>
              <span className="text-slate-200 font-bold">{event.source_ip}</span>
            </div>
            <div>
              <span className="text-slate-500 block uppercase text-[10px]">Resource</span>
              <span className="text-slate-200 font-medium">{event.resource || '-'}</span>
            </div>
            <div>
              <span className="text-slate-500 block uppercase text-[10px]">Action</span>
              <span className="text-slate-300">{event.action || '-'}</span>
            </div>
          </div>

          <div>
            <div className="flex items-center gap-2 mb-2 text-xs font-mono text-slate-400 uppercase">
              <FileJson className="w-4 h-4 text-cyan-400" />
              <span>Full Event Raw JSON Metadata</span>
            </div>
            <pre className="p-4 rounded-lg bg-slate-950 border border-slate-800 font-mono text-xs text-slate-300 overflow-x-auto">
              {JSON.stringify(event, null, 2)}
            </pre>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}
