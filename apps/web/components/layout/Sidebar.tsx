'use client';

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Shield, LayoutDashboard, AlertTriangle, Activity, Key, LogOut, Building } from 'lucide-react';
import { useAuth } from '../../lib/context/AuthContext';

export const Sidebar: React.FC = () => {
  const pathname = usePathname();
  const { tenant, user, logout } = useAuth();

  const navItems = [
    { name: 'Overview', href: '/dashboard', icon: LayoutDashboard },
    { name: 'Alerts', href: '/dashboard/alerts', icon: AlertTriangle },
    { name: 'Events', href: '/dashboard/events', icon: Activity },
    { name: 'Honeytokens', href: '/dashboard/honeytokens', icon: Key },
  ];

  return (
    <aside className="w-64 bg-slate-900 border-r border-slate-800 flex flex-col justify-between h-screen sticky top-0">
      <div>
        {/* Brand */}
        <div className="h-16 flex items-center px-5 border-b border-slate-800 gap-3">
          <div className="p-2 rounded bg-cyan-950 text-cyan-400 border border-cyan-800">
            <Shield className="w-5 h-5" />
          </div>
          <div>
            <h1 className="font-bold text-sm tracking-wider text-slate-100 uppercase">AEGIS SOC</h1>
            <p className="text-[10px] font-mono text-cyan-500 tracking-tight">Adaptive Cyber Defense</p>
          </div>
        </div>

        {/* Tenant Identity Context Badge */}
        {tenant && (
          <div className="mx-3 mt-4 p-2.5 rounded bg-slate-950/80 border border-slate-800/80">
            <div className="flex items-center gap-2 text-slate-400 text-xs mb-1">
              <Building className="w-3.5 h-3.5 text-cyan-500" />
              <span className="font-mono text-[11px] uppercase tracking-wider text-slate-400">Active Tenant</span>
            </div>
            <p className="text-xs font-semibold text-slate-200 truncate">{tenant.name}</p>
          </div>
        )}

        {/* Navigation */}
        <nav className="mt-4 px-3 space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = pathname === item.href || (item.href !== '/dashboard' && pathname.startsWith(item.href));

            return (
              <Link
                key={item.name}
                href={item.href}
                className={`flex items-center gap-3 px-3 py-2.5 rounded-md text-xs font-medium transition-colors ${
                  isActive
                    ? 'bg-cyan-950/60 text-cyan-300 border border-cyan-800/60 font-semibold'
                    : 'text-slate-400 hover:bg-slate-800/60 hover:text-slate-200 border border-transparent'
                }`}
              >
                <Icon className={`w-4 h-4 ${isActive ? 'text-cyan-400' : 'text-slate-500'}`} />
                <span>{item.name}</span>
              </Link>
            );
          })}
        </nav>
      </div>

      {/* User Session Footer */}
      <div className="p-3 border-t border-slate-800 bg-slate-950/40">
        {user && (
          <div className="mb-2 px-2">
            <p className="text-xs font-medium text-slate-300 truncate">{user.email}</p>
            <span className="inline-block mt-0.5 text-[10px] font-mono uppercase text-slate-500 bg-slate-800 px-1.5 py-0.5 rounded">
              {user.role}
            </span>
          </div>
        )}
        <button
          onClick={() => logout()}
          className="w-full flex items-center justify-center gap-2 px-3 py-2 text-xs font-mono text-slate-400 hover:text-red-400 hover:bg-red-950/40 rounded border border-slate-800 hover:border-red-900 transition"
        >
          <LogOut className="w-3.5 h-3.5" />
          <span>Sign Out</span>
        </button>
      </div>
    </aside>
  );
};
