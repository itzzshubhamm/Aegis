import React from 'react';
import { AlertSeverity } from '../../lib/types';

interface SeverityBadgeProps {
  severity: AlertSeverity | string;
  className?: string;
}

export const SeverityBadge: React.FC<SeverityBadgeProps> = ({ severity, className = '' }) => {
  const normalized = (severity || 'INFO').toUpperCase();

  let styles = 'bg-slate-800 text-slate-300 border-slate-700';

  switch (normalized) {
    case 'CRITICAL':
      styles = 'bg-red-950/80 text-red-400 border-red-800/80 font-bold animate-pulse';
      break;
    case 'HIGH':
      styles = 'bg-orange-950/80 text-orange-400 border-orange-800/80 font-semibold';
      break;
    case 'MEDIUM':
      styles = 'bg-amber-950/70 text-amber-300 border-amber-800/70';
      break;
    case 'LOW':
      styles = 'bg-blue-950/70 text-blue-300 border-blue-800/70';
      break;
    case 'INFO':
      styles = 'bg-slate-900 text-slate-400 border-slate-800';
      break;
  }

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-mono border ${styles} ${className}`}
    >
      {normalized}
    </span>
  );
};
