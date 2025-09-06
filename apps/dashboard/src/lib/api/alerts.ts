import { request } from './request';

export type AlertRule = {
  service: string;
  metric: string;
  threshold: number;
};

export async function fetchAlertRules(): Promise<{ items: AlertRule[] }> {
  return request<{ items: AlertRule[] }>('/v1/alert-rules');
}

export async function updateAlertRule(payload: AlertRule): Promise<void> {
  await request('/v1/alert-rules', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
}

export type AlertWindowItem = {
  service: string;
  metric: string;
  start: string;
  end: string | null;
  expected_cost: number;
  real_cost: number;
};

export type AlertWindowsResponse = {
  from_date: string;
  to_date: string;
  interval: number;
  items: AlertWindowItem[];
};

export async function fetchAlertWindows(): Promise<AlertWindowsResponse> {
  return request<AlertWindowsResponse>('/v1/alert-windows');
}
