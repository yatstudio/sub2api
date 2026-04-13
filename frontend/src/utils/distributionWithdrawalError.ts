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

export const extractDistributionWithdrawalReason = (error: unknown): string => {
  const candidate = [
    readPath(error, ['reason']),
    readPath(error, ['code']),
    readPath(error, ['response', 'data', 'reason']),
    readPath(error, ['response', 'data', 'code']),
    readPath(error, ['response', 'data', 'error', 'reason']),
    readPath(error, ['response', 'data', 'error', 'code']),
    // Keep generic error string as lower-priority fallback.
    readPath(error, ['error']),
    readPath(error, ['response', 'data', 'error'])
  ].find((value) => typeof value === 'string' && String(value).trim() !== '')

  return typeof candidate === 'string' ? candidate.trim().toUpperCase() : ''
}

const extractDistributionWithdrawalMessage = (error: unknown): string => {
  const candidate = [
    readPath(error, ['message']),
    readPath(error, ['response', 'data', 'message']),
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
  if (reason === 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT') return t('distribution.withdrawalErrors.dailyLimitCount')
  if (reason === 'DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT') return t('distribution.withdrawalErrors.dailyLimitAmount')

  return extractDistributionWithdrawalMessage(error) || t(fallbackKey)
}
