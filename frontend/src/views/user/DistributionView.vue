<template>
  <AppLayout>
    <div class="space-y-6 distribution-brand">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else>
        <section class="distribution-hero p-5 md:p-6">
          <div class="text-xs font-semibold uppercase tracking-[0.18em] distribution-hero-kicker">Sub2API Partner Network</div>
          <h1 class="mt-2 text-2xl font-semibold md:text-3xl">{{ t('distribution.title') }}</h1>
          <p class="mt-2 text-sm md:text-base opacity-90">{{ t('distribution.description') }}</p>
        </section>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
          <div class="card p-4 distribution-card">
            <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('distribution.inviteCode') }}</div>
            <div class="mt-1 text-lg font-semibold">{{ profile?.invite_code || '-' }}</div>
          </div>
          <div class="card p-4 distribution-card">
            <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('distribution.availableCommission') }}</div>
            <div class="mt-1 text-lg font-semibold">${{ toMoney(summary?.available_commission) }}</div>
          </div>
          <div class="card p-4 distribution-card">
            <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('distribution.thisMonthCommission') }}</div>
            <div class="mt-1 text-lg font-semibold">${{ toMoney(summary?.this_month_commission) }}</div>
          </div>
          <div class="card p-4 distribution-card">
            <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('distribution.teamContribution') }}</div>
            <div class="mt-1 text-lg font-semibold">${{ toMoney(summary?.total_team_contribution) }}</div>
          </div>
        </div>

        <div v-if="summary?.source_stats?.length" class="card p-4 distribution-card">
          <div class="mb-3 text-sm font-semibold">{{ t('distribution.channelStats') }}</div>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="item in summary?.source_stats || []"
              :key="item.source"
              class="inline-flex items-center gap-2 rounded-full border border-indigo-200 px-3 py-1 text-xs dark:border-indigo-900/50"
            >
              <span class="font-medium uppercase">{{ item.source }}</span>
              <span class="rounded bg-indigo-100 px-1.5 py-0.5 text-indigo-700 dark:bg-indigo-900/40 dark:text-indigo-200">{{ item.count }}</span>
            </span>
          </div>
        </div>

        <div class="card p-4 distribution-card">
          <div class="mb-3 text-sm font-semibold">{{ t('distribution.weeklyGoals') }}</div>
          <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
            <div v-for="goal in goalCards" :key="goal.key" class="rounded-lg border border-indigo-100 p-3 dark:border-indigo-900/40">
              <div class="text-xs text-gray-500">{{ goal.label }}</div>
              <div class="mt-1 text-base font-semibold">{{ goal.current }} / {{ goal.target }}</div>
              <div class="mt-2 h-2 overflow-hidden rounded bg-gray-100 dark:bg-dark-700">
                <div class="h-full rounded bg-indigo-500" :style="{ width: `${goal.progress}%` }"></div>
              </div>
            </div>
          </div>
        </div>

        <div class="card p-4 distribution-card">
          <div class="mb-3 text-sm font-semibold">{{ t('distribution.promoKit') }}</div>
          <div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
            <button class="btn distribution-pill" @click="copyInviteMaterial('wechat')">{{ t('distribution.copyWeChatPitch') }}</button>
            <button class="btn distribution-pill" @click="copyInviteMaterial('group')">{{ t('distribution.copyGroupPitch') }}</button>
            <button class="btn distribution-pill" @click="copyInviteLink()">{{ t('distribution.copyInviteLink') }}</button>
          </div>
        </div>

        <div class="card p-4 distribution-card">
          <div class="mb-3 text-sm font-semibold">{{ t('distribution.bindInviteCode') }}</div>
          <div class="mb-3 text-sm text-gray-600 dark:text-gray-300">
            {{ t('distribution.inviterUserId') }}:
            <span class="font-medium">{{ profile?.inviter_user_id ?? t('distribution.unbound') }}</span>
          </div>
          <div class="flex flex-col gap-3 sm:flex-row">
            <input
              v-model="inviteCodeInput"
              type="text"
              class="input flex-1"
              :placeholder="t('distribution.bindPlaceholder')"
              :disabled="binding || !!profile?.inviter_user_id"
            />
            <button
              class="btn btn-primary"
              :disabled="binding || !!profile?.inviter_user_id || !inviteCodeInput.trim()"
              @click="bindInviter"
            >
              {{ binding ? t('distribution.binding') : t('distribution.bindButton') }}
            </button>
          </div>
        </div>

        <div class="card p-4 distribution-card">
          <div class="mb-3 text-sm font-semibold">{{ t('distribution.withdrawalTitle') }}</div>
          <div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
            <input v-model.number="withdrawForm.amount" class="input" type="number" min="0" step="0.01" :placeholder="t('distribution.amount')" />
            <input v-model="withdrawForm.account_type" class="input" type="text" :placeholder="t('distribution.accountType')" />
            <input v-model="withdrawForm.account_ref" class="input" type="text" :placeholder="t('distribution.accountRef')" />
            <input v-model="withdrawForm.notes" class="input" type="text" :placeholder="t('distribution.notes')" />
          </div>
          <div class="mt-3">
            <button class="btn btn-primary" :disabled="submittingWithdrawal" @click="submitWithdrawal">
              {{ submittingWithdrawal ? t('distribution.submitting') : t('distribution.submitWithdrawal') }}
            </button>
          </div>
        </div>

        <div class="card p-4 distribution-card">
          <div class="mb-3 text-sm font-semibold">{{ t('distribution.referralsTitle') }}</div>
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-gray-500 dark:text-gray-400">
                <th class="py-2">ID</th>
                <th class="py-2">Email</th>
                <th class="py-2">Username</th>
                <th class="py-2">{{ t('distribution.teamContribution') }}</th>
                <th class="py-2">Time</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in referrals.items" :key="item.user_id" class="border-t border-gray-100 dark:border-dark-800">
                <td class="py-2">{{ item.user_id }}</td>
                <td class="py-2">{{ item.email || '-' }}</td>
                <td class="py-2">{{ item.username || '-' }}</td>
                <td class="py-2">${{ toMoney(item.total_contribution) }}</td>
                <td class="py-2">{{ formatDateTime(item.bound_at) }}</td>
              </tr>
              <tr v-if="referrals.items.length === 0">
                <td colspan="5" class="py-4 text-center text-gray-500">{{ t('distribution.noData') }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="card p-4 distribution-card">
          <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
            <div class="text-sm font-semibold">{{ t('distribution.teamTitle') }}</div>
            <div class="flex items-center gap-2 text-xs">
              <button class="btn distribution-pill" :class="selectedTeamLevel === 1 ? 'btn-primary' : ''" @click="changeTeamLevel(1)">{{ t('distribution.level1Team') }}</button>
              <button class="btn distribution-pill" :class="selectedTeamLevel === 2 ? 'btn-primary' : ''" @click="changeTeamLevel(2)">{{ t('distribution.level2Team') }}</button>
            </div>
          </div>
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-gray-500 dark:text-gray-400">
                <th class="py-2">ID</th>
                <th class="py-2">Email</th>
                <th class="py-2">Username</th>
                <th class="py-2">{{ t('distribution.teamContribution') }}</th>
                <th class="py-2">{{ t('distribution.commissionsTitle') }}</th>
                <th class="py-2">Time</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in teamMembers.items" :key="`${item.team_level}-${item.user_id}`" class="border-t border-gray-100 dark:border-dark-800">
                <td class="py-2">{{ item.user_id }}</td>
                <td class="py-2">{{ item.email || '-' }}</td>
                <td class="py-2">{{ item.username || '-' }}</td>
                <td class="py-2">${{ toMoney(item.total_contribution) }}</td>
                <td class="py-2">${{ toMoney(item.commission_generated) }}</td>
                <td class="py-2">{{ formatDateTime(item.bound_at) }}</td>
              </tr>
              <tr v-if="teamMembers.items.length === 0">
                <td colspan="6" class="py-4 text-center text-gray-500">{{ t('distribution.noData') }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="card p-4 distribution-card">
          <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
            <div class="text-sm font-semibold">{{ t('distribution.commissionsTitle') }}</div>
            <div class="flex items-center gap-2 text-xs">
              <button class="btn distribution-pill" :class="selectedCommissionLevel === 'all' ? 'btn-primary' : ''" @click="changeCommissionLevel('all')">{{ t('distribution.all') }}</button>
              <button class="btn distribution-pill" :class="selectedCommissionLevel === 1 ? 'btn-primary' : ''" @click="changeCommissionLevel(1)">{{ t('distribution.level1Team') }}</button>
              <button class="btn distribution-pill" :class="selectedCommissionLevel === 2 ? 'btn-primary' : ''" @click="changeCommissionLevel(2)">{{ t('distribution.level2Team') }}</button>
            </div>
          </div>
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-gray-500 dark:text-gray-400">
                <th class="py-2">ID</th>
                <th class="py-2">Invitee</th>
                <th class="py-2">Topup</th>
                <th class="py-2">Rate</th>
                <th class="py-2">Commission</th>
                <th class="py-2">L</th>
                <th class="py-2">Time</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in commissions.items" :key="item.id" class="border-t border-gray-100 dark:border-dark-800">
                <td class="py-2">{{ item.id }}</td>
                <td class="py-2">{{ item.invitee_email || item.invitee_user_id }}</td>
                <td class="py-2">${{ toMoney(item.topup_amount) }}</td>
                <td class="py-2">{{ (item.commission_rate * 100).toFixed(2) }}%</td>
                <td class="py-2">${{ toMoney(item.commission_amount) }}</td>
                <td class="py-2">{{ item.commission_level }}</td>
                <td class="py-2">{{ formatDateTime(item.created_at) }}</td>
              </tr>
              <tr v-if="commissions.items.length === 0">
                <td colspan="7" class="py-4 text-center text-gray-500">{{ t('distribution.noData') }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="card p-4 distribution-card">
          <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
            <div class="text-sm font-semibold">{{ t('distribution.withdrawalsTitle') }}</div>
            <div class="flex items-center gap-2 text-xs">
              <button class="btn distribution-pill" :class="selectedWithdrawalStatus === 'all' ? 'btn-primary' : ''" @click="changeWithdrawalStatus('all')">{{ t('distribution.all') }}</button>
              <button class="btn distribution-pill" :class="selectedWithdrawalStatus === 'pending' ? 'btn-primary' : ''" @click="changeWithdrawalStatus('pending')">{{ t('distribution.statusPending') }}</button>
              <button class="btn distribution-pill" :class="selectedWithdrawalStatus === 'approved' ? 'btn-primary' : ''" @click="changeWithdrawalStatus('approved')">{{ t('distribution.statusApproved') }}</button>
              <button class="btn distribution-pill" :class="selectedWithdrawalStatus === 'rejected' ? 'btn-primary' : ''" @click="changeWithdrawalStatus('rejected')">{{ t('distribution.statusRejected') }}</button>
            </div>
          </div>
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-gray-500 dark:text-gray-400">
                <th class="py-2">ID</th>
                <th class="py-2">{{ t('distribution.amount') }}</th>
                <th class="py-2">{{ t('distribution.accountType') }}</th>
                <th class="py-2">{{ t('distribution.accountRef') }}</th>
                <th class="py-2">Status</th>
                <th class="py-2">Time</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in withdrawals.items" :key="item.id" class="border-t border-gray-100 dark:border-dark-800">
                <td class="py-2">{{ item.id }}</td>
                <td class="py-2">${{ toMoney(item.amount) }}</td>
                <td class="py-2">{{ item.account_type || '-' }}</td>
                <td class="py-2">{{ item.account_ref }}</td>
                <td class="py-2">
                  <span class="rounded px-2 py-1 text-xs" :class="statusClass(item.status)">
                    {{ statusLabel(item.status) }}
                  </span>
                </td>
                <td class="py-2">{{ formatDateTime(item.created_at) }}</td>
              </tr>
              <tr v-if="withdrawals.items.length === 0">
                <td colspan="6" class="py-4 text-center text-gray-500">{{ t('distribution.noData') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { userAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import { resolveDistributionWithdrawalErrorMessage } from '@/utils/distributionWithdrawalError'
import type {
  DistributionCommissionRecord,
  DistributionProfile,
  DistributionReferral,
  DistributionSummary,
  DistributionTeamMember,
  DistributionWithdrawalRequest
} from '@/api/user'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const binding = ref(false)
const submittingWithdrawal = ref(false)

const profile = ref<DistributionProfile | null>(null)
const summary = ref<DistributionSummary | null>(null)

const referrals = ref<{ items: DistributionReferral[] }>({ items: [] })
const teamMembers = ref<{ items: DistributionTeamMember[] }>({ items: [] })
const commissions = ref<{ items: DistributionCommissionRecord[] }>({ items: [] })
const withdrawals = ref<{ items: DistributionWithdrawalRequest[] }>({ items: [] })

const selectedTeamLevel = ref<1 | 2>(1)
const selectedCommissionLevel = ref<'all' | 1 | 2>('all')
const selectedWithdrawalStatus = ref<'all' | 'pending' | 'approved' | 'rejected'>('all')

const inviteCodeInput = ref('')

const withdrawForm = ref({
  amount: 0,
  account_type: '',
  account_ref: '',
  notes: ''
})

const toMoney = (value?: number) => Number(value || 0).toFixed(2)

const goalCards = computed(() => {
  const referralCurrent = Number(profile.value?.total_referrals || 0)
  const referralTarget = 20
  const contributionCurrent = Number(summary.value?.total_team_contribution || 0)
  const contributionTarget = 5000
  const withdrawCurrent = Number(summary.value?.available_commission || 0)
  const withdrawTarget = 1000

  const build = (key: string, label: string, current: number, target: number) => ({
    key,
    label,
    current: Math.floor(current),
    target,
    progress: Math.min(100, Math.round((current / Math.max(1, target)) * 100))
  })

  return [
    build('referrals', t('distribution.goalReferrals'), referralCurrent, referralTarget),
    build('contribution', t('distribution.goalContribution'), contributionCurrent, contributionTarget),
    build('withdrawable', t('distribution.goalWithdrawable'), withdrawCurrent, withdrawTarget)
  ]
})

const buildInviteLink = (source?: string) => {
  const code = profile.value?.invite_code || ''
  if (!code) return ''
  const url = new URL('/register', window.location.origin)
  url.searchParams.set('invite_code', code)
  if (source) {
    url.searchParams.set('src', source)
  }
  return url.toString()
}

const copyText = async (text: string, successKey: string) => {
  if (!text) {
    appStore.showError(t('distribution.loadFailed'))
    return
  }
  try {
    await navigator.clipboard.writeText(text)
    appStore.showSuccess(t(successKey))
  } catch (error) {
    appStore.showError(t('errors.somethingWentWrong'))
  }
}

const copyInviteMaterial = async (channel: 'wechat' | 'group') => {
  const link = buildInviteLink(channel)
  const pitch = channel === 'wechat'
    ? t('distribution.wechatPitchTemplate', { link })
    : t('distribution.groupPitchTemplate', { link })
  await copyText(pitch, 'distribution.copySuccess')
}

const copyInviteLink = async () => {
  await copyText(buildInviteLink('direct'), 'distribution.copySuccess')
}

const statusLabel = (status: string) => {
  if (status === 'approved') return t('distribution.statusApproved')
  if (status === 'rejected') return t('distribution.statusRejected')
  return t('distribution.statusPending')
}

const statusClass = (status: string) => {
  if (status === 'approved') return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
  if (status === 'rejected') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300'
}

const withdrawalErrorMessage = (error: unknown) => resolveDistributionWithdrawalErrorMessage(error, t)

const loadCore = async () => {
  const [p, s, r] = await Promise.all([
    userAPI.getDistributionProfile(),
    userAPI.getDistributionSummary(),
    userAPI.listDistributionReferrals(1, 10)
  ])
  profile.value = p
  summary.value = s
  referrals.value = { items: r.items || [] }
}

const loadTeam = async () => {
  const team = await userAPI.listDistributionTeam(selectedTeamLevel.value, 1, 10)
  teamMembers.value = { items: team.items || [] }
}

const loadCommissions = async () => {
  const level = selectedCommissionLevel.value === 'all' ? undefined : selectedCommissionLevel.value
  const records = await userAPI.listDistributionCommissions(1, 10, level)
  commissions.value = { items: records.items || [] }
}

const loadWithdrawals = async () => {
  const status = selectedWithdrawalStatus.value === 'all' ? undefined : selectedWithdrawalStatus.value
  const records = await userAPI.listDistributionWithdrawals(1, 10, status)
  withdrawals.value = { items: records.items || [] }
}

const loadAll = async () => {
  loading.value = true
  try {
    await Promise.all([loadCore(), loadTeam(), loadCommissions(), loadWithdrawals()])
  } catch (error) {
    appStore.showError(t('distribution.loadFailed'))
  } finally {
    loading.value = false
  }
}

const bindInviter = async () => {
  const code = inviteCodeInput.value.trim()
  if (!code) return
  binding.value = true
  try {
    await userAPI.bindDistributionInviter(code)
    appStore.showSuccess(t('distribution.bindSuccess'))
    inviteCodeInput.value = ''
    await loadAll()
  } catch (error: any) {
    appStore.showError(error?.message || t('distribution.loadFailed'))
  } finally {
    binding.value = false
  }
}

const submitWithdrawal = async () => {
  submittingWithdrawal.value = true
  try {
    await userAPI.createDistributionWithdrawal({
      amount: Number(withdrawForm.value.amount || 0),
      account_type: withdrawForm.value.account_type.trim(),
      account_ref: withdrawForm.value.account_ref.trim(),
      notes: withdrawForm.value.notes.trim()
    })
    appStore.showSuccess(t('distribution.withdrawalSuccess'))
    withdrawForm.value = { amount: 0, account_type: '', account_ref: '', notes: '' }
    await loadAll()
  } catch (error: any) {
    appStore.showError(withdrawalErrorMessage(error))
  } finally {
    submittingWithdrawal.value = false
  }
}

const changeTeamLevel = async (level: 1 | 2) => {
  if (selectedTeamLevel.value === level) return
  selectedTeamLevel.value = level
  try {
    await loadTeam()
  } catch (error) {
    appStore.showError(t('distribution.loadFailed'))
  }
}

const changeCommissionLevel = async (level: 'all' | 1 | 2) => {
  if (selectedCommissionLevel.value === level) return
  selectedCommissionLevel.value = level
  try {
    await loadCommissions()
  } catch (error) {
    appStore.showError(t('distribution.loadFailed'))
  }
}

const changeWithdrawalStatus = async (status: 'all' | 'pending' | 'approved' | 'rejected') => {
  if (selectedWithdrawalStatus.value === status) return
  selectedWithdrawalStatus.value = status
  try {
    await loadWithdrawals()
  } catch (error) {
    appStore.showError(t('distribution.loadFailed'))
  }
}

onMounted(() => {
  loadAll()
})
</script>

<style scoped>
.distribution-hero {
  border-radius: 16px;
  color: #eef2ff;
  background:
    radial-gradient(1200px 280px at -10% 20%, rgba(34, 211, 238, 0.25), transparent 60%),
    radial-gradient(600px 260px at 110% 20%, rgba(99, 102, 241, 0.35), transparent 65%),
    linear-gradient(135deg, #1d2b64 0%, #283593 45%, #3949ab 100%);
  border: 1px solid rgba(129, 140, 248, 0.35);
}

.distribution-hero-kicker {
  color: rgba(191, 219, 254, 0.95);
}

.distribution-card {
  border: 1px solid rgba(129, 140, 248, 0.16);
  box-shadow: 0 6px 22px rgba(15, 23, 42, 0.06);
}

.distribution-brand .btn-primary {
  background: linear-gradient(135deg, #4f46e5, #2563eb);
  border-color: transparent;
}

.distribution-brand .distribution-pill {
  border-color: rgba(99, 102, 241, 0.28);
}
</style>
