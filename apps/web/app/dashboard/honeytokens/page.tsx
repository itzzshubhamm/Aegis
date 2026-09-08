'use client';

import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Key, Plus, Zap, CheckCircle2, AlertCircle, ShieldAlert, X } from 'lucide-react';
import { DashboardLayout } from '../../../components/layout/DashboardLayout';
import { StatusBadge } from '../../../components/ui/StatusBadge';
import { LoadingState, ErrorState, EmptyState } from '../../../components/ui/States';
import { honeytokensApi } from '../../../lib/api/honeytokens';

export default function HoneytokensPage() {
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [name, setName] = useState('');
  const [type, setType] = useState('API_KEY');
  const [tokenValue, setTokenValue] = useState('');
  const [actionFeedback, setActionFeedback] = useState<string | null>(null);

  const {
    data: honeytokens,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['honeytokens'],
    queryFn: () => honeytokensApi.getHoneytokens(),
  });

  const createMutation = useMutation({
    mutationFn: () => honeytokensApi.createHoneytoken({ name, type, token_value: tokenValue }),
    onSuccess: (newToken) => {
      queryClient.invalidateQueries({ queryKey: ['honeytokens'] });
      setIsModalOpen(false);
      setName('');
      setTokenValue('');
      setActionFeedback(`Honeytoken '${newToken.name}' created successfully.`);
      setTimeout(() => setActionFeedback(null), 4000);
    },
    onError: (err: any) => {
      const msg = err.response?.data?.error || err.message || 'Failed to create honeytoken';
      setActionFeedback(`Error: ${msg}`);
    },
  });

  const triggerMutation = useMutation({
    mutationFn: (id: string) => honeytokensApi.triggerHoneytoken(id),
    onSuccess: (res) => {
      queryClient.invalidateQueries({ queryKey: ['honeytokens'] });
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
      queryClient.invalidateQueries({ queryKey: ['events'] });
      setActionFeedback(`Deception Honeytoken Triggered! Event dispatched to Kafka stream (Event ID: ${res.event_id}). Detection engine will evaluate.`);
      setTimeout(() => setActionFeedback(null), 5000);
    },
    onError: (err: any) => {
      const msg = err.response?.data?.error || err.message || 'Failed to trigger honeytoken';
      setActionFeedback(`Error: ${msg}`);
    },
  });

  const handleCreateSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    createMutation.mutate();
  };

  return (
    <DashboardLayout onRefresh={refetch}>
      <div className="space-y-6">
        {/* Top Header & Actions Bar */}
        <div className="bg-slate-900 border border-slate-800 rounded-lg p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <Key className="w-5 h-5 text-cyan-400" />
            <div>
              <h2 className="text-base font-bold text-slate-100 uppercase tracking-wider">
                Deception Honeytokens
              </h2>
              <p className="text-xs font-mono text-slate-400">
                {(honeytokens || []).length} Fake Deception Credentials Configured
              </p>
            </div>
          </div>

          <button
            onClick={() => setIsModalOpen(true)}
            className="flex items-center gap-2 px-3 py-2 text-xs font-mono font-semibold bg-cyan-600 hover:bg-cyan-500 text-slate-950 rounded shadow-md transition"
          >
            <Plus className="w-4 h-4" />
            <span>Create Honeytoken</span>
          </button>
        </div>

        {/* Action Feedback Notification Banner */}
        {actionFeedback && (
          <div
            className={`p-3 rounded-lg border text-xs font-mono flex items-center justify-between ${
              actionFeedback.startsWith('Error')
                ? 'bg-red-950/60 border-red-800 text-red-300'
                : 'bg-cyan-950/60 border-cyan-800 text-cyan-300'
            }`}
          >
            <div className="flex items-center gap-2">
              {actionFeedback.startsWith('Error') ? (
                <AlertCircle className="w-4 h-4 shrink-0 text-red-400" />
              ) : (
                <CheckCircle2 className="w-4 h-4 shrink-0 text-cyan-400" />
              )}
              <span>{actionFeedback}</span>
            </div>
          </div>
        )}

        {/* Honeytokens Table */}
        {isLoading ? (
          <LoadingState message="Loading deception honeytokens from Aegis backend..." />
        ) : isError ? (
          <ErrorState onRetry={refetch} />
        ) : (honeytokens || []).length === 0 ? (
          <EmptyState
            title="No Honeytokens Deployed"
            message="Deploy honeytoken fake credentials to trap unauthorized attackers."
          />
        ) : (
          <div className="bg-slate-900 border border-slate-800 rounded-lg overflow-hidden shadow-xl">
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-slate-800 text-[11px] font-mono uppercase text-slate-400 bg-slate-950/60">
                    <th className="py-3 px-4">Name</th>
                    <th className="py-3 px-4">Type</th>
                    <th className="py-3 px-4">Token Value</th>
                    <th className="py-3 px-4">Status</th>
                    <th className="py-3 px-4">Last Triggered</th>
                    <th className="py-3 px-4">Created At</th>
                    <th className="py-3 px-4 text-right">Testing Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60 text-xs">
                  {honeytokens?.map((token) => (
                    <tr key={token.id} className="hover:bg-slate-800/40 transition">
                      <td className="py-3 px-4 font-semibold text-slate-100">{token.name}</td>
                      <td className="py-3 px-4">
                        <span className="inline-block px-2 py-0.5 rounded text-[10px] font-mono bg-slate-800 text-cyan-300 border border-slate-700">
                          {token.type}
                        </span>
                      </td>
                      <td className="py-3 px-4 font-mono text-slate-400 max-w-[180px] truncate">
                        {token.token_value}
                      </td>
                      <td className="py-3 px-4">
                        <StatusBadge status={token.status} />
                      </td>
                      <td className="py-3 px-4 font-mono text-slate-400 text-[11px]">
                        {token.last_triggered_at
                          ? new Date(token.last_triggered_at).toLocaleString()
                          : 'Never'}
                      </td>
                      <td className="py-3 px-4 font-mono text-slate-500 text-[11px]">
                        {new Date(token.created_at).toLocaleDateString()}
                      </td>
                      <td className="py-3 px-4 text-right whitespace-nowrap">
                        <button
                          onClick={() => triggerMutation.mutate(token.id)}
                          disabled={triggerMutation.isPending}
                          className="inline-flex items-center gap-1.5 px-3 py-1 text-xs font-mono bg-rose-950 hover:bg-rose-900 text-rose-300 rounded border border-rose-800 transition disabled:opacity-50"
                          title="Simulate attacker accessing honeytoken -> Kafka event -> CRITICAL Alert"
                        >
                          <Zap className="w-3.5 h-3.5" />
                          <span>Simulate Trigger</span>
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Create Honeytoken Modal */}
        {isModalOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 backdrop-blur-sm p-4">
            <div className="w-full max-w-md bg-slate-900 border border-slate-800 rounded-xl p-6 shadow-2xl relative">
              <div className="flex items-center justify-between border-b border-slate-800 pb-3 mb-4">
                <div className="flex items-center gap-2 text-cyan-400 font-mono text-xs font-bold uppercase">
                  <ShieldAlert className="w-4 h-4" />
                  <span>Deploy Deception Honeytoken</span>
                </div>
                <button
                  onClick={() => setIsModalOpen(false)}
                  className="text-slate-400 hover:text-slate-200"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>

              <form onSubmit={handleCreateSubmit} className="space-y-4">
                <div>
                  <label className="block text-xs font-mono uppercase text-slate-400 mb-1">
                    Honeytoken Identifier Name
                  </label>
                  <input
                    type="text"
                    required
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="AWS Admin Secret Key"
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-xs font-mono text-slate-200 focus:outline-none focus:border-cyan-500"
                  />
                </div>

                <div>
                  <label className="block text-xs font-mono uppercase text-slate-400 mb-1">
                    Honeytoken Type
                  </label>
                  <select
                    value={type}
                    onChange={(e) => setType(e.target.value)}
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-xs font-mono text-slate-200 focus:outline-none focus:border-cyan-500"
                  >
                    <option value="API_KEY">API_KEY</option>
                    <option value="AWS_CRED">AWS_CRED</option>
                    <option value="DB_PASS">DB_PASS</option>
                    <option value="HONEY_URL">HONEY_URL</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-mono uppercase text-slate-400 mb-1">
                    Token Value (Optional - Auto Generated if empty)
                  </label>
                  <input
                    type="text"
                    value={tokenValue}
                    onChange={(e) => setTokenValue(e.target.value)}
                    placeholder="ak_test_fake_secret_key_123"
                    className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded text-xs font-mono text-slate-200 focus:outline-none focus:border-cyan-500"
                  />
                </div>

                <div className="flex justify-end gap-3 pt-3 border-t border-slate-800">
                  <button
                    type="button"
                    onClick={() => setIsModalOpen(false)}
                    className="px-4 py-2 text-xs font-mono text-slate-400 hover:bg-slate-800 rounded border border-slate-800"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={createMutation.isPending}
                    className="px-4 py-2 text-xs font-mono font-bold bg-cyan-600 hover:bg-cyan-500 text-slate-950 rounded shadow"
                  >
                    {createMutation.isPending ? 'Deploying...' : 'Deploy Honeytoken'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </DashboardLayout>
  );
}
