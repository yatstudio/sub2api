/**
 * User API endpoints
 * Handles user profile management and password changes
 */

import { apiClient } from './client'
import type { User, ChangePasswordRequest, PaginatedResponse } from '@/types'

/**
 * Get current user profile
 * @returns User profile data
 */
export async function getProfile(): Promise<User> {
  const { data } = await apiClient.get<User>('/user/profile')
  return data
}

/**
 * Update current user profile
 * @param profile - Profile data to update
 * @returns Updated user profile data
 */
export async function updateProfile(profile: {
  username?: string
}): Promise<User> {
  const { data } = await apiClient.put<User>('/user', profile)
  return data
}

/**
 * Change current user password
 * @param passwords - Old and new password
 * @returns Success message
 */
export async function changePassword(
  oldPassword: string,
  newPassword: string
): Promise<{ message: string }> {
  const payload: ChangePasswordRequest = {
    old_password: oldPassword,
    new_password: newPassword
  }

  const { data } = await apiClient.put<{ message: string }>('/user/password', payload)
  return data
}

export interface DistributionProfile {
  user_id: number
  inviter_user_id: number | null
  invite_code: string
  commission_rate: number
  total_referrals: number
  total_commission_earned: number
  total_referral_contribution: number
  created_at: string
  updated_at: string
}

export interface DistributionSourceStat {
  source: string
  material?: string
  version?: string
  count: number
}

export interface DistributionSummary {
  user_id: number
  invite_code: string
  total_commission_earned: number
  total_commission_withdrawn: number
  pending_withdrawal_amount: number
  available_commission: number
  this_month_commission: number
  level1_team_count: number
  level2_team_count: number
  total_team_contribution: number
  source_stats?: DistributionSourceStat[]
}

export interface DistributionReferral {
  user_id: number
  email: string
  username: string
  bound_at: string
  total_contribution: number
}

export interface DistributionTeamMember {
  user_id: number
  email: string
  username: string
  bound_at: string
  total_contribution: number
  commission_generated: number
  team_level: number
}

export interface DistributionCommissionRecord {
  id: number
  inviter_user_id: number
  invitee_user_id: number
  invitee_email: string
  invitee_username: string
  topup_amount: number
  commission_rate: number
  commission_amount: number
  commission_level: number
  notes?: string
  created_at: string
}

export interface DistributionWithdrawalRequest {
  id: number
  user_id: number
  amount: number
  account_type: string
  account_ref: string
  status: 'pending' | 'approved' | 'rejected'
  notes?: string
  review_note?: string
  reviewed_by_user_id?: number | null
  reviewed_at?: string | null
  created_at: string
  updated_at: string
}

export async function getDistributionProfile(): Promise<DistributionProfile> {
  const { data } = await apiClient.get<DistributionProfile>('/user/distribution/profile')
  return data
}

export async function getDistributionSummary(): Promise<DistributionSummary> {
  const { data } = await apiClient.get<DistributionSummary>('/user/distribution/summary')
  return data
}

export async function bindDistributionInviter(inviteCode: string): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>('/user/distribution/bind', {
    invite_code: inviteCode
  })
  return data
}

export async function listDistributionReferrals(
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<DistributionReferral>> {
  const { data } = await apiClient.get<PaginatedResponse<DistributionReferral>>('/user/distribution/referrals', {
    params: { page, page_size: pageSize }
  })
  return data
}

export async function listDistributionTeam(
  level: 1 | 2,
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<DistributionTeamMember>> {
  const { data } = await apiClient.get<PaginatedResponse<DistributionTeamMember>>('/user/distribution/team', {
    params: { level, page, page_size: pageSize }
  })
  return data
}

export async function listDistributionCommissions(
  page: number = 1,
  pageSize: number = 20,
  level?: 1 | 2
): Promise<PaginatedResponse<DistributionCommissionRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<DistributionCommissionRecord>>('/user/distribution/commissions', {
    params: { level, page, page_size: pageSize }
  })
  return data
}

export async function createDistributionWithdrawal(payload: {
  amount: number
  account_type: string
  account_ref: string
  notes?: string
}): Promise<DistributionWithdrawalRequest> {
  const { data } = await apiClient.post<DistributionWithdrawalRequest>('/user/distribution/withdrawals', payload)
  return data
}

export async function listDistributionWithdrawals(
  page: number = 1,
  pageSize: number = 20,
  status?: 'pending' | 'approved' | 'rejected'
): Promise<PaginatedResponse<DistributionWithdrawalRequest>> {
  const { data } = await apiClient.get<PaginatedResponse<DistributionWithdrawalRequest>>('/user/distribution/withdrawals', {
    params: { status, page, page_size: pageSize }
  })
  return data
}

export const userAPI = {
  getProfile,
  updateProfile,
  changePassword,
  getDistributionProfile,
  getDistributionSummary,
  bindDistributionInviter,
  listDistributionReferrals,
  listDistributionTeam,
  listDistributionCommissions,
  createDistributionWithdrawal,
  listDistributionWithdrawals
}

export default userAPI
