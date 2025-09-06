import { request } from './request';

export type UsageItem = {
  service: string;
  metric: string;
  cost: number;
  timestamp: string; // ISO date
};

export type UsageResponse = {
  from_date: string;
  to_date: string;
  interval: number; // seconds
  items: UsageItem[];
};

export async function fetchUsage(): Promise<UsageResponse> {
  return request<UsageResponse>('/v1/usage');
}

export type PercentileItem = {
  service: string;
  metric: string;
  p50: number;
  p90: number;
  p95: number;
  pmax: number;
};

export type PercentilesResponse = {
  from_date: string;
  to_date: string;
  interval: number; // seconds
  items: PercentileItem[];
};

export async function fetchUsagePercentiles(): Promise<PercentilesResponse> {
  return request<PercentilesResponse>('/v1/usage-percentiles');
}
