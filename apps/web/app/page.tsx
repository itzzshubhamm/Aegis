'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../lib/context/AuthContext';
import { LoadingState } from '../components/ui/States';

export default function RootPage() {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading) {
      if (isAuthenticated) {
        router.replace('/dashboard');
      } else {
        router.replace('/login');
      }
    }
  }, [isAuthenticated, isLoading, router]);

  return (
    <div className="h-screen w-screen flex items-center justify-center bg-slate-950">
      <LoadingState message="Redirecting to Aegis Console..." />
    </div>
  );
}
