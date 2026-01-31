/**
 * Purchase API endpoints (user)
 */

import { apiClient } from './client'
import type { Group, PaginatedResponse, SubscriptionOrder } from '@/types'

export async function listPlans(): Promise<Group[]> {
  const { data } = await apiClient.get<Group[]>('/purchase/plans')
  return data
}

export async function createOrder(payload: { group_id: number; notes?: string }): Promise<SubscriptionOrder> {
  const { data } = await apiClient.post<SubscriptionOrder>('/purchase/orders', payload)
  return data
}

export async function listOrders(
  page: number = 1,
  pageSize: number = 20,
  filters?: { status?: string }
): Promise<PaginatedResponse<SubscriptionOrder>> {
  const { data } = await apiClient.get<PaginatedResponse<SubscriptionOrder>>('/purchase/orders', {
    params: { page, page_size: pageSize, ...filters }
  })
  return data
}

export async function getOrder(id: number): Promise<SubscriptionOrder> {
  const { data } = await apiClient.get<SubscriptionOrder>(`/purchase/orders/${id}`)
  return data
}

export const purchaseAPI = { listPlans, createOrder, listOrders, getOrder }

export default purchaseAPI
