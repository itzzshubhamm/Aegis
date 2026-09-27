'use client';

import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../../lib/context/AuthContext';
import { Sidebar } from './Sidebar';
import { TopBar } from './TopBar';
import { LoadingState } from '../ui/States';

export const DashboardLayout: React.FC<{ children: React.ReactNode; onRefresh?: () => void }> = ({
  children,
  onRefresh,
}) => {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.replace('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <div className="h-screen w-screen flex items-center justify-center bg-slate-950">
        <LoadingState message="Authenticating Aegis SOC session..." />
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="flex min-h-screen bg-slate-950 text-slate-100">
      <Sidebar />
      <div className="flex-1 flex flex-col min-w-0">
        <TopBar onRefresh={onRefresh} />
        <main className="p-6 flex-1 overflow-auto">{children}</main>
      </div>
    </div>
  );
};
