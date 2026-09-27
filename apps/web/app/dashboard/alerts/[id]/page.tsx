'use client';

import React, { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, ShieldAlert, CheckCircle2, Clock, FileJson, AlertCircle } from 'lucide-react';
import { DashboardLayout } from '../../../../components/layout/DashboardLayout';
import { SeverityBadge } from '../../../../components/ui/SeverityBadge';
import { StatusBadge } from '../../../../components/ui/StatusBadge';
import { LoadingState, ErrorState } from '../../../../components/ui/States';
import { alertsApi } from '../../../../lib/api/alerts';
import { AlertStatus } from '../../../../lib/types';

export default function AlertDetailPage() {
  const params = useParams();
  const router = useRouter();
  const queryClient = useQueryClient();
  const alertId = params.id as string;

  const [updateFeedback, setUpdateFeedback] = useState<string | null>(null);

  const {
    data: alert,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['alert', alertId],
    queryFn: () => alertsApi.getAlertById(alertId),
    enabled: !!alertId,
  });

  const updateStatusMutation = useMutation({
    mutationFn: (newStatus: AlertStatus) => alertsApi.updateAlertStatus(alertId, newStatus),
    onSuccess: (updatedAlert) => {
      queryClient.setQueryData(['alert', alertId], updatedAlert);
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
      setUpdateFeedback(`Alert status updated to ${updatedAlert.status}`);
      setTimeout(() => setUpdateFeedback(null), 4000);
    },
    onError: (err: any) => {
      const msg = err.response?.data?.error || err.message || 'Failed to update alert status';
      setUpdateFeedback(`Error: ${msg}`);
    },
  });

  if (isLoading) {
    return (
      <DashboardLayout>
        <LoadingState message="Fetching alert details from Aegis database..." />
      </DashboardLayout>
    );
  }

  if (isError || !alert) {
    return (
      <DashboardLayout>
        <ErrorState
          message="Alert not found or access denied due to tenant isolation"
          onRetry={refetch}
        />
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout onRefresh={refetch}>
      <div className="space-y-6 max-w-5xl mx-auto">
        {/* Back Link & Header */}
        <div className="flex items-center justify-between">
          <button
            onClick={() => router.back()}
            className="inline-flex items-center gap-2 px-3 py-1.5 text-xs font-mono text-slate-400 hover:text-slate-200 bg-slate-900 hover:bg-slate-800 rounded border border-slate-800 transition"
          >
            <ArrowLeft className="w-3.5 h-3.5" />
            <span>Back to Alerts List</span>
          </button>

          <div className="flex items-center gap-2">
            <SeverityBadge severity={alert.severity} />
            <StatusBadge status={alert.status} />
          </div>
        </div>

        {/* Feedback Alert */}
        {updateFeedback && (
          <div
            className={`p-3 rounded-lg border text-xs font-mono flex items-center justify-between ${
              updateFeedback.startsWith('Error')
                ? 'bg-red-950/60 border-red-800 text-red-300'
                : 'bg-emerald-950/60 border-emerald-800 text-emerald-300'
            }`}
          >
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4" />
              <span>{updateFeedback}</span>
            </div>
          </div>
        )}

        {/* Alert Overview Card */}
        <div className="bg-slate-900 border border-slate-800 rounded-lg p-6 space-y-6 shadow-xl">
          <div className="flex items-start justify-between gap-4 border-b border-slate-800 pb-4">
            <div>
              <div className="flex items-center gap-2 text-xs font-mono text-cyan-400 uppercase mb-1">
                <ShieldAlert className="w-4 h-4" />
                <span>Detection Type: {alert.detection_type}</span>
              </div>
              <h1 className="text-lg font-bold text-slate-100">{alert.description}</h1>
            </div>
          </div>

          {/* Key Metric Metadata Grid */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 p-4 rounded-lg bg-slate-950/60 border border-slate-800/80 font-mono text-xs">
            <div>
              <span className="text-slate-500 block uppercase text-[10px]">Alert ID</span>
              <span className="text-slate-300 font-medium truncate block">{alert.id}</span>
            </div>
            <div>
              <span className="text-slate-500 block uppercase text-[10px]">Source IP</span>
              <span className="text-slate-200 font-bold">{alert.source_ip}</span>
            </div>
            <div>
              <span className="text-slate-500 block uppercase text-[10px]">Affected Asset</span>
              <span className="text-slate-200 font-medium">{alert.affected_asset}</span>
            </div>
            <div>
              <span className="text-slate-500 block uppercase text-[10px]">Timestamp</span>
              <span className="text-slate-300">{new Date(alert.timestamp).toLocaleString()}</span>
            </div>
          </div>

          {/* Analyst Status Action Controls */}
          <div className="border-t border-b border-slate-800 py-4">
            <h4 className="text-xs font-mono uppercase tracking-wider text-slate-400 mb-3">
              Analyst Incident Workflow Status
            </h4>
            <div className="flex flex-wrap items-center gap-3">
              <button
                onClick={() => updateStatusMutation.mutate('OPEN')}
                disabled={alert.status === 'OPEN' || updateStatusMutation.isPending}
                className={`px-4 py-2 rounded text-xs font-mono font-semibold transition border ${
                  alert.status === 'OPEN'
                    ? 'bg-red-950 border-red-800 text-red-300 cursor-default ring-1 ring-red-700'
                    : 'bg-slate-950 hover:bg-red-950/40 border-slate-800 text-slate-400 hover:text-red-300'
                }`}
              >
                Set OPEN
              </button>

              <button
                onClick={() => updateStatusMutation.mutate('ACKNOWLEDGED')}
                disabled={alert.status === 'ACKNOWLEDGED' || updateStatusMutation.isPending}
                className={`px-4 py-2 rounded text-xs font-mono font-semibold transition border ${
                  alert.status === 'ACKNOWLEDGED'
                    ? 'bg-blue-950 border-blue-800 text-blue-300 cursor-default ring-1 ring-blue-700'
                    : 'bg-slate-950 hover:bg-blue-950/40 border-slate-800 text-slate-400 hover:text-blue-300'
                }`}
              >
                Set ACKNOWLEDGED
              </button>

              <button
                onClick={() => updateStatusMutation.mutate('RESOLVED')}
                disabled={alert.status === 'RESOLVED' || updateStatusMutation.isPending}
                className={`px-4 py-2 rounded text-xs font-mono font-semibold transition border ${
                  alert.status === 'RESOLVED'
                    ? 'bg-emerald-950 border-emerald-800 text-emerald-300 cursor-default ring-1 ring-emerald-700'
                    : 'bg-slate-950 hover:bg-emerald-950/40 border-slate-800 text-slate-400 hover:text-emerald-300'
                }`}
              >
                Set RESOLVED
              </button>
            </div>
          </div>

          {/* JSON Metadata Viewer */}
          <div>
            <div className="flex items-center gap-2 mb-2 text-xs font-mono text-slate-400 uppercase">
              <FileJson className="w-4 h-4 text-cyan-400" />
              <span>Event Context Metadata</span>
            </div>
            <pre className="p-4 rounded-lg bg-slate-950 border border-slate-800 font-mono text-xs text-slate-300 overflow-x-auto">
              {JSON.stringify(alert.metadata || {}, null, 2)}
            </pre>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}
