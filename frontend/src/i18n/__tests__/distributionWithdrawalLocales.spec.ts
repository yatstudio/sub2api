import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('distribution withdrawal risk error locale keys', () => {
  it('contains readable zh messages', () => {
    expect(zh.distribution.withdrawalErrors.cooldown).toBe('当前处于提现冷却期，请稍后再试')
    expect(zh.distribution.withdrawalErrors.dailyLimitCount).toBe('今日提现次数已达上限')
    expect(zh.distribution.withdrawalErrors.dailyLimitAmount).toBe('今日提现金额已达上限')
  })

  it('contains readable en messages', () => {
    expect(en.distribution.withdrawalErrors.cooldown).toBe('You are still in the withdrawal cooldown period. Please try again later.')
    expect(en.distribution.withdrawalErrors.dailyLimitCount).toBe('Daily withdrawal request count limit reached.')
    expect(en.distribution.withdrawalErrors.dailyLimitAmount).toBe('Daily withdrawal amount limit reached.')
  })
})
