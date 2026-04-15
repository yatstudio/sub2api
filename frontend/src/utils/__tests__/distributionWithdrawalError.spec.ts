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

  it('extracts reason from response.data.code payload shape', () => {
    const reason = extractDistributionWithdrawalReason({
      response: {
        data: {
          code: 'distribution_withdrawal_daily_amount_limit'
        }
      }
    })

    expect(reason).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT')

    const canonicalCountReason = extractDistributionWithdrawalReason({
      response: {
        data: {
          code: 'distribution_withdrawal_daily_limit'
        }
      }
    })

    expect(canonicalCountReason).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT')
  })

  it('extracts reason from top-level data.code payload shape', () => {
    const reason = extractDistributionWithdrawalReason({
      data: {
        code: 'distribution_withdrawal_daily_limit_count'
      }
    })

    expect(reason).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT')

    const canonicalCountReason = extractDistributionWithdrawalReason({
      data: {
        code: 'distribution_withdrawal_daily_limit'
      }
    })

    expect(canonicalCountReason).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT')
  })

  it('extracts reason from top-level code payload shape', () => {
    const reason = extractDistributionWithdrawalReason({
      code: 'distribution_withdrawal_daily_limit_amount'
    })

    expect(reason).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT')

    const canonicalCountReason = extractDistributionWithdrawalReason({
      code: 'distribution_withdrawal_daily_limit'
    })

    expect(canonicalCountReason).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT')
  })

  it('extracts reason from top-level error object reason/code payload shape', () => {
    expect(extractDistributionWithdrawalReason({
      error: { reason: 'distribution_withdrawal_cooldown' }
    })).toBe('DISTRIBUTION_WITHDRAWAL_COOLDOWN')

    expect(extractDistributionWithdrawalReason({
      error: { reason: 'distribution_withdrawal_daily_limit' }
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT')

    expect(extractDistributionWithdrawalReason({
      error: { reason: 'distribution_withdrawal_daily_limit_amount' }
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT')

    expect(extractDistributionWithdrawalReason({
      error: { code: 'distribution_withdrawal_daily_limit_amount' }
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT')

    expect(extractDistributionWithdrawalReason({
      error: { code: 'distribution_withdrawal_daily_limit' }
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT')
  })

  it('extracts known reason from top-level error.message token payload shape', () => {
    expect(extractDistributionWithdrawalReason({
      error: { message: 'request rejected: distribution_withdrawal_daily_limit_count' }
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT')

    expect(extractDistributionWithdrawalReason({
      error: { message: 'request rejected: distribution_withdrawal_daily_limit_amount' }
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT')
  })

  it('extracts reason from nested top-level error.error payload shape', () => {
    expect(extractDistributionWithdrawalReason({
      error: {
        error: { reason: 'distribution_withdrawal_cooldown' }
      }
    })).toBe('DISTRIBUTION_WITHDRAWAL_COOLDOWN')

    expect(extractDistributionWithdrawalReason({
      error: {
        error: { code: 'distribution_withdrawal_daily_limit' }
      }
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT')

    expect(extractDistributionWithdrawalReason({
      error: {
        error: { message: 'request rejected: distribution_withdrawal_daily_limit_amount' }
      }
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT')
  })

  it('extracts reason when backend returns plain string error code', () => {
    expect(extractDistributionWithdrawalReason({ error: 'distribution_withdrawal_cooldown' }))
      .toBe('DISTRIBUTION_WITHDRAWAL_COOLDOWN')
    expect(extractDistributionWithdrawalReason({ response: { data: { error: 'distribution_withdrawal_daily_amount_limit' } } }))
      .toBe('DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT')
  })

  it('extracts known reason from free-form message text', () => {
    expect(extractDistributionWithdrawalReason({
      response: { data: { message: 'request rejected: distribution_withdrawal_daily_limit_count' } }
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT')

    expect(extractDistributionWithdrawalReason({
      data: { message: 'request rejected: distribution_withdrawal_daily_limit_amount' }
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT')

    expect(extractDistributionWithdrawalReason({
      message: 'request rejected: distribution_withdrawal_cooldown'
    })).toBe('DISTRIBUTION_WITHDRAWAL_COOLDOWN')

    expect(extractDistributionWithdrawalReason({
      message: 'request rejected: distribution_withdrawal_daily_limit_count'
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT')

    expect(extractDistributionWithdrawalReason({
      message: 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_AMOUNT: cap reached'
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT')
  })

  it('extracts known reasons from nested response.data.error.message payload', () => {
    expect(extractDistributionWithdrawalReason({
      response: { data: { error: { message: 'blocked by distribution_withdrawal_cooldown' } } }
    })).toBe('DISTRIBUTION_WITHDRAWAL_COOLDOWN')

    expect(extractDistributionWithdrawalReason({
      response: { data: { error: { message: 'hit distribution_withdrawal_daily_limit_count' } } }
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT')

    expect(extractDistributionWithdrawalReason({
      response: { data: { error: { message: 'risk cap: distribution_withdrawal_daily_amount_limit' } } }
    })).toBe('DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT')
  })

  it('prefers structured reason/code over generic error string', () => {
    const reason = extractDistributionWithdrawalReason({
      error: 'request failed',
      response: {
        data: {
          error: {
            code: 'distribution_withdrawal_daily_limit'
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

    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_COUNT' }, t))
      .toBe('t:distribution.withdrawalErrors.dailyLimitCount')

    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT' }, t))
      .toBe('t:distribution.withdrawalErrors.dailyLimitAmount')

    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_AMOUNT' }, t))
      .toBe('t:distribution.withdrawalErrors.dailyLimitAmount')
  })

  it('falls back to backend message then fallback key', () => {
    expect(resolveDistributionWithdrawalErrorMessage({ response: { data: { message: 'raw backend msg' } } }, t))
      .toBe('raw backend msg')
    expect(resolveDistributionWithdrawalErrorMessage({ data: { message: 'raw backend msg from data' } }, t))
      .toBe('raw backend msg from data')
    expect(resolveDistributionWithdrawalErrorMessage({}, t))
      .toBe('t:distribution.loadFailed')
  })
})
