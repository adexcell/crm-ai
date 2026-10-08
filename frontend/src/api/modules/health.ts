import { api } from '@/api/client';
import type { HealthStatus } from '@/api/types';


export async function checkLiveness(): Promise<{ status: string }> {
  return api.get<{ status: string }>('/health/live')
}

export async function checkReadiness(): Promise<HealthStatus> {
  return api.get<HealthStatus>('/health/ready')
}
