import { describe, expect, it } from 'vitest'
import {
  extractDistributionWithdrawalReason,
  resolveDistributionWithdrawalErrorMessage
} from '@/utils/distributionWithdrawalError'

const t = (key: string) => `t:${key}`

describe('distributionWithdrawalError', () => {
  it('extracts reason from nested api error payload', () => {
    const reason = extractDistributionWithdrawalReason({
      response: {
        data: {
          error: {
            reason: 'distribution_withdrawal_daily_limit'
          }
        }
      }
    })

    expect(reason).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT')
  })

  it('maps cooldown reason to readable i18n message', () => {
    const message = resolveDistributionWithdrawalErrorMessage({
      response: {
        data: {
          error: {
            code: 'DISTRIBUTION_WITHDRAWAL_COOLDOWN'
          }
        }
      }
    }, t)

    expect(message).toBe('t:distribution.withdrawalErrors.cooldown')
  })

  it('maps daily count and amount reasons to specific messages', () => {
    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT' }, t))
      .toBe('t:distribution.withdrawalErrors.dailyLimitCount')

    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT' }, t))
      .toBe('t:distribution.withdrawalErrors.dailyLimitAmount')
  })

  it('falls back to backend message then fallback key', () => {
    expect(resolveDistributionWithdrawalErrorMessage({ response: { data: { message: 'raw backend msg' } } }, t))
      .toBe('raw backend msg')
    expect(resolveDistributionWithdrawalErrorMessage({}, t))
      .toBe('t:distribution.loadFailed')
  })
})
