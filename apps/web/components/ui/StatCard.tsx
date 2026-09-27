import React from 'react';
import { LucideIcon } from 'lucide-react';

interface StatCardProps {
  title: string;
  value: number | string;
  description?: string;
  icon: LucideIcon;
  variant?: 'default' | 'critical' | 'warning' | 'success' | 'info';
}

export const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  description,
  icon: Icon,
  variant = 'default',
}) => {
  let borderStyle = 'border-slate-800 bg-slate-900/80';
  let iconBg = 'bg-slate-800 text-slate-300';
  let valueColor = 'text-slate-100';

  if (variant === 'critical') {
    borderStyle = 'border-red-900/60 bg-red-950/20';
    iconBg = 'bg-red-950 text-red-400 border border-red-800';
    valueColor = 'text-red-400';
  } else if (variant === 'warning') {
    borderStyle = 'border-amber-900/60 bg-amber-950/20';
    iconBg = 'bg-amber-950 text-amber-400 border border-amber-800';
    valueColor = 'text-amber-400';
  } else if (variant === 'success') {
    borderStyle = 'border-emerald-900/60 bg-emerald-950/20';
    iconBg = 'bg-emerald-950 text-emerald-400 border border-emerald-800';
    valueColor = 'text-emerald-400';
  } else if (variant === 'info') {
    borderStyle = 'border-blue-900/60 bg-blue-950/20';
    iconBg = 'bg-blue-950 text-blue-400 border border-blue-800';
    valueColor = 'text-blue-400';
  }

  return (
    <div className={`p-4 rounded-lg border ${borderStyle} shadow-sm transition-all`}>
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium uppercase tracking-wider text-slate-400">
          {title}
        </span>
        <div className={`p-2 rounded-md ${iconBg}`}>
          <Icon className="w-4 h-4" />
        </div>
      </div>
      <div className={`mt-3 text-2xl font-bold font-mono ${valueColor}`}>{value}</div>
      {description && (
        <p className="mt-1 text-xs text-slate-500 font-sans">{description}</p>
      )}
    </div>
  );
};
