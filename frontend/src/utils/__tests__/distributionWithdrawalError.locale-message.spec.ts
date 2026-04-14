import { describe, expect, it } from 'vitest'
import zh from '@/i18n/locales/zh'
import en from '@/i18n/locales/en'
import { resolveDistributionWithdrawalErrorMessage } from '@/utils/distributionWithdrawalError'

const translatorFromLocale = (locale: any) => (key: string): string => {
  if (key === 'distribution.withdrawalErrors.cooldown') return locale.distribution.withdrawalErrors.cooldown
  if (key === 'distribution.withdrawalErrors.dailyLimitCount') return locale.distribution.withdrawalErrors.dailyLimitCount
  if (key === 'distribution.withdrawalErrors.dailyLimitAmount') return locale.distribution.withdrawalErrors.dailyLimitAmount
  return `missing:${key}`
}

describe('distributionWithdrawalError locale message alignment', () => {
  it('returns readable zh messages for cooldown/daily count/daily amount limits', () => {
    const t = translatorFromLocale(zh)

    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_COOLDOWN' }, t))
      .toBe('当前处于提现冷却期，请稍后再试')
    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT' }, t))
      .toBe('今日提现次数已达上限')
    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_COUNT' }, t))
      .toBe('今日提现次数已达上限')
    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT' }, t))
      .toBe('今日提现金额已达上限')
    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_AMOUNT' }, t))
      .toBe('今日提现金额已达上限')
    expect(resolveDistributionWithdrawalErrorMessage({
      response: { data: { error: { message: 'request blocked: distribution_withdrawal_daily_limit_count' } } }
    }, t)).toBe('今日提现次数已达上限')
    expect(resolveDistributionWithdrawalErrorMessage({
      response: { data: { error: { message: 'request blocked: distribution_withdrawal_daily_limit_amount' } } }
    }, t)).toBe('今日提现金额已达上限')
    expect(resolveDistributionWithdrawalErrorMessage({
      data: { error: { message: 'request blocked: distribution_withdrawal_daily_limit_count' } }
    }, t)).toBe('今日提现次数已达上限')
    expect(resolveDistributionWithdrawalErrorMessage({
      data: { error: { message: 'request blocked: distribution_withdrawal_daily_limit_amount' } }
    }, t)).toBe('今日提现金额已达上限')
    expect(resolveDistributionWithdrawalErrorMessage({
      data: { message: 'request blocked: distribution_withdrawal_daily_limit_count' }
    }, t)).toBe('今日提现次数已达上限')
    expect(resolveDistributionWithdrawalErrorMessage({
      data: { message: 'request blocked: distribution_withdrawal_daily_limit_amount' }
    }, t)).toBe('今日提现金额已达上限')
  })

  it('returns readable en messages for cooldown/daily count/daily amount limits', () => {
    const t = translatorFromLocale(en)

    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_COOLDOWN' }, t))
      .toBe('You are still in the withdrawal cooldown period. Please try again later.')
    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT' }, t))
      .toBe('Daily withdrawal request count limit reached.')
    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_COUNT' }, t))
      .toBe('Daily withdrawal request count limit reached.')
    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT' }, t))
      .toBe('Daily withdrawal amount limit reached.')
    expect(resolveDistributionWithdrawalErrorMessage({ reason: 'DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_AMOUNT' }, t))
      .toBe('Daily withdrawal amount limit reached.')
    expect(resolveDistributionWithdrawalErrorMessage({
      response: { data: { error: { message: 'request blocked: distribution_withdrawal_daily_limit_count' } } }
    }, t)).toBe('Daily withdrawal request count limit reached.')
    expect(resolveDistributionWithdrawalErrorMessage({
      response: { data: { error: { message: 'request blocked: distribution_withdrawal_daily_limit_amount' } } }
    }, t)).toBe('Daily withdrawal amount limit reached.')
    expect(resolveDistributionWithdrawalErrorMessage({
      data: { error: { message: 'request blocked: distribution_withdrawal_daily_limit_count' } }
    }, t)).toBe('Daily withdrawal request count limit reached.')
    expect(resolveDistributionWithdrawalErrorMessage({
      data: { error: { message: 'request blocked: distribution_withdrawal_daily_limit_amount' } }
    }, t)).toBe('Daily withdrawal amount limit reached.')
    expect(resolveDistributionWithdrawalErrorMessage({
      data: { message: 'request blocked: distribution_withdrawal_daily_limit_count' }
    }, t)).toBe('Daily withdrawal request count limit reached.')
    expect(resolveDistributionWithdrawalErrorMessage({
      data: { message: 'request blocked: distribution_withdrawal_daily_limit_amount' }
    }, t)).toBe('Daily withdrawal amount limit reached.')
  })
})
