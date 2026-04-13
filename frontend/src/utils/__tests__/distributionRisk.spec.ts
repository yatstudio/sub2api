import { describe, expect, it } from 'vitest'
import { buildDistributionReviewNote, evaluateWithdrawalRiskLevel } from '@/utils/distributionRisk'

describe('distributionRisk', () => {
  it('marks high risk when amount reaches daily limit', () => {
    expect(evaluateWithdrawalRiskLevel({ amount: 1200, riskThreshold: 800, dailyLimitAmount: 1200 })).toBe('high')
  })

  it('marks high risk when amount is 2x threshold', () => {
    expect(evaluateWithdrawalRiskLevel({ amount: 1600, riskThreshold: 800, dailyLimitAmount: 5000 })).toBe('high')
  })

  it('marks medium risk when amount reaches threshold but below high rules', () => {
    expect(evaluateWithdrawalRiskLevel({ amount: 900, riskThreshold: 800, dailyLimitAmount: 5000 })).toBe('medium')
  })

  it('marks low risk when below threshold', () => {
    expect(evaluateWithdrawalRiskLevel({ amount: 300, riskThreshold: 800, dailyLimitAmount: 5000 })).toBe('low')
  })

  it('builds review note with tag and note', () => {
    expect(buildDistributionReviewNote('manual_check', 'validated by ops')).toBe('[manual_check] validated by ops')
  })

  it('builds review note from tag only', () => {
    expect(buildDistributionReviewNote('risk_flag', '   ')).toBe('[risk_flag]')
  })
})
