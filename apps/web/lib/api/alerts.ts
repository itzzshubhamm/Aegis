import { apiClient } from './client';
import { Alert, AlertFilterParams, AlertStatus } from '../types';

export const alertsApi = {
  getAlerts: async (params?: AlertFilterParams): Promise<Alert[]> => {
    const res = await apiClient.get<Alert[]>('/alerts', { params });
    return res.data;
  },

  getAlertById: async (id: string): Promise<Alert> => {
    const res = await apiClient.get<Alert>(`/alerts/${id}`);
    return res.data;
  },

  updateAlertStatus: async (id: string, status: AlertStatus): Promise<Alert> => {
    const res = await apiClient.patch<Alert>(`/alerts/${id}`, { status });
    return res.data;
  },
};
