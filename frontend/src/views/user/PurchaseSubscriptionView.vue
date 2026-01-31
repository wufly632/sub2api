<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('purchase.title') }}
          </h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            {{ t('purchase.description') }}
          </p>
        </div>

        <div class="flex items-center gap-2">
          <a
            v-if="isValidUrl"
            :href="purchaseUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-secondary btn-sm"
          >
            <Icon name="externalLink" size="sm" class="mr-1.5" :stroke-width="2" />
            {{ t('purchase.openInNewTab') }}
          </a>
        </div>
      </div>

      <div v-if="loadingSettings" class="flex justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <div v-else-if="!purchaseEnabled" class="card p-10 text-center">
        <div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700">
          <Icon name="creditCard" size="lg" class="text-gray-400" />
        </div>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('purchase.notEnabledTitle') }}
        </h3>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ t('purchase.notEnabledDesc') }}
        </p>
      </div>

      <template v-else>
        <div v-if="purchaseInstructions" class="card p-6">
          <div class="prose prose-sm max-w-none dark:prose-invert" v-html="purchaseInstructions"></div>
        </div>

        <div class="card p-6">
          <div class="mb-4 flex items-center justify-between">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('purchase.plans') }}</h3>
            <button
              @click="loadPlans"
              :disabled="loadingPlans"
              class="btn btn-secondary btn-sm"
            >
              <Icon name="refresh" size="sm" :class="loadingPlans ? 'animate-spin' : ''" />
            </button>
          </div>

          <div v-if="loadingPlans" class="flex justify-center py-10">
            <div class="h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
          </div>

          <div v-else-if="plans.length === 0" class="text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('purchase.noPlans') }}
          </div>

          <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            <div v-for="plan in plans" :key="plan.id" class="rounded-xl border border-gray-200 p-5 dark:border-dark-700">
              <div class="flex items-start justify-between">
                <div>
                  <h4 class="text-base font-semibold text-gray-900 dark:text-white">{{ plan.name }}</h4>
                  <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                    {{ plan.description || '-' }}
                  </p>
                </div>
                <span class="badge badge-primary">{{ t('purchase.subscription') }}</span>
              </div>

              <div class="mt-4 flex items-baseline gap-2">
                <span class="text-2xl font-bold text-gray-900 dark:text-white">
                  ￥{{ (plan.purchase_price ?? 0).toFixed(2) }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  / {{ plan.default_validity_days || 30 }} {{ t('purchase.days') }}
                </span>
              </div>

              <ul class="mt-4 space-y-1 text-sm text-gray-500 dark:text-dark-400">
                <li>
                  {{ t('purchase.dailyLimit') }}:
                  <span class="text-gray-700 dark:text-gray-300">
                    {{ plan.daily_limit_usd ? `$${plan.daily_limit_usd}` : t('purchase.unlimited') }}
                  </span>
                </li>
                <li>
                  {{ t('purchase.weeklyLimit') }}:
                  <span class="text-gray-700 dark:text-gray-300">
                    {{ plan.weekly_limit_usd ? `$${plan.weekly_limit_usd}` : t('purchase.unlimited') }}
                  </span>
                </li>
                <li>
                  {{ t('purchase.monthlyLimit') }}:
                  <span class="text-gray-700 dark:text-gray-300">
                    {{ plan.monthly_limit_usd ? `$${plan.monthly_limit_usd}` : t('purchase.unlimited') }}
                  </span>
                </li>
              </ul>

              <button
                class="btn btn-primary mt-5 w-full"
                :disabled="creatingOrderId === plan.id"
                @click="handleCreateOrder(plan.id)"
              >
                <span v-if="creatingOrderId === plan.id" class="flex items-center justify-center">
                  <span class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-white/70 border-t-transparent"></span>
                  {{ t('purchase.creating') }}
                </span>
                <span v-else>{{ t('purchase.createOrder') }}</span>
              </button>
            </div>
          </div>
        </div>

        <div class="card p-6">
          <div class="mb-4 flex items-center justify-between">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('purchase.orders') }}</h3>
            <button
              @click="loadOrders"
              :disabled="loadingOrders"
              class="btn btn-secondary btn-sm"
            >
              <Icon name="refresh" size="sm" :class="loadingOrders ? 'animate-spin' : ''" />
            </button>
          </div>

          <div v-if="loadingOrders" class="flex justify-center py-8">
            <div class="h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
          </div>

          <div v-else-if="orders.length === 0" class="text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('purchase.noOrders') }}
          </div>

          <DataTable v-else :columns="orderColumns" :data="orders" :loading="loadingOrders">
            <template #cell-order_no="{ value }">
              <span class="font-mono text-sm text-gray-900 dark:text-gray-100">{{ value }}</span>
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
                {{ t('purchase.orderStatus.' + value) }}
              </span>
            </template>
            <template #cell-created_at="{ value }">
              <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
            </template>
            <template #cell-actions="{ row }">
              <div class="flex items-center gap-2">
                <button
                  v-if="row.status === 'pending' && (row.payment_qrcode || row.payment_url)"
                  class="btn btn-primary btn-sm"
                  @click="openPaymentDialog(row)"
                >
                  {{ t('purchase.viewQr') }}
                </button>
                <span v-else class="text-gray-400 dark:text-dark-500">-</span>
              </div>
            </template>
          </DataTable>

          <Pagination
            v-if="orderPagination.total > 0"
            class="mt-4"
            :page="orderPagination.page"
            :page-size="orderPagination.page_size"
            :total="orderPagination.total"
            @page-change="handleOrderPageChange"
            @page-size-change="handleOrderPageSizeChange"
          />
        </div>
      </template>
    </div>

    <BaseDialog
      :show="paymentDialogVisible"
      :title="t('purchase.qrTitle')"
      width="narrow"
      @close="closePaymentDialog"
    >
      <div v-if="paymentDialogOrder" class="space-y-4">
        <div class="text-center text-sm text-gray-600 dark:text-dark-300">
          <div class="text-xs text-gray-500 dark:text-dark-400">
            {{ t('purchase.columns.orderNo') }}: {{ paymentDialogOrder.order_no }}
          </div>
          <div class="mt-2 flex items-center justify-center gap-2">
            <span class="text-sm font-medium text-gray-600 dark:text-dark-300">
              {{ t('purchase.columns.amount') }}:
            </span>
            <span class="rounded-md bg-emerald-50 px-2 py-0.5 text-base font-semibold text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
              {{ paymentDialogOrder.currency }} {{ paymentDialogOrder.amount.toFixed(2) }}
            </span>
          </div>
        </div>
        <div class="rounded-lg border border-gray-200 bg-white p-4 text-center dark:border-dark-700 dark:bg-dark-800">
          <img
            v-if="paymentDialogQRCode"
            :src="paymentDialogQRCode"
            :alt="t('purchase.qrTitle')"
            class="mx-auto h-48 w-48 rounded-md object-contain"
          />
          <div v-else class="text-sm text-gray-500 dark:text-dark-400">
            {{ t('purchase.qrUnavailable') }}
          </div>
        </div>
        <div class="flex items-center justify-between rounded-md bg-gray-50 px-3 py-2 text-sm text-gray-600 dark:bg-dark-900/40 dark:text-dark-200">
          <span class="font-medium">{{ t('purchase.qrCountdown') }}</span>
          <span class="font-mono text-base font-semibold text-gray-900 dark:text-white">
            {{ paymentCountdownText }}
          </span>
        </div>
        <div class="text-sm text-gray-600 dark:text-dark-300">
          {{ t('purchase.qrTip') }}
        </div>
      </div>

      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <button class="btn btn-secondary" @click="closePaymentDialog">
            {{ t('common.close') }}
          </button>
          <button
            v-if="paymentDialogPaymentUrl"
            class="btn btn-primary"
            @click="openPayment(paymentDialogPaymentUrl)"
          >
            {{ t('purchase.openPaymentPage') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { purchaseAPI } from '@/api'
import type { Group, SubscriptionOrder } from '@/types'
import type { Column } from '@/components/common/types'
import { useAppStore } from '@/stores'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()
const subscriptionStore = useSubscriptionStore()

const loadingSettings = ref(false)
const loadingPlans = ref(false)
const loadingOrders = ref(false)
const plans = ref<Group[]>([])
const orders = ref<SubscriptionOrder[]>([])
const creatingOrderId = ref<number | null>(null)
const paymentDialogVisible = ref(false)
const paymentDialogOrder = ref<SubscriptionOrder | null>(null)
const paymentCountdownSeconds = ref(0)
let paymentCountdownTimer: ReturnType<typeof setInterval> | null = null
let paymentPollingTimer: ReturnType<typeof setInterval> | null = null
let paymentPollingInFlight = false
let paymentPollingToken = 0
let paymentAutoRefreshInFlight = false
let paymentAutoRefreshDisabled = false

const purchaseEnabled = computed(() => {
  return appStore.cachedPublicSettings?.purchase_subscription_enabled ?? false
})

const purchaseUrl = computed(() => {
  return (appStore.cachedPublicSettings?.purchase_subscription_url || '').trim()
})

const purchaseInstructions = computed(() => {
  return (appStore.cachedPublicSettings?.purchase_instructions || '').trim()
})

const isValidUrl = computed(() => {
  const url = purchaseUrl.value
  return url.startsWith('http://') || url.startsWith('https://')
})

const paymentDialogQRCode = computed(() => paymentDialogOrder.value?.payment_qrcode || '')
const paymentDialogPaymentUrl = computed(() => paymentDialogOrder.value?.payment_url || '')
const paymentCountdownText = computed(() => {
  const total = Math.max(0, paymentCountdownSeconds.value)
  const minutes = Math.floor(total / 60)
  const seconds = total % 60
  return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
})

const orderPagination = reactive({
  page: 1,
  page_size: 10,
  total: 0,
  pages: 0
})

const orderColumns = computed<Column[]>(() => [
  { key: 'order_no', label: t('purchase.columns.orderNo'), sortable: false },
  { key: 'group', label: t('purchase.columns.plan'), sortable: false },
  { key: 'amount', label: t('purchase.columns.amount'), sortable: false },
  { key: 'status', label: t('purchase.columns.status'), sortable: false },
  { key: 'created_at', label: t('purchase.columns.createdAt'), sortable: false },
  { key: 'actions', label: t('purchase.columns.actions'), sortable: false }
])

const loadPlans = async () => {
  loadingPlans.value = true
  try {
    plans.value = await purchaseAPI.listPlans()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('purchase.loadPlansFailed'))
  } finally {
    loadingPlans.value = false
  }
}

const loadOrders = async () => {
  loadingOrders.value = true
  try {
    const res = await purchaseAPI.listOrders(orderPagination.page, orderPagination.page_size)
    orders.value = res.items
    orderPagination.total = res.total
    orderPagination.page = res.page
    orderPagination.page_size = res.page_size
    orderPagination.pages = res.pages
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('purchase.loadOrdersFailed'))
  } finally {
    loadingOrders.value = false
  }
}

const handleCreateOrder = async (groupId: number) => {
  creatingOrderId.value = groupId
  try {
    const order = await purchaseAPI.createOrder({ group_id: groupId })
    if (order.status === 'paid') {
      appStore.showSuccess(t('purchase.orderCreated'))
    } else {
      appStore.showSuccess(t('purchase.orderCreatedPending'))
      if (order.payment_qrcode || order.payment_url) {
        openPaymentDialog(order)
      } else {
        appStore.showError(t('purchase.paymentUnavailable'))
      }
    }
    await Promise.all([loadOrders(), loadPlans(), subscriptionStore.fetchActiveSubscriptions(true)])
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('purchase.createOrderFailed'))
  } finally {
    creatingOrderId.value = null
  }
}

const openPayment = (url: string) => {
  if (!url) return
  window.open(url, '_blank', 'noopener,noreferrer')
}

const openPaymentDialog = (order: SubscriptionOrder) => {
  paymentDialogOrder.value = order
  paymentDialogVisible.value = true
  startPaymentCountdown(order)
  startPaymentPolling(order)
}

const closePaymentDialog = () => {
  paymentDialogVisible.value = false
  paymentDialogOrder.value = null
  clearPaymentCountdown()
  clearPaymentPolling()
  paymentAutoRefreshDisabled = false
}

const startPaymentCountdown = (order: SubscriptionOrder) => {
  clearPaymentCountdown()
  paymentAutoRefreshDisabled = false
  const expiresInSeconds = 5 * 60
  const createdAt = order.created_at ? new Date(order.created_at).getTime() : Date.now()
  const expiresAt = createdAt + expiresInSeconds * 1000

  const updateCountdown = () => {
    const remaining = Math.max(0, Math.ceil((expiresAt - Date.now()) / 1000))
    paymentCountdownSeconds.value = remaining
    if (remaining <= 0) {
      if (
        !paymentAutoRefreshDisabled &&
        !paymentAutoRefreshInFlight &&
        paymentDialogVisible.value &&
        paymentDialogOrder.value?.status === 'pending'
      ) {
        void refreshPaymentOrder()
      }
    }
  }

  updateCountdown()
  paymentCountdownTimer = setInterval(updateCountdown, 1000)
}

const clearPaymentCountdown = () => {
  if (paymentCountdownTimer) {
    clearInterval(paymentCountdownTimer)
    paymentCountdownTimer = null
  }
  paymentCountdownSeconds.value = 0
}

const startPaymentPolling = (order: SubscriptionOrder) => {
  clearPaymentPolling()
  const poll = async () => {
    if (!paymentDialogOrder.value || paymentPollingInFlight) return
    const pollToken = paymentPollingToken
    const currentOrderID = paymentDialogOrder.value.id
    paymentPollingInFlight = true
    try {
      const latest = await purchaseAPI.getOrder(currentOrderID)
      if (pollToken !== paymentPollingToken) {
        return
      }
      if (!paymentDialogOrder.value || paymentDialogOrder.value.id !== latest.id) {
        return
      }
      paymentDialogOrder.value = latest
      updateOrderInList(latest)
      if (latest.status !== 'pending') {
        clearPaymentPolling()
        if (latest.status === 'paid') {
          appStore.showSuccess(t('purchase.paymentSuccess'))
          await Promise.all([loadOrders(), subscriptionStore.fetchActiveSubscriptions(true)])
          closePaymentDialog()
        } else if (latest.status === 'canceled') {
          appStore.showError(t('purchase.paymentCanceled'))
          await loadOrders()
        }
      }
    } catch (error: any) {
      appStore.showError(error.response?.data?.detail || t('purchase.loadOrdersFailed'))
    } finally {
      paymentPollingInFlight = false
    }
  }
  paymentDialogOrder.value = order
  paymentPollingToken += 1
  void poll()
  paymentPollingTimer = setInterval(poll, 5000)
}

const clearPaymentPolling = () => {
  if (paymentPollingTimer) {
    clearInterval(paymentPollingTimer)
    paymentPollingTimer = null
  }
  paymentPollingInFlight = false
}

const updateOrderInList = (order: SubscriptionOrder) => {
  const index = orders.value.findIndex((item) => item.id === order.id)
  if (index >= 0) {
    orders.value.splice(index, 1, order)
  }
}

const refreshPaymentOrder = async () => {
  if (!paymentDialogOrder.value || paymentAutoRefreshInFlight) return
  paymentAutoRefreshInFlight = true
  appStore.showInfo(t('purchase.qrExpiredRefreshing'))
  try {
    const newOrder = await purchaseAPI.createOrder({ group_id: paymentDialogOrder.value.group_id })
    updateOrderInList(newOrder)
    await loadOrders()

    if (newOrder.status === 'paid') {
      appStore.showSuccess(t('purchase.paymentSuccess'))
      await subscriptionStore.fetchActiveSubscriptions(true)
      closePaymentDialog()
      return
    }

    if (newOrder.payment_qrcode || newOrder.payment_url) {
      paymentDialogOrder.value = newOrder
      startPaymentCountdown(newOrder)
      startPaymentPolling(newOrder)
      appStore.showSuccess(t('purchase.qrRefreshed'))
      return
    }

    paymentAutoRefreshDisabled = true
    clearPaymentCountdown()
    appStore.showError(t('purchase.paymentUnavailable'))
  } catch (error: any) {
    paymentAutoRefreshDisabled = true
    clearPaymentCountdown()
    appStore.showError(error.response?.data?.detail || t('purchase.qrRefreshFailed'))
  } finally {
    paymentAutoRefreshInFlight = false
  }
}

const handleOrderPageChange = (page: number) => {
  orderPagination.page = page
  loadOrders()
}

const handleOrderPageSizeChange = (pageSize: number) => {
  orderPagination.page_size = pageSize
  orderPagination.page = 1
  loadOrders()
}

onMounted(async () => {
  if (!appStore.publicSettingsLoaded) {
    loadingSettings.value = true
    try {
      await appStore.fetchPublicSettings()
    } finally {
      loadingSettings.value = false
    }
  }
  if (purchaseEnabled.value) {
    await Promise.all([loadPlans(), loadOrders()])
  }
})

onUnmounted(() => {
  clearPaymentCountdown()
  clearPaymentPolling()
})
</script>
