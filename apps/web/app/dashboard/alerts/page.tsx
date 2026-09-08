'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { AlertTriangle, Filter, Search, Eye } from 'lucide-react';
import { DashboardLayout } from '../../../components/layout/DashboardLayout';
import { SeverityBadge } from '../../../components/ui/SeverityBadge';
import { StatusBadge } from '../../../components/ui/StatusBadge';
import { LoadingState, ErrorState, EmptyState } from '../../../components/ui/States';
import { alertsApi } from '../../../lib/api/alerts';

export default function AlertsPage() {
  const [severityFilter, setSeverityFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [detectionTypeFilter, setDetectionTypeFilter] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState<string>('');

  const {
    data: alerts,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['alerts', severityFilter, statusFilter, detectionTypeFilter],
    queryFn: () =>
      alertsApi.getAlerts({
        severity: severityFilter || undefined,
        status: statusFilter || undefined,
        detection_type: detectionTypeFilter || undefined,
        limit: 100,
      }),
  });

  const filteredAlerts = (alerts || []).filter((alert) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      alert.description.toLowerCase().includes(q) ||
      alert.affected_asset.toLowerCase().includes(q) ||
      alert.source_ip.toLowerCase().includes(q) ||
      alert.detection_type.toLowerCase().includes(q) ||
      alert.id.toLowerCase().includes(q)
    );
  });

  return (
    <DashboardLayout onRefresh={refetch}>
      <div className="space-y-6">
        {/* Header & Filter Controls Bar */}
        <div className="bg-slate-900 border border-slate-800 rounded-lg p-4 space-y-4">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-amber-400" />
              <div>
                <h2 className="text-base font-bold text-slate-100 uppercase tracking-wider">
                  Security Alerts Directory
                </h2>
                <p className="text-xs font-mono text-slate-400">
                  {filteredAlerts.length} Detections Evaluated
                </p>
              </div>
            </div>

            {/* Search Input */}
            <div className="relative w-full sm:w-64">
              <Search className="w-4 h-4 text-slate-500 absolute left-3 top-2.5" />
              <input
                type="text"
                placeholder="Search alerts, IP, asset..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-9 pr-3 py-1.5 bg-slate-950 border border-slate-800 rounded text-xs font-mono text-slate-200 focus:outline-none focus:border-cyan-500 transition"
              />
            </div>
          </div>

          {/* Filter Dropdowns */}
          <div className="flex flex-wrap items-center gap-3 pt-3 border-t border-slate-800 text-xs">
            <div className="flex items-center gap-1.5 text-slate-400 font-mono">
              <Filter className="w-3.5 h-3.5 text-cyan-400" />
              <span>Filters:</span>
            </div>

            {/* Severity Filter */}
            <select
              value={severityFilter}
              onChange={(e) => setSeverityFilter(e.target.value)}
              className="bg-slate-950 border border-slate-800 text-slate-300 font-mono rounded px-2.5 py-1 focus:outline-none focus:border-cyan-500"
            >
              <option value="">All Severities</option>
              <option value="CRITICAL">CRITICAL</option>
              <option value="HIGH">HIGH</option>
              <option value="MEDIUM">MEDIUM</option>
              <option value="LOW">LOW</option>
            </select>

            {/* Status Filter */}
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="bg-slate-950 border border-slate-800 text-slate-300 font-mono rounded px-2.5 py-1 focus:outline-none focus:border-cyan-500"
            >
              <option value="">All Statuses</option>
              <option value="OPEN">OPEN</option>
              <option value="ACKNOWLEDGED">ACKNOWLEDGED</option>
              <option value="RESOLVED">RESOLVED</option>
            </select>

            {/* Detection Type Filter */}
            <select
              value={detectionTypeFilter}
              onChange={(e) => setDetectionTypeFilter(e.target.value)}
              className="bg-slate-950 border border-slate-800 text-slate-300 font-mono rounded px-2.5 py-1 focus:outline-none focus:border-cyan-500"
            >
              <option value="">All Detection Types</option>
              <option value="brute_force">Brute Force (brute_force)</option>
              <option value="unauthorized_access">Unauthorized Access</option>
              <option value="honeytoken">Honeytoken Deception</option>
            </select>

            {(severityFilter || statusFilter || detectionTypeFilter || searchQuery) && (
              <button
                onClick={() => {
                  setSeverityFilter('');
                  setStatusFilter('');
                  setDetectionTypeFilter('');
                  setSearchQuery('');
                }}
                className="text-[11px] font-mono text-cyan-400 hover:underline ml-auto"
              >
                Reset Filters
              </button>
            )}
          </div>
        </div>

        {/* Alerts Table */}
        {isLoading ? (
          <LoadingState message="Loading alert telemetries from PostgreSQL database..." />
        ) : isError ? (
          <ErrorState onRetry={refetch} />
        ) : filteredAlerts.length === 0 ? (
          <EmptyState
            title="No Alerts Found"
            message="No security alerts match the selected criteria."
          />
        ) : (
          <div className="bg-slate-900 border border-slate-800 rounded-lg overflow-hidden shadow-xl">
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-slate-800 text-[11px] font-mono uppercase text-slate-400 bg-slate-950/60">
                    <th className="py-3 px-4">Severity</th>
                    <th className="py-3 px-4">Detection Type</th>
                    <th className="py-3 px-4">Source IP</th>
                    <th className="py-3 px-4">Affected Asset</th>
                    <th className="py-3 px-4">Description</th>
                    <th className="py-3 px-4">Status</th>
                    <th className="py-3 px-4">Timestamp</th>
                    <th className="py-3 px-4 text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60 text-xs">
                  {filteredAlerts.map((alert) => (
                    <tr key={alert.id} className="hover:bg-slate-800/40 transition">
                      <td className="py-3 px-4">
                        <SeverityBadge severity={alert.severity} />
                      </td>
                      <td className="py-3 px-4 font-mono font-medium text-slate-200">
                        {alert.detection_type}
                      </td>
                      <td className="py-3 px-4 font-mono text-slate-400">{alert.source_ip}</td>
                      <td className="py-3 px-4 text-slate-300 font-medium max-w-[150px] truncate">
                        {alert.affected_asset}
                      </td>
                      <td className="py-3 px-4 text-slate-400 max-w-xs truncate">
                        {alert.description}
                      </td>
                      <td className="py-3 px-4">
                        <StatusBadge status={alert.status} />
                      </td>
                      <td className="py-3 px-4 font-mono text-slate-500 text-[11px] whitespace-nowrap">
                        {new Date(alert.timestamp).toLocaleString()}
                      </td>
                      <td className="py-3 px-4 text-right whitespace-nowrap">
                        <Link
                          href={`/dashboard/alerts/${alert.id}`}
                          className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-mono bg-cyan-950 hover:bg-cyan-900 text-cyan-300 rounded border border-cyan-800 transition"
                        >
                          <Eye className="w-3.5 h-3.5" />
                          <span>Investigate</span>
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
