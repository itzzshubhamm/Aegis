'use client';

import React from 'react';
import { usePathname } from 'next/navigation';
import { ShieldAlert, RefreshCw } from 'lucide-react';
import { useAuth } from '../../lib/context/AuthContext';

export const TopBar: React.FC<{ onRefresh?: () => void }> = ({ onRefresh }) => {
  const pathname = usePathname();
  const { tenant } = useAuth();

  const getTitle = () => {
    if (pathname === '/dashboard') return 'Security Overview';
    if (pathname.startsWith('/dashboard/alerts')) return 'Security Alerts';
    if (pathname.startsWith('/dashboard/events')) return 'Security Telemetry Events';
    if (pathname.startsWith('/dashboard/honeytokens')) return 'Deception & Honeytokens';
    return 'SOC Dashboard';
  };

  return (
    <header className="h-16 bg-slate-900/80 backdrop-blur border-b border-slate-800 px-6 flex items-center justify-between sticky top-0 z-10">
      <div className="flex items-center gap-3">
        <h2 className="text-base font-bold text-slate-100 tracking-tight">{getTitle()}</h2>
        <span className="text-xs font-mono text-slate-500 bg-slate-800/80 px-2 py-0.5 rounded">
          LIVE STREAM
        </span>
      </div>

      <div className="flex items-center gap-4">
        {onRefresh && (
          <button
            onClick={onRefresh}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono text-slate-300 hover:text-slate-100 bg-slate-800 hover:bg-slate-700 rounded border border-slate-700 transition"
            title="Refresh Data"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            <span>Refresh</span>
          </button>
        )}

        <div className="flex items-center gap-2 px-2.5 py-1 rounded bg-cyan-950/40 border border-cyan-900/60 text-cyan-400 text-xs font-mono">
          <ShieldAlert className="w-3.5 h-3.5" />
          <span>Tenant: {tenant?.name || 'Isolated'}</span>
        </div>
      </div>
    </header>
  );
};
