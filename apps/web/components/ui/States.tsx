import React from 'react';
import { AlertCircle, Inbox, Loader2 } from 'lucide-react';

export const LoadingState: React.FC<{ message?: string }> = ({
  message = 'Loading security telemetries...',
}) => {
  return (
    <div className="flex flex-col items-center justify-center p-12 text-slate-400 bg-slate-900/40 rounded-lg border border-slate-800">
      <Loader2 className="w-8 h-8 animate-spin text-cyan-500 mb-3" />
      <p className="text-sm font-mono">{message}</p>
    </div>
  );
};

export const ErrorState: React.FC<{ message?: string; onRetry?: () => void }> = ({
  message = 'Failed to load telemetries from Aegis backend',
  onRetry,
}) => {
  return (
    <div className="flex flex-col items-center justify-center p-8 text-red-400 bg-red-950/30 rounded-lg border border-red-900/50">
      <AlertCircle className="w-8 h-8 mb-2 text-red-400" />
      <p className="text-sm font-mono text-center mb-3">{message}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="px-3 py-1.5 text-xs font-mono bg-red-900/60 hover:bg-red-800 text-red-200 rounded border border-red-700 transition"
        >
          Retry Connection
        </button>
      )}
    </div>
  );
};

export const EmptyState: React.FC<{ title?: string; message?: string }> = ({
  title = 'No Telemetry Records Found',
  message = 'No security events or alerts match the selected criteria.',
}) => {
  return (
    <div className="flex flex-col items-center justify-center p-12 text-slate-500 bg-slate-900/30 rounded-lg border border-slate-800/80">
      <Inbox className="w-10 h-10 mb-3 text-slate-600" />
      <h4 className="text-sm font-semibold text-slate-300 mb-1">{title}</h4>
      <p className="text-xs text-slate-500 font-mono text-center max-w-sm">{message}</p>
    </div>
  );
};
