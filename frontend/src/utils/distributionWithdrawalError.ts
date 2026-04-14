type Translator = (key: string) => string

const readPath = (input: any, path: string[]): unknown => {
  let current = input
  for (const key of path) {
    if (current == null || typeof current !== 'object' || !(key in current)) {
      return undefined
    }
    current = current[key]
  }
  return current
}

const DISTRIBUTION_WITHDRAWAL_REASON_MAP: Record<string, string> = {
  DISTRIBUTION_WITHDRAWAL_COOLDOWN: 'DISTRIBUTION_WITHDRAWAL_COOLDOWN',
  DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT: 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT',
  DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_COUNT: 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT',
  DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT: 'DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT',
  DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_AMOUNT: 'DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT'
}

const tryResolveKnownDistributionReason = (value: unknown): string => {
  if (typeof value !== 'string') return ''
  const trimmed = value.trim()
  if (!trimmed) return ''

  const normalized = trimmed.toUpperCase()
  if (DISTRIBUTION_WITHDRAWAL_REASON_MAP[normalized]) {
    return DISTRIBUTION_WITHDRAWAL_REASON_MAP[normalized]
  }

  const lower = trimmed.toLowerCase()
  if (lower.includes('distribution_withdrawal_daily_amount_limit') || lower.includes('distribution_withdrawal_daily_limit_amount')) {
    return 'DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT'
  }
  if (lower.includes('distribution_withdrawal_daily_limit_count') || lower.includes('distribution_withdrawal_daily_limit')) {
    return 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT'
  }
  if (lower.includes('distribution_withdrawal_cooldown')) {
    return 'DISTRIBUTION_WITHDRAWAL_COOLDOWN'
  }

  return ''
}

export const extractDistributionWithdrawalReason = (error: unknown): string => {
  const reasonOrCodeFields = [
    readPath(error, ['reason']),
    readPath(error, ['code']),
    readPath(error, ['data', 'reason']),
    readPath(error, ['data', 'code']),
    readPath(error, ['data', 'error', 'reason']),
    readPath(error, ['data', 'error', 'code']),
    readPath(error, ['response', 'data', 'reason']),
    readPath(error, ['response', 'data', 'code']),
    readPath(error, ['response', 'data', 'error', 'reason']),
    readPath(error, ['response', 'data', 'error', 'code'])
  ]

  for (const field of reasonOrCodeFields) {
    const knownReason = tryResolveKnownDistributionReason(field)
    if (knownReason) return knownReason
  }

  const fallbackTextFields = [
    readPath(error, ['error']),
    readPath(error, ['data', 'error']),
    readPath(error, ['response', 'data', 'error']),
    readPath(error, ['message']),
    readPath(error, ['data', 'message']),
    readPath(error, ['response', 'data', 'message']),
    readPath(error, ['data', 'error', 'message']),
    readPath(error, ['response', 'data', 'error', 'message'])
  ]

  for (const field of fallbackTextFields) {
    const knownReason = tryResolveKnownDistributionReason(field)
    if (knownReason) return knownReason
  }

  const candidate = reasonOrCodeFields.find((value) => typeof value === 'string' && String(value).trim() !== '')
  return typeof candidate === 'string' ? candidate.trim().toUpperCase() : ''
}

const extractDistributionWithdrawalMessage = (error: unknown): string => {
  const candidate = [
    readPath(error, ['message']),
    readPath(error, ['data', 'message']),
    readPath(error, ['response', 'data', 'message']),
    readPath(error, ['data', 'error', 'message']),
    readPath(error, ['response', 'data', 'error', 'message'])
  ].find((value) => typeof value === 'string' && String(value).trim() !== '')

  return typeof candidate === 'string' ? candidate.trim() : ''
}

export const resolveDistributionWithdrawalErrorMessage = (
  error: unknown,
  t: Translator,
  fallbackKey = 'distribution.loadFailed'
): string => {
  const reason = extractDistributionWithdrawalReason(error)

  if (reason === 'DISTRIBUTION_WITHDRAWAL_COOLDOWN') return t('distribution.withdrawalErrors.cooldown')
  if (reason === 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT' || reason === 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_COUNT') {
    return t('distribution.withdrawalErrors.dailyLimitCount')
  }
  if (
    reason === 'DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT'
    || reason === 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_AMOUNT'
  ) {
    return t('distribution.withdrawalErrors.dailyLimitAmount')
  }

  return extractDistributionWithdrawalMessage(error) || t(fallbackKey)
}
