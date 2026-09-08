import { apiClient } from './client';
import { SecurityEvent, EventFilterParams } from '../types';

export const eventsApi = {
  getEvents: async (params?: EventFilterParams): Promise<SecurityEvent[]> => {
    const res = await apiClient.get<SecurityEvent[]>('/events', { params });
    return res.data;
  },

  getEventById: async (id: string): Promise<SecurityEvent> => {
    const res = await apiClient.get<SecurityEvent>(`/events/${id}`);
    return res.data;
  },

  ingestEvent: async (data: Partial<SecurityEvent>): Promise<SecurityEvent> => {
    const res = await apiClient.post<SecurityEvent>('/events/ingest', data);
    return res.data;
  },
};
