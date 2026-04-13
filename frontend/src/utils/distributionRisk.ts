export type WithdrawalRiskLevel = 'low' | 'medium' | 'high'

export interface WithdrawalRiskContext {
  amount: number
  riskThreshold: number
  dailyLimitAmount: number
}

const safeNumber = (value: number | null | undefined): number => {
  const n = Number(value)
  return Number.isFinite(n) ? n : 0
}

export const evaluateWithdrawalRiskLevel = ({ amount, riskThreshold, dailyLimitAmount }: WithdrawalRiskContext): WithdrawalRiskLevel => {
  const normalizedAmount = safeNumber(amount)
  const normalizedRiskThreshold = safeNumber(riskThreshold)
  const normalizedDailyLimitAmount = safeNumber(dailyLimitAmount)

  if (normalizedDailyLimitAmount > 0 && normalizedAmount >= normalizedDailyLimitAmount) {
    return 'high'
  }

  if (normalizedRiskThreshold > 0 && normalizedAmount >= normalizedRiskThreshold * 2) {
    return 'high'
  }

  if (normalizedRiskThreshold > 0 && normalizedAmount >= normalizedRiskThreshold) {
    return 'medium'
  }

  return 'low'
}

export const buildDistributionReviewNote = (reasonTag: string, note: string): string => {
  const trimmedTag = reasonTag.trim()
  const trimmedNote = note.trim()

  if (!trimmedTag && !trimmedNote) return ''
  if (!trimmedTag) return trimmedNote
  if (!trimmedNote) return `[${trimmedTag}]`
  return `[${trimmedTag}] ${trimmedNote}`
}
