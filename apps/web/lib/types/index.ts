export interface User {
  id: string;
  tenant_id: string;
  email: string;
  role: string;
  created_at: string;
}

export interface Tenant {
  id: string;
  name: string;
  created_at: string;
}

export interface AuthResponse {
  user: User;
  tenant: Tenant;
  access_token: string;
  refresh_token: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  tenant_name: string;
  email: string;
  password: string;
  role?: string;
}

export interface SecurityEvent {
  id: string;
  tenant_id: string;
  event_type: string;
  source: string;
  source_ip: string;
  user_id: string;
  resource: string;
  action: string;
  severity: string;
  status: string;
  metadata: any;
  timestamp: string;
  created_at: string;
}

export interface EventFilterParams {
  limit?: number;
  offset?: number;
  page?: number;
}

export type AlertSeverity = 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW' | 'INFO';
export type AlertStatus = 'OPEN' | 'ACKNOWLEDGED' | 'RESOLVED' | 'CLOSED' | 'INVESTIGATING';

export interface Alert {
  id: string;
  tenant_id: string;
  severity: AlertSeverity;
  detection_type: string;
  source_ip: string;
  affected_asset: string;
  description: string;
  status: AlertStatus;
  timestamp: string;
  created_at: string;
  metadata: any;
}

export interface AlertFilterParams {
  severity?: string;
  status?: string;
  detection_type?: string;
  limit?: number;
  offset?: number;
  page?: number;
}

export interface UpdateAlertStatusRequest {
  status: AlertStatus;
}

export interface Honeytoken {
  id: string;
  tenant_id: string;
  name: string;
  type: string;
  token_value: string;
  status: 'ACTIVE' | 'TRIGGERED';
  last_triggered_at?: string | null;
  created_at: string;
}

export interface CreateHoneytokenRequest {
  name: string;
  type: string;
  token_value?: string;
}

export interface TriggerHoneytokenResponse {
  message: string;
  honeytoken_id: string;
  event_id: string;
  status: string;
}
