import React from 'react';
import { AlertStatus } from '../../lib/types';

interface StatusBadgeProps {
  status: AlertStatus | string;
  className?: string;
}

export const StatusBadge: React.FC<StatusBadgeProps> = ({ status, className = '' }) => {
  const normalized = (status || 'OPEN').toUpperCase();

  let styles = 'bg-slate-800 text-slate-400 border-slate-700';

  switch (normalized) {
    case 'OPEN':
      styles = 'bg-red-950/60 text-red-300 border-red-800/60';
      break;
    case 'ACKNOWLEDGED':
    case 'INVESTIGATING':
      styles = 'bg-blue-950/60 text-blue-300 border-blue-800/60';
      break;
    case 'RESOLVED':
    case 'CLOSED':
    case 'ACTIVE':
      styles = 'bg-emerald-950/60 text-emerald-300 border-emerald-800/60';
      break;
    case 'TRIGGERED':
      styles = 'bg-rose-950 text-rose-300 border-rose-800 animate-pulse font-semibold';
      break;
  }

  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-mono border ${styles} ${className}`}
    >
      {normalized}
    </span>
  );
};
