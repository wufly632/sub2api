/**
 * Admin Orders API endpoints
 */

import { apiClient } from '../client'
import type { PaginatedResponse, AdminSubscriptionOrder } from '@/types'

export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    status?: string
    order_no?: string
    user_id?: number
    group_id?: number
  }
): Promise<PaginatedResponse<AdminSubscriptionOrder>> {
  const { data } = await apiClient.get<PaginatedResponse<AdminSubscriptionOrder>>('/admin/orders', {
    params: {
      page,
      page_size: pageSize,
      ...filters
    }
  })
  return data
}

export async function getById(id: number): Promise<AdminSubscriptionOrder> {
  const { data } = await apiClient.get<AdminSubscriptionOrder>(`/admin/orders/${id}`)
  return data
}

export async function markPaid(id: number): Promise<AdminSubscriptionOrder> {
  const { data } = await apiClient.post<AdminSubscriptionOrder>(`/admin/orders/${id}/mark-paid`)
  return data
}

export async function cancel(id: number): Promise<AdminSubscriptionOrder> {
  const { data } = await apiClient.post<AdminSubscriptionOrder>(`/admin/orders/${id}/cancel`)
  return data
}

export const ordersAPI = { list, getById, markPaid, cancel }

export default ordersAPI
