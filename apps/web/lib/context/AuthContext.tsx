'use client';

import React, { createContext, useContext, useState, useEffect } from 'react';
import { User, Tenant, LoginRequest, RegisterRequest } from '../types';
import { authApi } from '../api/auth';
import { getAccessToken, clearTokens } from '../api/client';

interface AuthContextType {
  user: User | null;
  tenant: Tenant | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (data: LoginRequest) => Promise<void>;
  register: (data: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  useEffect(() => {
    // Restore session from localStorage on initial render
    const token = getAccessToken();
    if (token) {
      const storedUser = localStorage.getItem('aegis_user');
      const storedTenant = localStorage.getItem('aegis_tenant');
      if (storedUser && storedTenant) {
        try {
          setUser(JSON.parse(storedUser));
          setTenant(JSON.parse(storedTenant));
        } catch {
          clearTokens();
        }
      }
    }
    setIsLoading(false);
  }, []);

  const login = async (data: LoginRequest) => {
    setIsLoading(true);
    try {
      const res = await authApi.login(data);
      setUser(res.user);
      setTenant(res.tenant);
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (data: RegisterRequest) => {
    setIsLoading(true);
    try {
      const res = await authApi.register(data);
      setUser(res.user);
      setTenant(res.tenant);
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    setIsLoading(true);
    try {
      await authApi.logout();
    } finally {
      setUser(null);
      setTenant(null);
      setIsLoading(false);
      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        tenant,
        isAuthenticated: !!user && !!getAccessToken(),
        isLoading,
        login,
        register,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
