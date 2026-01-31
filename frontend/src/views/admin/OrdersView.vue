<template>
  <AppLayout>
    <TablePageLayout>
      <template #actions>
        <div class="flex justify-end gap-3">
          <button
            @click="loadOrders"
            :disabled="loading"
            class="btn btn-secondary"
            :title="t('common.refresh')"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </template>

      <template #filters>
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <input
              v-model="filters.order_no"
              type="text"
              class="input w-full sm:w-56"
              :placeholder="t('admin.orders.searchOrderNo')"
              @input="handleSearch"
            />
            <input
              v-model.number="filters.user_id"
              type="number"
              class="input w-full sm:w-40"
              :placeholder="t('admin.orders.filterUserId')"
              @input="handleSearch"
            />
            <input
              v-model.number="filters.group_id"
              type="number"
              class="input w-full sm:w-40"
              :placeholder="t('admin.orders.filterGroupId')"
              @input="handleSearch"
            />
            <Select
              v-model="filters.status"
              :options="statusOptions"
              class="w-full sm:w-40"
              @change="loadOrders"
            />
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="orders" :loading="loading">
          <template #cell-order_no="{ value }">
            <span class="font-mono text-sm text-gray-900 dark:text-gray-100">{{ value }}</span>
          </template>

          <template #cell-user="{ row }">
            <div class="text-sm text-gray-700 dark:text-gray-300">
              <div class="font-medium">{{ row.user?.email || '-' }}</div>
              <div class="text-xs text-gray-400">#{{ row.user_id }}</div>
            </div>
          </template>

          <template #cell-group="{ row }">
            <div class="text-sm text-gray-700 dark:text-gray-300">
              <div class="font-medium">{{ row.group?.name || `#${row.group_id}` }}</div>
              <div class="text-xs text-gray-400">{{ row.group?.platform || '-' }}</div>
            </div>
          </template>

          <template #cell-amount="{ row }">
            <span class="text-sm font-medium text-gray-900 dark:text-white">
              {{ row.currency }} {{ row.amount.toFixed(2) }}
            </span>
          </template>

          <template #cell-validity_days="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ value }} {{ t('admin.orders.days') }}</span>
          </template>

          <template #cell-status="{ value }">
            <span
              :class="[
                'badge',
                value === 'paid'
                  ? 'badge-success'
                  : value === 'pending'
                    ? 'badge-warning'
                    : 'badge-danger'
              ]"
            >
              {{ t('admin.orders.status.' + value) }}
            </span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-paid_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ value ? formatDateTime(value) : '-' }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-2">
              <button
                v-if="row.status === 'pending'"
                @click="openConfirm(row, 'paid')"
                class="btn btn-primary btn-sm"
              >
                {{ t('admin.orders.markPaid') }}
              </button>
              <button
                v-if="row.status === 'pending'"
                @click="openConfirm(row, 'cancel')"
                class="btn btn-danger btn-sm"
              >
                {{ t('admin.orders.cancel') }}
              </button>
              <span v-if="row.status !== 'pending'" class="text-gray-400 dark:text-dark-500">-</span>
            </div>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :page-size="pagination.page_size"
          :total="pagination.total"
          @page-change="handlePageChange"
          @page-size-change="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <ConfirmDialog
      :show="confirmVisible"
      :title="confirmTitle"
      :message="confirmMessage"
      :confirm-text="confirmButtonText"
      :confirm-variant="confirmVariant"
      @confirm="handleConfirm"
      @cancel="confirmVisible = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api'
import type { AdminSubscriptionOrder } from '@/types'
import type { Column } from '@/components/common/types'
import { formatDateTime } from '@/utils/format'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()

const orders = ref<AdminSubscriptionOrder[]>([])
const loading = ref(false)

const filters = reactive({
  status: '',
  order_no: '',
  user_id: undefined as number | undefined,
  group_id: undefined as number | undefined
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

const columns = computed<Column[]>(() => [
  { key: 'order_no', label: t('admin.orders.columns.orderNo'), sortable: false },
  { key: 'user', label: t('admin.orders.columns.user'), sortable: false },
  { key: 'group', label: t('admin.orders.columns.group'), sortable: false },
  { key: 'amount', label: t('admin.orders.columns.amount'), sortable: false },
  { key: 'validity_days', label: t('admin.orders.columns.validity'), sortable: false },
  { key: 'status', label: t('admin.orders.columns.status'), sortable: false },
  { key: 'created_at', label: t('admin.orders.columns.createdAt'), sortable: false },
  { key: 'paid_at', label: t('admin.orders.columns.paidAt'), sortable: false },
  { key: 'actions', label: t('admin.orders.columns.actions'), sortable: false }
])

const statusOptions = computed(() => [
  { value: '', label: t('admin.orders.allStatus') },
  { value: 'pending', label: t('admin.orders.status.pending') },
  { value: 'paid', label: t('admin.orders.status.paid') },
  { value: 'canceled', label: t('admin.orders.status.canceled') }
])

let searchTimeout: ReturnType<typeof setTimeout> | null = null
const handleSearch = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.page = 1
    loadOrders()
  }, 300)
}

const loadOrders = async () => {
  loading.value = true
  try {
    const res = await adminAPI.orders.list(pagination.page, pagination.page_size, {
      status: filters.status || undefined,
      order_no: filters.order_no || undefined,
      user_id: filters.user_id || undefined,
      group_id: filters.group_id || undefined
    })
    orders.value = res.items
    pagination.total = res.total
    pagination.page = res.page
    pagination.page_size = res.page_size
    pagination.pages = res.pages
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.orders.loadFailed'))
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadOrders()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadOrders()
}

const confirmVisible = ref(false)
const confirmAction = ref<'paid' | 'cancel' | null>(null)
const selectedOrder = ref<AdminSubscriptionOrder | null>(null)

const confirmTitle = computed(() => {
  if (!confirmAction.value) return ''
  return confirmAction.value === 'paid'
    ? t('admin.orders.confirmPaidTitle')
    : t('admin.orders.confirmCancelTitle')
})

const confirmMessage = computed(() => {
  if (!confirmAction.value || !selectedOrder.value) return ''
  return confirmAction.value === 'paid'
    ? t('admin.orders.confirmPaidMessage', { orderNo: selectedOrder.value.order_no })
    : t('admin.orders.confirmCancelMessage', { orderNo: selectedOrder.value.order_no })
})

const confirmButtonText = computed(() => {
  if (!confirmAction.value) return ''
  return confirmAction.value === 'paid' ? t('admin.orders.markPaid') : t('admin.orders.cancel')
})

const confirmVariant = computed(() => (confirmAction.value === 'cancel' ? 'danger' : 'primary'))

const openConfirm = (order: AdminSubscriptionOrder, action: 'paid' | 'cancel') => {
  selectedOrder.value = order
  confirmAction.value = action
  confirmVisible.value = true
}

const handleConfirm = async () => {
  if (!selectedOrder.value || !confirmAction.value) return
  try {
    if (confirmAction.value === 'paid') {
      await adminAPI.orders.markPaid(selectedOrder.value.id)
      appStore.showSuccess(t('admin.orders.markPaidSuccess'))
    } else {
      await adminAPI.orders.cancel(selectedOrder.value.id)
      appStore.showSuccess(t('admin.orders.cancelSuccess'))
    }
    confirmVisible.value = false
    selectedOrder.value = null
    confirmAction.value = null
    loadOrders()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.orders.actionFailed'))
  }
}

onMounted(() => {
  loadOrders()
})
</script>
