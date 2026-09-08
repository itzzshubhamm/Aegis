import { apiClient, setTokens, clearTokens } from './client';
import { AuthResponse, LoginRequest, RegisterRequest } from '../types';

export const authApi = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const res = await apiClient.post<AuthResponse>('/auth/login', data);
    setTokens(res.data.access_token, res.data.refresh_token);
    if (typeof window !== 'undefined') {
      localStorage.setItem('aegis_user', JSON.stringify(res.data.user));
      localStorage.setItem('aegis_tenant', JSON.stringify(res.data.tenant));
    }
    return res.data;
  },

  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    const res = await apiClient.post<AuthResponse>('/auth/register', data);
    setTokens(res.data.access_token, res.data.refresh_token);
    if (typeof window !== 'undefined') {
      localStorage.setItem('aegis_user', JSON.stringify(res.data.user));
      localStorage.setItem('aegis_tenant', JSON.stringify(res.data.tenant));
    }
    return res.data;
  },

  logout: async (): Promise<void> => {
    try {
      await apiClient.post('/auth/logout');
    } catch {
      // Ignore network/server errors during logout
    } finally {
      clearTokens();
    }
  },
};
