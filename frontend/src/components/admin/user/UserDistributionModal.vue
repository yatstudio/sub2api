<template>
  <BaseDialog :show="show" :title="t('admin.users.distribution.title')" width="wide" @close="handleClose">
    <div v-if="user" class="space-y-4 distribution-admin">
      <section class="rounded-xl distribution-admin-hero p-4">
        <div class="text-xs uppercase tracking-[0.16em] opacity-80">Sub2API Partner Ops</div>
        <div class="mt-1 text-lg font-semibold">{{ user.email }}</div>
        <div class="text-sm opacity-90">{{ user.username || '-' }}</div>
      </section>

      <div v-if="loading" class="flex justify-center py-8">
        <svg class="h-8 w-8 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
      </div>

      <template v-else>
        <div class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4">
          <div class="rounded-xl border border-indigo-200/70 bg-white p-3 dark:border-indigo-900/60 dark:bg-dark-800">
            <div class="text-xs text-gray-500">{{ t('distribution.inviteCode') }}</div>
            <div class="mt-1 text-lg font-semibold">{{ profile?.invite_code || '-' }}</div>
          </div>
          <div class="rounded-xl border border-indigo-200/70 bg-white p-3 dark:border-indigo-900/60 dark:bg-dark-800">
            <div class="text-xs text-gray-500">{{ t('distribution.availableCommission') }}</div>
            <div class="mt-1 text-lg font-semibold">${{ toMoney(summary?.available_commission) }}</div>
          </div>
          <div class="rounded-xl border border-indigo-200/70 bg-white p-3 dark:border-indigo-900/60 dark:bg-dark-800">
            <div class="text-xs text-gray-500">{{ t('distribution.level1Team') }}</div>
            <div class="mt-1 text-lg font-semibold">{{ summary?.level1_team_count || 0 }}</div>
          </div>
          <div class="rounded-xl border border-indigo-200/70 bg-white p-3 dark:border-indigo-900/60 dark:bg-dark-800">
            <div class="text-xs text-gray-500">{{ t('distribution.level2Team') }}</div>
            <div class="mt-1 text-lg font-semibold">{{ summary?.level2_team_count || 0 }}</div>
          </div>
        </div>

        <section class="rounded-xl border border-indigo-200/70 p-4 dark:border-indigo-900/60">
          <div class="mb-3 text-sm font-semibold">{{ t('admin.users.distribution.commissionRate') }}</div>
          <div class="flex flex-wrap items-center gap-2">
            <input v-model.number="commissionRateInput" type="number" min="0" max="1" step="0.01" class="input w-40" />
            <button class="btn btn-primary" :disabled="savingRate" @click="saveCommissionRate">
              {{ savingRate ? t('common.loading') : t('common.save') }}
            </button>
          </div>
        </section>

        <section class="rounded-xl border border-indigo-200/70 p-4 dark:border-indigo-900/60">
          <div class="mb-3 flex items-center justify-between gap-2">
            <div class="text-sm font-semibold">{{ t('distribution.teamTitle') }}</div>
            <div class="flex items-center gap-2 text-xs">
              <button class="btn distribution-pill" :class="teamLevel === 1 ? 'btn-primary' : ''" @click="changeTeamLevel(1)">{{ t('distribution.level1Team') }}</button>
              <button class="btn distribution-pill" :class="teamLevel === 2 ? 'btn-primary' : ''" @click="changeTeamLevel(2)">{{ t('distribution.level2Team') }}</button>
            </div>
          </div>
          <div class="max-h-44 overflow-y-auto text-sm">
            <div v-for="item in team.items" :key="`${item.team_level}-${item.user_id}`" class="flex items-center justify-between border-b border-gray-100 py-2 dark:border-dark-700">
              <div>
                <div class="font-medium">{{ item.email || item.username || item.user_id }}</div>
                <div class="text-xs text-gray-500">ID: {{ item.user_id }}</div>
              </div>
              <div class="text-right">
                <div>${{ toMoney(item.total_contribution) }}</div>
                <div class="text-xs text-gray-500">+${{ toMoney(item.commission_generated) }}</div>
              </div>
            </div>
            <div v-if="team.items.length === 0" class="py-4 text-center text-gray-500">{{ t('distribution.noData') }}</div>
          </div>
        </section>

        <section class="rounded-xl border border-indigo-200/70 p-4 dark:border-indigo-900/60">
          <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
            <div class="text-sm font-semibold">{{ t('distribution.withdrawalsTitle') }}</div>
            <div class="flex flex-wrap items-center gap-2">
              <select v-model="withdrawalStatus" class="input w-40" @change="loadWithdrawals">
                <option value="all">{{ t('distribution.all') }}</option>
                <option value="pending">{{ t('distribution.statusPending') }}</option>
                <option value="approved">{{ t('distribution.statusApproved') }}</option>
                <option value="rejected">{{ t('distribution.statusRejected') }}</option>
              </select>
              <div class="flex flex-wrap items-center gap-2 rounded-lg border border-indigo-200/70 px-2 py-1 dark:border-indigo-900/50">
                <span class="text-xs text-gray-500">{{ t('admin.users.distribution.riskThreshold') }}</span>
                <input v-model.number="riskThresholdInput" type="number" min="0" step="10" class="input w-24" />
                <span class="text-xs text-gray-500">{{ t('admin.users.distribution.cooldownDays') }}</span>
                <input v-model.number="riskCooldownDaysInput" type="number" min="0" step="1" class="input w-20" />
                <span class="text-xs text-gray-500">{{ t('admin.users.distribution.dailyLimitCount') }}</span>
                <input v-model.number="riskDailyLimitCountInput" type="number" min="0" step="1" class="input w-20" />
                <span class="text-xs text-gray-500">{{ t('admin.users.distribution.dailyLimitAmount') }}</span>
                <input v-model.number="riskDailyLimitAmountInput" type="number" min="0" step="10" class="input w-24" />
                <button class="btn" :disabled="savingRiskThreshold" @click="saveRiskThreshold">{{ savingRiskThreshold ? t('common.loading') : t('common.save') }}</button>
              </div>
              <button class="btn btn-primary" :disabled="selectedPendingIds.length === 0 || batchReviewing" @click="batchReview('approved')">
                {{ t('admin.users.distribution.batchApprove') }} ({{ selectedPendingIds.length }})
              </button>
              <button class="btn" :disabled="selectedPendingIds.length === 0 || batchReviewing" @click="batchReview('rejected')">
                {{ t('admin.users.distribution.batchReject') }}
              </button>
            </div>
          </div>
          <div class="max-h-52 overflow-y-auto text-sm">
            <div v-for="item in withdrawals.items" :key="item.id" class="border-b border-gray-100 py-2 dark:border-dark-700">
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-start gap-2">
                  <input
                    v-if="item.status === 'pending'"
                    type="checkbox"
                    :checked="selectedPendingIds.includes(item.id)"
                    class="mt-1"
                    @change="togglePendingSelection(item.id)"
                  />
                  <div>
                    <div class="font-medium">
                      #{{ item.id }} · ${{ toMoney(item.amount) }} · {{ item.account_ref }}
                      <span v-if="item.amount >= riskThreshold" class="ml-2 rounded bg-rose-100 px-1.5 py-0.5 text-[10px] font-semibold text-rose-700">{{ t('admin.users.distribution.riskHighAmount') }}</span>
                    </div>
                    <div class="text-xs text-gray-500">{{ statusLabel(item.status) }} · {{ formatDateTime(item.created_at) }}</div>
                  </div>
                </div>
                <div v-if="item.status === 'pending'" class="flex gap-2">
                  <button class="btn btn-primary" :disabled="reviewingId === item.id || batchReviewing" @click="review(item.id, 'approved')">{{ t('distribution.statusApproved') }}</button>
                  <button class="btn" :disabled="reviewingId === item.id || batchReviewing" @click="review(item.id, 'rejected')">{{ t('distribution.statusRejected') }}</button>
                </div>
              </div>
            </div>
            <div v-if="withdrawals.items.length === 0" class="py-4 text-center text-gray-500">{{ t('distribution.noData') }}</div>
          </div>
        </section>
      </template>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import type { AdminUser } from '@/types'
import type { DistributionProfile, DistributionSummary, DistributionTeamMember, DistributionWithdrawalRequest } from '@/api/user'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
const emit = defineEmits(['close'])

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const savingRate = ref(false)
const reviewingId = ref<number | null>(null)
const batchReviewing = ref(false)
const savingRiskThreshold = ref(false)

const riskThreshold = ref(1000)
const riskThresholdInput = ref(1000)
const riskCooldownDaysInput = ref(0)
const riskDailyLimitCountInput = ref(1)
const riskDailyLimitAmountInput = ref(10000)
const selectedPendingIds = ref<number[]>([])

const profile = ref<DistributionProfile | null>(null)
const summary = ref<DistributionSummary | null>(null)
const team = ref<{ items: DistributionTeamMember[] }>({ items: [] })
const withdrawals = ref<{ items: DistributionWithdrawalRequest[] }>({ items: [] })

const teamLevel = ref<1 | 2>(1)
const withdrawalStatus = ref<'all' | 'pending' | 'approved' | 'rejected'>('pending')
const commissionRateInput = ref(0.1)

const toMoney = (value?: number) => Number(value || 0).toFixed(2)

const statusLabel = (status: string) => {
  if (status === 'approved') return t('distribution.statusApproved')
  if (status === 'rejected') return t('distribution.statusRejected')
  return t('distribution.statusPending')
}

const loadTeam = async () => {
  if (!props.user) return
  const data = await adminAPI.users.listUserDistributionTeam(props.user.id, teamLevel.value, 1, 10)
  team.value = { items: data.items || [] }
}

const loadWithdrawals = async () => {
  if (!props.user) return
  const status = withdrawalStatus.value === 'all' ? undefined : withdrawalStatus.value
  const data = await adminAPI.users.listUserDistributionWithdrawals(props.user.id, 1, 20, status)
  withdrawals.value = { items: data.items || [] }
  selectedPendingIds.value = []
}

const togglePendingSelection = (withdrawalId: number) => {
  if (selectedPendingIds.value.includes(withdrawalId)) {
    selectedPendingIds.value = selectedPendingIds.value.filter((id) => id !== withdrawalId)
  } else {
    selectedPendingIds.value = [...selectedPendingIds.value, withdrawalId]
  }
}

const load = async () => {
  if (!props.user) return
  loading.value = true
  try {
    const [p, s, riskSettings] = await Promise.all([
      adminAPI.users.getUserDistributionProfile(props.user.id),
      adminAPI.users.getUserDistributionSummary(props.user.id),
      adminAPI.users.getDistributionRiskSettings()
    ])
    profile.value = p
    summary.value = s
    commissionRateInput.value = Number(p.commission_rate || 0)
    riskThreshold.value = Number(riskSettings.withdrawal_risk_threshold || 0)
    riskThresholdInput.value = riskThreshold.value
    riskCooldownDaysInput.value = Number(riskSettings.withdrawal_cooldown_days || 0)
    riskDailyLimitCountInput.value = Number(riskSettings.withdrawal_daily_limit_count || 0)
    riskDailyLimitAmountInput.value = Number(riskSettings.withdrawal_daily_limit_amount || 0)
    await Promise.all([loadTeam(), loadWithdrawals()])
  } catch (error: any) {
    appStore.showError(error?.message || t('distribution.loadFailed'))
  } finally {
    loading.value = false
  }
}

const saveCommissionRate = async () => {
  if (!props.user) return
  savingRate.value = true
  try {
    await adminAPI.users.updateUserDistributionCommissionRate(props.user.id, Number(commissionRateInput.value || 0))
    appStore.showSuccess(t('common.saved'))
    await load()
  } catch (error: any) {
    appStore.showError(error?.message || t('errors.somethingWentWrong'))
  } finally {
    savingRate.value = false
  }
}

const changeTeamLevel = async (level: 1 | 2) => {
  if (teamLevel.value === level) return
  teamLevel.value = level
  await loadTeam()
}

const saveRiskThreshold = async () => {
  savingRiskThreshold.value = true
  try {
    const saved = await adminAPI.users.updateDistributionRiskSettings({
      withdrawal_risk_threshold: Number(riskThresholdInput.value || 0),
      withdrawal_cooldown_days: Number(riskCooldownDaysInput.value || 0),
      withdrawal_daily_limit_count: Number(riskDailyLimitCountInput.value || 0),
      withdrawal_daily_limit_amount: Number(riskDailyLimitAmountInput.value || 0)
    })
    riskThreshold.value = Number(saved.withdrawal_risk_threshold || 0)
    riskThresholdInput.value = riskThreshold.value
    riskCooldownDaysInput.value = Number(saved.withdrawal_cooldown_days || 0)
    riskDailyLimitCountInput.value = Number(saved.withdrawal_daily_limit_count || 0)
    riskDailyLimitAmountInput.value = Number(saved.withdrawal_daily_limit_amount || 0)
    appStore.showSuccess(t('common.saved'))
  } catch (error: any) {
    appStore.showError(error?.message || t('errors.somethingWentWrong'))
  } finally {
    savingRiskThreshold.value = false
  }
}

const review = async (withdrawalId: number, status: 'approved' | 'rejected') => {
  if (!props.user) return
  reviewingId.value = withdrawalId
  try {
    const reviewNote = window.prompt(t('admin.users.distribution.reviewNotePrompt')) || ''
    await adminAPI.users.reviewUserDistributionWithdrawal(props.user.id, withdrawalId, {
      status,
      review_note: reviewNote
    })
    appStore.showSuccess(t('common.saved'))
    await loadWithdrawals()
  } catch (error: any) {
    appStore.showError(error?.message || t('errors.somethingWentWrong'))
  } finally {
    reviewingId.value = null
  }
}

const batchReview = async (status: 'approved' | 'rejected') => {
  if (!props.user || selectedPendingIds.value.length === 0) return
  batchReviewing.value = true
  try {
    const reviewNote = window.prompt(t('admin.users.distribution.reviewNotePrompt')) || ''
    for (const withdrawalId of selectedPendingIds.value) {
      await adminAPI.users.reviewUserDistributionWithdrawal(props.user.id, withdrawalId, {
        status,
        review_note: reviewNote
      })
    }
    appStore.showSuccess(t('common.saved'))
    await loadWithdrawals()
  } catch (error: any) {
    appStore.showError(error?.message || t('errors.somethingWentWrong'))
  } finally {
    batchReviewing.value = false
  }
}

watch(
  () => props.show,
  (v) => {
    if (v) {
      load()
    }
  }
)

const handleClose = () => {
  emit('close')
}
</script>

<style scoped>
.distribution-admin-hero {
  color: #eef2ff;
  background: linear-gradient(135deg, #1e3a8a 0%, #3730a3 55%, #4f46e5 100%);
}

.distribution-pill {
  border-color: rgba(99, 102, 241, 0.28);
}
</style>
