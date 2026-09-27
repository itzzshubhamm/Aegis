import { apiClient } from './client';
import { Honeytoken, CreateHoneytokenRequest, TriggerHoneytokenResponse } from '../types';

export const honeytokensApi = {
  getHoneytokens: async (): Promise<Honeytoken[]> => {
    const res = await apiClient.get<Honeytoken[]>('/deception/honeytokens');
    return res.data;
  },

  getHoneytokenById: async (id: string): Promise<Honeytoken> => {
    const res = await apiClient.get<Honeytoken>(`/deception/honeytokens/${id}`);
    return res.data;
  },

  createHoneytoken: async (data: CreateHoneytokenRequest): Promise<Honeytoken> => {
    const res = await apiClient.post<Honeytoken>('/deception/honeytokens', data);
    return res.data;
  },

  triggerHoneytoken: async (id: string): Promise<TriggerHoneytokenResponse> => {
    const res = await apiClient.post<TriggerHoneytokenResponse>(`/deception/trigger/${id}`);
    return res.data;
  },
};
