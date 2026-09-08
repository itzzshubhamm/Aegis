'use client';

import React from 'react';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { Activity, AlertTriangle, ShieldAlert, Zap, Key, ArrowRight } from 'lucide-react';
import { DashboardLayout } from '../../components/layout/DashboardLayout';
import { StatCard } from '../../components/ui/StatCard';
import { SeverityBadge } from '../../components/ui/SeverityBadge';
import { StatusBadge } from '../../components/ui/StatusBadge';
import { LoadingState, ErrorState, EmptyState } from '../../components/ui/States';
import { alertsApi } from '../../lib/api/alerts';
import { eventsApi } from '../../lib/api/events';
import { honeytokensApi } from '../../lib/api/honeytokens';

export default function DashboardOverviewPage() {
  // Query telemetries from backend
  const {
    data: alerts,
    isLoading: isLoadingAlerts,
    isError: isErrorAlerts,
    refetch: refetchAlerts,
  } = useQuery({
    queryKey: ['alerts'],
    queryFn: () => alertsApi.getAlerts({ limit: 50 }),
  });

  const {
    data: events,
    isLoading: isLoadingEvents,
    isError: isErrorEvents,
    refetch: refetchEvents,
  } = useQuery({
    queryKey: ['events'],
    queryFn: () => eventsApi.getEvents({ limit: 50 }),
  });

  const {
    data: honeytokens,
    isLoading: isLoadingHoneytokens,
    isError: isErrorHoneytokens,
    refetch: refetchHoneytokens,
  } = useQuery({
    queryKey: ['honeytokens'],
    queryFn: () => honeytokensApi.getHoneytokens(),
  });

  const handleRefresh = () => {
    refetchAlerts();
    refetchEvents();
    refetchHoneytokens();
  };

  const isLoading = isLoadingAlerts || isLoadingEvents || isLoadingHoneytokens;
  const isError = isErrorAlerts || isErrorEvents || isErrorHoneytokens;

  // Calculate live metrics
  const totalEvents = events?.length || 0;
  const openAlerts = alerts?.filter((a) => a.status === 'OPEN').length || 0;
  const criticalAlerts = alerts?.filter((a) => a.severity === 'CRITICAL').length || 0;
  const highAlerts = alerts?.filter((a) => a.severity === 'HIGH').length || 0;
  const triggeredHoneytokens = honeytokens?.filter((h) => h.status === 'TRIGGERED').length || 0;

  const recentAlerts = alerts?.slice(0, 5) || [];
  const recentEvents = events?.slice(0, 5) || [];

  return (
    <DashboardLayout onRefresh={handleRefresh}>
      <div className="space-y-6">
        {/* Metric Cards Row */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          <StatCard
            title="Total Events"
            value={totalEvents}
            description="Ingested security telemetries"
            icon={Activity}
            variant="info"
          />
          <StatCard
            title="Open Alerts"
            value={openAlerts}
            description="Requires analyst investigation"
            icon={AlertTriangle}
            variant={openAlerts > 0 ? 'warning' : 'default'}
          />
          <StatCard
            title="Critical Alerts"
            value={criticalAlerts}
            description="Immediate action required"
            icon={ShieldAlert}
            variant={criticalAlerts > 0 ? 'critical' : 'default'}
          />
          <StatCard
            title="High Severity"
            value={highAlerts}
            description="Elevated risk detections"
            icon={Zap}
            variant={highAlerts > 0 ? 'warning' : 'default'}
          />
          <StatCard
            title="Honeytoken Triggers"
            value={triggeredHoneytokens}
            description="Deception mechanism alerts"
            icon={Key}
            variant={triggeredHoneytokens > 0 ? 'critical' : 'default'}
          />
        </div>

        {isLoading ? (
          <LoadingState message="Fetching live telemetries from Go Detection Pipeline..." />
        ) : isError ? (
          <ErrorState onRetry={handleRefresh} />
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Recent Alerts Section */}
            <div className="bg-slate-900 border border-slate-800 rounded-lg p-5">
              <div className="flex items-center justify-between mb-4 border-b border-slate-800 pb-3">
                <div className="flex items-center gap-2">
                  <AlertTriangle className="w-4 h-4 text-amber-400" />
                  <h3 className="text-sm font-bold uppercase tracking-wider text-slate-200">
                    Recent Detections & Alerts
                  </h3>
                </div>
                <Link
                  href="/dashboard/alerts"
                  className="text-xs font-mono text-cyan-400 hover:text-cyan-300 flex items-center gap-1 transition"
                >
                  <span>View All Alerts</span>
                  <ArrowRight className="w-3 h-3" />
                </Link>
              </div>

              {recentAlerts.length === 0 ? (
                <EmptyState title="No Active Alerts" message="No security alerts generated yet." />
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-left border-collapse">
                    <thead>
                      <tr className="border-b border-slate-800 text-[11px] font-mono uppercase text-slate-400">
                        <th className="py-2 px-3">Severity</th>
                        <th className="py-2 px-3">Type</th>
                        <th className="py-2 px-3">Asset</th>
                        <th className="py-2 px-3">Status</th>
                        <th className="py-2 px-3 text-right">Action</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/60 text-xs">
                      {recentAlerts.map((alert) => (
                        <tr key={alert.id} className="hover:bg-slate-800/40 transition">
                          <td className="py-2.5 px-3">
                            <SeverityBadge severity={alert.severity} />
                          </td>
                          <td className="py-2.5 px-3 font-mono font-medium text-slate-200">
                            {alert.detection_type}
                          </td>
                          <td className="py-2.5 px-3 text-slate-400 max-w-[120px] truncate">
                            {alert.affected_asset}
                          </td>
                          <td className="py-2.5 px-3">
                            <StatusBadge status={alert.status} />
                          </td>
                          <td className="py-2.5 px-3 text-right">
                            <Link
                              href={`/dashboard/alerts/${alert.id}`}
                              className="text-[11px] font-mono text-cyan-400 hover:underline"
                            >
                              Investigate
                            </Link>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>

            {/* Recent Telemetry Events Section */}
            <div className="bg-slate-900 border border-slate-800 rounded-lg p-5">
              <div className="flex items-center justify-between mb-4 border-b border-slate-800 pb-3">
                <div className="flex items-center gap-2">
                  <Activity className="w-4 h-4 text-cyan-400" />
                  <h3 className="text-sm font-bold uppercase tracking-wider text-slate-200">
                    Live Security Events Log
                  </h3>
                </div>
                <Link
                  href="/dashboard/events"
                  className="text-xs font-mono text-cyan-400 hover:text-cyan-300 flex items-center gap-1 transition"
                >
                  <span>View Event Stream</span>
                  <ArrowRight className="w-3 h-3" />
                </Link>
              </div>

              {recentEvents.length === 0 ? (
                <EmptyState title="No Events Ingested" message="No security events received yet." />
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-left border-collapse">
                    <thead>
                      <tr className="border-b border-slate-800 text-[11px] font-mono uppercase text-slate-400">
                        <th className="py-2 px-3">Event Type</th>
                        <th className="py-2 px-3">Source IP</th>
                        <th className="py-2 px-3">Resource</th>
                        <th className="py-2 px-3 text-right">Time</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/60 text-xs">
                      {recentEvents.map((evt) => (
                        <tr key={evt.id} className="hover:bg-slate-800/40 transition">
                          <td className="py-2.5 px-3 font-mono font-medium text-slate-200">
                            {evt.event_type}
                          </td>
                          <td className="py-2.5 px-3 font-mono text-slate-400">{evt.source_ip}</td>
                          <td className="py-2.5 px-3 text-slate-400 max-w-[130px] truncate">
                            {evt.resource || evt.source || '-'}
                          </td>
                          <td className="py-2.5 px-3 font-mono text-right text-slate-500 text-[11px]">
                            {new Date(evt.timestamp).toLocaleTimeString()}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}
