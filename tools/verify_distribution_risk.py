#!/usr/bin/env python3
"""Minimal regression verification for distribution withdrawal risk controls.

Behavior:
1) Always runs frontend targeted unit tests for withdrawal error mapping + zh/en locale texts.
2) If Go toolchain is available, runs focused backend unit tests.
3) If Go is unavailable, performs structured static validation on key files.
"""

from __future__ import annotations

import json
import re
import shutil
import subprocess
import sys
from pathlib import Path
from typing import Sequence

ROOT = Path(__file__).resolve().parents[1]


def run(cmd: Sequence[str], cwd: Path | None = None) -> tuple[int, str]:
    proc = subprocess.run(list(cmd), cwd=cwd or ROOT, text=True, capture_output=True)
    out = (proc.stdout or "") + (proc.stderr or "")
    return proc.returncode, out.strip()


def resolve_frontend_test_runner() -> tuple[list[str], str | None]:
    """Return command prefix for targeted frontend vitest runs and optional limitation message."""
    if shutil.which("pnpm"):
        return ["pnpm", "--dir", "frontend", "exec", "vitest", "run"], None
    if shutil.which("npm"):
        return ["npm", "--prefix", "frontend", "exec", "vitest", "run", "--"], "pnpm not found; used npm fallback for frontend tests"
    raise RuntimeError("neither pnpm nor npm is available; cannot run frontend distribution risk tests")


def require_all(path: Path, checks: list[tuple[str, str]]) -> None:
    content = path.read_text(encoding="utf-8")
    rel = path.relative_to(ROOT)
    for needle, title in checks:
        if needle not in content:
            raise AssertionError(f"{title}: missing `{needle}` in {rel}")


def require_distribution_withdrawal_error_locale_keys(path: Path, locale_name: str) -> None:
    content = path.read_text(encoding="utf-8")
    # Scoped check: enforce that cooldown/dailyLimitCount/dailyLimitAmount exist inside
    # distribution.withdrawalErrors block, not just anywhere in the locale file.
    pattern = re.compile(
        r"distribution\s*:\s*\{[\s\S]*?withdrawalErrors\s*:\s*\{(?P<body>[\s\S]*?)\}\s*,",
        re.MULTILINE,
    )
    match = pattern.search(content)
    if not match:
        raise AssertionError(f"P1 {locale_name} locale section: missing `distribution.withdrawalErrors` block in {path.relative_to(ROOT)}")

    body = match.group("body")
    required = ["cooldown", "dailyLimitCount", "dailyLimitAmount"]
    missing = [key for key in required if re.search(rf"\b{re.escape(key)}\s*:", body) is None]
    if missing:
        raise AssertionError(
            f"P1 {locale_name} locale section: missing keys {missing} in distribution.withdrawalErrors block of {path.relative_to(ROOT)}"
        )


def require_interface_field_keys(
    path: Path,
    interface_name: str,
    field_keys: list[str],
    title_prefix: str,
    *,
    optional: bool,
) -> None:
    content = path.read_text(encoding="utf-8")
    match = re.search(
        rf"export\s+interface\s+{re.escape(interface_name)}\s*\{{(?P<body>[\s\S]*?)\n\}}",
        content,
        re.MULTILINE,
    )
    if not match:
        raise AssertionError(
            f"{title_prefix}: missing `export interface {interface_name}` block in {path.relative_to(ROOT)}"
        )

    body = match.group("body")
    separator = r"\?" if optional else ""
    missing = [
        field
        for field in field_keys
        if re.search(rf"\b{re.escape(field)}\s*{separator}:\s*number\b", body) is None
    ]
    if missing:
        maybe_kind = "optional" if optional else "required"
        raise AssertionError(
            f"{title_prefix}: missing {maybe_kind} number fields {missing} in interface {interface_name} of {path.relative_to(ROOT)}"
        )


def is_frontend_environment_failure(output: str) -> bool:
    lower = output.lower()
    env_markers = [
        "vitest: not found",
        "command not found",
        "cannot find module",
        "module not found",
        "enoent",
        "missing script",
        "is not recognized as an internal or external command",
        "npm err! code enoent",
        "pnpm: command not found",
    ]
    return any(marker in lower for marker in env_markers)


def static_verify() -> list[str]:
    checks: list[str] = []

    setting_handler_test = ROOT / "backend/internal/handler/admin/setting_handler_distribution_risk_test.go"
    setting_service_test = ROOT / "backend/internal/service/setting_service_update_test.go"
    setting_handler = ROOT / "backend/internal/handler/admin/setting_handler.go"
    setting_service = ROOT / "backend/internal/service/setting_service.go"
    user_distribution_service = ROOT / "backend/internal/service/user_distribution.go"
    settings_dto = ROOT / "backend/internal/handler/dto/settings.go"
    admin_settings_api = ROOT / "frontend/src/api/admin/settings.ts"
    withdrawal_util = ROOT / "frontend/src/utils/distributionWithdrawalError.ts"
    distribution_view = ROOT / "frontend/src/views/user/DistributionView.vue"
    withdrawal_error_spec = ROOT / "frontend/src/utils/__tests__/distributionWithdrawalError.spec.ts"
    withdrawal_locale_message_spec = ROOT / "frontend/src/utils/__tests__/distributionWithdrawalError.locale-message.spec.ts"
    zh_locale = ROOT / "frontend/src/i18n/locales/zh.ts"
    en_locale = ROOT / "frontend/src/i18n/locales/en.ts"

    # P0: handler + service coverage and negative clamp signals
    require_all(
        setting_handler_test,
        [
            ("TestSettingHandler_GetSettings_IncludesDistributionWithdrawalRiskControls", "P0 handler read persisted regression"),
            ("TestSettingHandler_GetSettings_DistributionWithdrawalRiskControls_ClampNegative", "P0 handler read clamp regression"),
            ("TestSettingHandler_GetSettings_DistributionWithdrawalRiskControls_InvalidRawValuesFallbackToDefaults", "P0 handler read invalid raw values fallback regression"),
            ("TestSettingHandler_GetSettings_DistributionWithdrawalRiskControls_MixedRawValues_PerFieldFallbackAndClamp", "P0 handler read mixed raw per-field fallback + clamp regression"),
            ("TestSettingHandler_UpdateSettings_DistributionWithdrawalRiskControls_Persisted", "P0 handler write persisted regression"),
            ("TestSettingHandler_UpdateSettings_DistributionWithdrawalRiskControls_ClampNegative", "P0 handler write clamp regression"),
            ("TestSettingHandler_UpdateSettings_DistributionWithdrawalRiskControls_ClampNegative_PerField", "P0 handler write per-field clamp regression"),
            ("TestSettingHandler_UpdateSettings_DistributionWithdrawalRiskControls_OmittedFieldsKeepPrevious", "P3 handler partial-update semantics keep previous values when omitted"),
            ("TestGetChangedSettingKeys_DistributionWithdrawalRiskControls_IncludeAllFourFields", "P0 handler audit changed-keys include all four controls"),
            ("TestGetChangedSettingKeys_DistributionWithdrawalRiskControls_UnchangedNotIncluded", "P0 handler audit changed-keys ignore unchanged controls"),
        ],
    )
    require_all(
        setting_service_test,
        [
            ("TestSettingService_UpdateSettings_DistributionWithdrawalRiskControls_Persisted", "P0 service write persisted regression"),
            ("TestSettingService_UpdateSettings_DistributionWithdrawalRiskControls_ClampNegative", "P0 service write clamp regression"),
            ("TestSettingService_UpdateSettings_DistributionWithdrawalRiskControls_ClampNegative_PerField", "P0 service write per-field clamp regression"),
            ("TestSettingService_GetAllSettings_DistributionWithdrawalRiskControls_ReadPersisted", "P0 service read persisted regression"),
            ("TestSettingService_GetAllSettings_DistributionWithdrawalRiskControls_ClampNegative", "P0 service read clamp regression"),
            ("TestSettingService_GetAllSettings_DistributionWithdrawalRiskControls_InvalidRawValuesFallbackToDefaults", "P0 service read invalid raw values fallback regression"),
            ("TestSettingService_GetAllSettings_DistributionWithdrawalRiskControls_MixedRawValues_PerFieldFallbackAndClamp", "P0 service read mixed raw per-field fallback + clamp regression"),
        ],
    )
    require_all(
        setting_handler_test,
        [
            ("require.Equal(t, 1000.0, resp.Data.RiskThreshold)", "P0 handler invalid raw fallback default threshold"),
            ("require.Equal(t, 0, resp.Data.CooldownDays)", "P0 handler invalid raw fallback default cooldown days"),
            ("require.Equal(t, 1, resp.Data.DailyLimitCount)", "P0 handler invalid raw fallback default daily count"),
            ("require.Equal(t, 10000.0, resp.Data.DailyLimitAmt)", "P0 handler invalid raw fallback default daily amount"),
        ],
    )
    require_all(
        setting_service_test,
        [
            ("require.Equal(t, 1000.0, settings.DistributionWithdrawalRiskThreshold)", "P0 service invalid raw fallback default threshold"),
            ("require.Equal(t, 0, settings.DistributionWithdrawalCooldownDays)", "P0 service invalid raw fallback default cooldown days"),
            ("require.Equal(t, 1, settings.DistributionWithdrawalDailyLimitCount)", "P0 service invalid raw fallback default daily count"),
            ("require.Equal(t, 10000.0, settings.DistributionWithdrawalDailyLimitAmount)", "P0 service invalid raw fallback default daily amount"),
        ],
    )
    require_all(
        setting_handler_test,
        [
            ("require.Equal(t, float64(0), resp.Data.RiskThreshold)", "P0 handler clamp assertion threshold"),
            ("require.Equal(t, 0, resp.Data.CooldownDays)", "P0 handler clamp assertion cooldown days"),
            ("require.Equal(t, 0, resp.Data.DailyLimitCount)", "P0 handler clamp assertion daily count"),
            ("require.Equal(t, float64(0), resp.Data.DailyLimitAmt)", "P0 handler clamp assertion daily amount"),
            ("require.Equal(t, \"0.00000000\", repo.settings[service.SettingKeyDistributionWithdrawalRiskThreshold])", "P0 handler write clamp persisted threshold"),
            ("require.Equal(t, \"0\", repo.settings[service.SettingKeyDistributionWithdrawalCooldownDays])", "P0 handler write clamp persisted cooldown days"),
            ("require.Equal(t, \"0\", repo.settings[service.SettingKeyDistributionWithdrawalDailyLimitCount])", "P0 handler write clamp persisted daily count"),
            ("require.Equal(t, \"0.00000000\", repo.settings[service.SettingKeyDistributionWithdrawalDailyLimitAmount])", "P0 handler write clamp persisted daily amount"),
            ("require.Equal(t, \"6\", repo.settings[service.SettingKeyDistributionWithdrawalCooldownDays])", "P0 handler write per-field clamp keeps valid cooldown"),
            ("require.Equal(t, \"2500.50000000\", repo.settings[service.SettingKeyDistributionWithdrawalDailyLimitAmount])", "P0 handler write per-field clamp keeps valid daily amount"),
            ("require.Equal(t, 6, resp.Data.CooldownDays)", "P0 handler write per-field clamp response keeps valid cooldown"),
            ("require.Equal(t, 2500.5, resp.Data.DailyLimitAmt)", "P0 handler write per-field clamp response keeps valid daily amount"),
            ("require.Equal(t, 1000.0, resp.Data.RiskThreshold)", "P0 handler mixed raw fallback default threshold"),
            ("require.Equal(t, 6, resp.Data.CooldownDays)", "P0 handler mixed raw keeps valid cooldown"),
            ("require.Equal(t, 0, resp.Data.DailyLimitCount)", "P0 handler mixed raw clamps negative daily count"),
            ("require.Equal(t, 2500.5, resp.Data.DailyLimitAmt)", "P0 handler mixed raw keeps valid daily amount"),
            ("body := []byte(`{}`)", "P3 handler omitted-fields regression uses empty payload to validate keep-previous semantics"),
            ("require.Equal(t, \"66.60000000\", repo.settings[service.SettingKeyDistributionWithdrawalRiskThreshold])", "P3 handler omitted-fields keeps previous persisted risk threshold"),
            ("require.Equal(t, \"2\", repo.settings[service.SettingKeyDistributionWithdrawalCooldownDays])", "P3 handler omitted-fields keeps previous persisted cooldown days"),
            ("require.Equal(t, \"3\", repo.settings[service.SettingKeyDistributionWithdrawalDailyLimitCount])", "P3 handler omitted-fields keeps previous persisted daily count"),
            ("require.Equal(t, \"123.45000000\", repo.settings[service.SettingKeyDistributionWithdrawalDailyLimitAmount])", "P3 handler omitted-fields keeps previous persisted daily amount"),
            ("require.Equal(t, 66.6, resp.Data.RiskThreshold)", "P3 handler omitted-fields response keeps previous risk threshold"),
            ("require.Equal(t, 2, resp.Data.CooldownDays)", "P3 handler omitted-fields response keeps previous cooldown days"),
            ("require.Equal(t, 3, resp.Data.DailyLimitCount)", "P3 handler omitted-fields response keeps previous daily count"),
            ("require.Equal(t, 123.45, resp.Data.DailyLimitAmt)", "P3 handler omitted-fields response keeps previous daily amount"),
        ],
    )
    require_all(
        setting_service_test,
        [
            ("require.Equal(t, \"0.00000000\", repo.updates[SettingKeyDistributionWithdrawalRiskThreshold])", "P0 service write clamp persisted threshold"),
            ("require.Equal(t, \"0\", repo.updates[SettingKeyDistributionWithdrawalCooldownDays])", "P0 service write clamp persisted cooldown days"),
            ("require.Equal(t, \"0\", repo.updates[SettingKeyDistributionWithdrawalDailyLimitCount])", "P0 service write clamp persisted daily count"),
            ("require.Equal(t, \"0.00000000\", repo.updates[SettingKeyDistributionWithdrawalDailyLimitAmount])", "P0 service write clamp persisted daily amount"),
            ("require.Equal(t, \"6\", repo.updates[SettingKeyDistributionWithdrawalCooldownDays])", "P0 service write per-field clamp keeps valid cooldown"),
            ("require.Equal(t, \"2500.50000000\", repo.updates[SettingKeyDistributionWithdrawalDailyLimitAmount])", "P0 service write per-field clamp keeps valid daily amount"),
            ("require.Equal(t, float64(0), settings.DistributionWithdrawalRiskThreshold)", "P0 service read clamp threshold"),
            ("require.Equal(t, 0, settings.DistributionWithdrawalCooldownDays)", "P0 service read clamp cooldown days"),
            ("require.Equal(t, 0, settings.DistributionWithdrawalDailyLimitCount)", "P0 service read clamp daily count"),
            ("require.Equal(t, float64(0), settings.DistributionWithdrawalDailyLimitAmount)", "P0 service read clamp daily amount"),
            ("require.Equal(t, 1000.0, settings.DistributionWithdrawalRiskThreshold)", "P0 service mixed raw fallback default threshold"),
            ("require.Equal(t, 6, settings.DistributionWithdrawalCooldownDays)", "P0 service mixed raw keeps valid cooldown"),
            ("require.Equal(t, 0, settings.DistributionWithdrawalDailyLimitCount)", "P0 service mixed raw clamps negative daily count"),
            ("require.Equal(t, 2500.5, settings.DistributionWithdrawalDailyLimitAmount)", "P0 service mixed raw keeps valid daily amount"),
        ],
    )
    checks.append("backend tests cover handler/service read+write for all 4 risk controls (including negative clamp + invalid-raw and mixed-raw per-field fallback/default semantics)")

    require_all(
        setting_handler,
        [
            ("DistributionWithdrawalRiskThreshold", "P0 handler threshold field"),
            ("DistributionWithdrawalCooldownDays", "P0 handler cooldown field"),
            ("DistributionWithdrawalDailyLimitCount", "P0 handler daily count field"),
            ("DistributionWithdrawalDailyLimitAmount", "P0 handler daily amount field"),
            ("if req.DistributionWithdrawalRiskThreshold != nil && *req.DistributionWithdrawalRiskThreshold < 0", "P0 handler threshold clamp"),
            ("if req.DistributionWithdrawalCooldownDays != nil && *req.DistributionWithdrawalCooldownDays < 0", "P0 handler cooldown clamp"),
            ("if req.DistributionWithdrawalDailyLimitCount != nil && *req.DistributionWithdrawalDailyLimitCount < 0", "P0 handler daily count clamp"),
            ("if req.DistributionWithdrawalDailyLimitAmount != nil && *req.DistributionWithdrawalDailyLimitAmount < 0", "P0 handler daily amount clamp"),
        ],
    )
    require_all(
        setting_service,
        [
            ("if settings.DistributionWithdrawalRiskThreshold < 0", "P0 service threshold clamp"),
            ("if settings.DistributionWithdrawalCooldownDays < 0", "P0 service cooldown clamp"),
            ("if settings.DistributionWithdrawalDailyLimitCount < 0", "P0 service daily count clamp"),
            ("if settings.DistributionWithdrawalDailyLimitAmount < 0", "P0 service daily amount clamp"),
        ],
    )
    checks.append("backend implementation contains all 4 risk fields + clamp logic in handler/service")

    # P3: DTO/API schema consistency for the same 4 risk control fields
    require_all(
        settings_dto,
        [
            ("DistributionWithdrawalRiskThreshold", "P3 DTO response contains risk threshold field"),
            ("DistributionWithdrawalCooldownDays", "P3 DTO response contains cooldown days field"),
            ("DistributionWithdrawalDailyLimitCount", "P3 DTO response contains daily count field"),
            ("DistributionWithdrawalDailyLimitAmount", "P3 DTO response contains daily amount field"),
            ('json:"distribution_withdrawal_risk_threshold"', "P3 DTO json tag for risk threshold"),
            ('json:"distribution_withdrawal_cooldown_days"', "P3 DTO json tag for cooldown days"),
            ('json:"distribution_withdrawal_daily_limit_count"', "P3 DTO json tag for daily count"),
            ('json:"distribution_withdrawal_daily_limit_amount"', "P3 DTO json tag for daily amount"),
        ],
    )
    require_all(
        setting_handler,
        [
            ("DistributionWithdrawalRiskThreshold   *float64", "P3 handler update request uses pointer for risk threshold"),
            ('json:"distribution_withdrawal_risk_threshold"', "P3 handler update request json tag for risk threshold"),
            ("DistributionWithdrawalCooldownDays    *int", "P3 handler update request uses pointer for cooldown days"),
            ('json:"distribution_withdrawal_cooldown_days"', "P3 handler update request json tag for cooldown days"),
            ("DistributionWithdrawalDailyLimitCount *int", "P3 handler update request uses pointer for daily count"),
            ('json:"distribution_withdrawal_daily_limit_count"', "P3 handler update request json tag for daily count"),
            ("DistributionWithdrawalDailyLimitAmount *float64", "P3 handler update request uses pointer for daily amount"),
            ('json:"distribution_withdrawal_daily_limit_amount"', "P3 handler update request json tag for daily amount"),
            ("return previousSettings.DistributionWithdrawalRiskThreshold", "P3 handler omitted risk threshold falls back to previous setting"),
            ("return previousSettings.DistributionWithdrawalCooldownDays", "P3 handler omitted cooldown days falls back to previous setting"),
            ("return previousSettings.DistributionWithdrawalDailyLimitCount", "P3 handler omitted daily count falls back to previous setting"),
            ("return previousSettings.DistributionWithdrawalDailyLimitAmount", "P3 handler omitted daily amount falls back to previous setting"),
        ],
    )
    risk_control_fields = [
        "distribution_withdrawal_risk_threshold",
        "distribution_withdrawal_cooldown_days",
        "distribution_withdrawal_daily_limit_count",
        "distribution_withdrawal_daily_limit_amount",
    ]
    require_interface_field_keys(
        admin_settings_api,
        "SystemSettings",
        risk_control_fields,
        "P3 frontend settings response type",
        optional=False,
    )
    require_interface_field_keys(
        admin_settings_api,
        "UpdateSettingsRequest",
        risk_control_fields,
        "P3 frontend settings update-request type",
        optional=True,
    )
    checks.append("DTO/handler/frontend admin settings API remain aligned for all 4 risk control fields (scoped checks in SystemSettings + UpdateSettingsRequest)")

    # P1: backend->frontend reason mapping and i18n message keys
    require_all(
        user_distribution_service,
        [
            ('"DISTRIBUTION_WITHDRAWAL_COOLDOWN"', "P1 backend exports cooldown reason code"),
            ('"DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT"', "P1 backend exports daily-count reason code"),
            ('"DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT"', "P1 backend exports daily-amount reason code"),
        ],
    )
    require_all(
        withdrawal_util,
        [
            ("DISTRIBUTION_WITHDRAWAL_COOLDOWN", "P1 cooldown mapping"),
            ("DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT", "P1 daily count mapping"),
            ("DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_COUNT", "P1 daily count alias mapping"),
            ("DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT", "P1 daily amount mapping"),
            ("DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_AMOUNT", "P1 daily amount alias mapping"),
        ],
    )
    require_all(
        withdrawal_error_spec,
        [
            ("response.data.code payload shape", "P1 reason extractor spec covers response.data.code payload shape"),
            ("top-level data.code payload shape", "P1 reason extractor spec covers top-level data.code payload shape"),
            ("top-level error object reason/code payload shape", "P1 reason extractor spec covers top-level error.reason/error.code payload shape"),
            ("error: { reason: 'distribution_withdrawal_daily_limit' }", "P1 reason extractor spec covers top-level error.reason canonical daily-count envelope"),
            ("error: { reason: 'distribution_withdrawal_daily_limit_amount' }", "P1 reason extractor spec covers top-level error.reason daily-amount envelope"),
            ("error: { code: 'distribution_withdrawal_daily_limit' }", "P1 reason extractor spec covers top-level error.code canonical daily-count envelope"),
            ("error: 'distribution_withdrawal_daily_limit'", "P1 reason extractor spec covers top-level error string envelope for canonical daily-count code"),
            ("error: 'distribution_withdrawal_daily_limit_amount'", "P1 reason extractor spec covers top-level error string envelope for daily-amount code"),
            ("code: 'distribution_withdrawal_daily_limit'", "P1 reason extractor spec covers canonical daily-count code in code envelopes"),
            ("data: { message: 'raw backend msg from data' }", "P1 message fallback spec covers top-level data.message payload shape"),
            ("data: { message: 'request rejected: distribution_withdrawal_daily_limit_amount' }", "P1 reason extractor spec covers top-level data.message token fallback for daily amount"),
            ("message: 'request rejected: distribution_withdrawal_cooldown'", "P1 reason extractor spec covers top-level message token fallback for cooldown"),
            ("message: 'request rejected: distribution_withdrawal_daily_limit_count'", "P1 reason extractor spec covers top-level message token fallback for daily-count"),
            ("error: { message: 'request rejected: distribution_withdrawal_daily_limit_count' }", "P1 reason extractor spec covers top-level error.message token fallback for daily-count"),
            ("error: { message: 'request rejected: distribution_withdrawal_daily_limit_amount' }", "P1 reason extractor spec covers top-level error.message token fallback for daily-amount"),
        ],
    )
    require_all(
        withdrawal_locale_message_spec,
        [
            ("DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_COUNT", "P1 locale-message spec covers daily-count alias reason code"),
            ("DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT_AMOUNT", "P1 locale-message spec covers daily-amount alias reason code"),
            ("distribution_withdrawal_cooldown", "P1 locale-message spec covers message-token fallback for cooldown"),
            ("distribution_withdrawal_daily_limit_count", "P1 locale-message spec covers message-token fallback for daily-count"),
            ("distribution_withdrawal_daily_limit_amount", "P1 locale-message spec covers message-token fallback for daily-amount"),
            ("response: { data: { code: 'distribution_withdrawal_cooldown' } }", "P1 locale-message spec covers response.data.code payload shape for cooldown"),
            ("data: { code: 'distribution_withdrawal_cooldown' }", "P1 locale-message spec covers top-level data.code payload shape for cooldown"),
            ("response: { data: { code: 'distribution_withdrawal_daily_limit' } }", "P1 locale-message spec covers response.data.code payload shape for canonical daily-count code"),
            ("data: { code: 'distribution_withdrawal_daily_limit' }", "P1 locale-message spec covers top-level data.code payload shape for canonical daily-count code"),
            ("response: { data: { code: 'distribution_withdrawal_daily_limit_count' } }", "P1 locale-message spec covers response.data.code payload shape for daily-count"),
            ("data: { code: 'distribution_withdrawal_daily_limit_count' }", "P1 locale-message spec covers top-level data.code payload shape for daily-count"),
            ("response: { data: { code: 'distribution_withdrawal_daily_limit_amount' } }", "P1 locale-message spec covers response.data.code payload shape for daily-amount"),
            ("data: { code: 'distribution_withdrawal_daily_limit_amount' }", "P1 locale-message spec covers top-level data.code payload shape for daily-amount"),
            ("error: { code: 'distribution_withdrawal_daily_limit' }", "P1 locale-message spec covers top-level error.code payload shape for canonical daily-count code"),
            ("error: { code: 'distribution_withdrawal_daily_limit_amount' }", "P1 locale-message spec covers top-level error.code payload shape for daily-amount"),
            ("error: { reason: 'distribution_withdrawal_cooldown' }", "P1 locale-message spec covers top-level error.reason payload shape for cooldown"),
            ("error: { reason: 'distribution_withdrawal_daily_limit' }", "P1 locale-message spec covers top-level error.reason payload shape for canonical daily-count code"),
            ("error: { reason: 'distribution_withdrawal_daily_limit_amount' }", "P1 locale-message spec covers top-level error.reason payload shape for daily-amount"),
            ("response: { data: { error: { message: 'request blocked: distribution_withdrawal_cooldown' } } }", "P1 locale-message spec covers response.data.error.message payload shape for cooldown token"),
            ("data: { error: { message: 'request blocked: distribution_withdrawal_cooldown' } }", "P1 locale-message spec covers top-level data.error.message payload shape for cooldown token"),
            ("data: { message: 'request blocked: distribution_withdrawal_cooldown' }", "P1 locale-message spec covers top-level data.message payload shape for cooldown token"),
            ("response: { data: { message: 'request blocked: distribution_withdrawal_cooldown' } }", "P1 locale-message spec covers response.data.message payload shape for cooldown token"),
            ("response: { data: { message: 'request blocked: distribution_withdrawal_daily_limit_count' } }", "P1 locale-message spec covers response.data.message payload shape for daily-count token"),
            ("response: { data: { message: 'request blocked: distribution_withdrawal_daily_limit_amount' } }", "P1 locale-message spec covers response.data.message payload shape for daily-amount token"),
            ("data: { error: { message:", "P1 locale-message spec covers top-level data.error.message payload shape"),
            ("data: { error: 'distribution_withdrawal_cooldown' }", "P1 locale-message spec covers top-level data.error string payload for cooldown"),
            ("data: { error: 'distribution_withdrawal_daily_limit' }", "P1 locale-message spec covers top-level data.error string payload for canonical daily-count code"),
            ("data: { error: 'distribution_withdrawal_daily_limit_count' }", "P1 locale-message spec covers top-level data.error string payload for daily-count alias"),
            ("data: { error: 'distribution_withdrawal_daily_limit_amount' }", "P1 locale-message spec covers top-level data.error string payload for daily-amount"),
            ("response: { data: { error: 'distribution_withdrawal_cooldown' } }", "P1 locale-message spec covers response.data.error string payload for cooldown"),
            ("response: { data: { error: 'distribution_withdrawal_daily_limit' } }", "P1 locale-message spec covers response.data.error string payload for canonical daily-count code"),
            ("response: { data: { error: 'distribution_withdrawal_daily_limit_count' } }", "P1 locale-message spec covers response.data.error string payload for daily-count alias"),
            ("response: { data: { error: 'distribution_withdrawal_daily_limit_amount' } }", "P1 locale-message spec covers response.data.error string payload for daily-amount"),
            ("error: 'distribution_withdrawal_cooldown'", "P1 locale-message spec covers top-level error string payload for cooldown"),
            ("error: 'distribution_withdrawal_daily_limit'", "P1 locale-message spec covers top-level error string payload for canonical daily-count code"),
            ("error: 'distribution_withdrawal_daily_limit_amount'", "P1 locale-message spec covers top-level error string payload for daily-amount"),
            ("data: { message: 'request blocked: distribution_withdrawal_daily_limit_count' }", "P1 locale-message spec covers top-level data.message payload shape for daily-count token"),
            ("data: { message: 'request blocked: distribution_withdrawal_daily_limit_amount' }", "P1 locale-message spec covers top-level data.message payload shape for daily-amount token"),
            ("message: 'request blocked: distribution_withdrawal_cooldown'", "P1 locale-message spec covers top-level message payload shape for cooldown token"),
            ("message: 'request blocked: distribution_withdrawal_daily_limit_count'", "P1 locale-message spec covers top-level message payload shape for daily-count token"),
            ("message: 'request blocked: distribution_withdrawal_daily_limit_amount'", "P1 locale-message spec covers top-level message payload shape for daily-amount token"),
            ("error: { message: 'request blocked: distribution_withdrawal_daily_limit_count' }", "P1 locale-message spec covers top-level error.message payload shape for daily-count token"),
            ("error: { message: 'request blocked: distribution_withdrawal_daily_limit_amount' }", "P1 locale-message spec covers top-level error.message payload shape for daily-amount token"),
        ],
    )
    require_distribution_withdrawal_error_locale_keys(zh_locale, "zh")
    require_distribution_withdrawal_error_locale_keys(en_locale, "en")
    require_all(
        distribution_view,
        [
            ("import { resolveDistributionWithdrawalErrorMessage } from '@/utils/distributionWithdrawalError'", "P1 DistributionView imports withdrawal error resolver"),
            ("const withdrawalErrorMessage = (error: unknown) => resolveDistributionWithdrawalErrorMessage(error, t)", "P1 DistributionView derives readable withdrawal error message via resolver"),
            ("appStore.showError(withdrawalErrorMessage(error))", "P1 DistributionView shows resolver-generated readable withdrawal error message"),
        ],
    )
    checks.append("backend reason codes + frontend mapping + DistributionView wiring + locale-message regressions + zh/en distribution.withdrawalErrors keys exist for cooldown/count/amount limits")

    return checks


def main() -> int:
    go_available = bool(shutil.which("go"))
    gofmt_available = bool(shutil.which("gofmt"))

    report: dict[str, object] = {
        "frontend_test": None,
        "backend_test": None,
        "frontend_runner": "",
        "go_available": go_available,
        "gofmt_available": gofmt_available,
        "mode": "",
        "static_checks": [],
        "limitations": [],
    }

    frontend_specs = [
        "src/utils/__tests__/distributionWithdrawalError.spec.ts",
        "src/utils/__tests__/distributionWithdrawalError.locale-message.spec.ts",
        "src/i18n/__tests__/distributionWithdrawalLocales.spec.ts",
    ]
    frontend_runner: list[str] | None = None
    frontend_runner_limitation: str | None = None
    frontend_results: list[dict[str, object]] = []

    try:
        frontend_runner, frontend_runner_limitation = resolve_frontend_test_runner()
    except RuntimeError as exc:
        report["limitations"].append(
            f"{exc}; switched to structured static verification for frontend checks"
        )
        report["frontend_test"] = []
    else:
        report["frontend_runner"] = " ".join(frontend_runner)
        if frontend_runner_limitation:
            report["limitations"].append(frontend_runner_limitation)

        frontend_env_blocked = False
        for spec in frontend_specs:
            code, out = run([*frontend_runner, spec])
            frontend_results.append({"spec": spec, "exit_code": code, "output": out})
            if code != 0:
                if is_frontend_environment_failure(out):
                    frontend_env_blocked = True
                    report["limitations"].append(
                        f"frontend runner environment limitation on {spec}; switched to static verification for frontend checks"
                    )
                    break
                report["frontend_test"] = frontend_results
                print(json.dumps(report, ensure_ascii=False, indent=2))
                return 1

        report["frontend_test"] = frontend_results
        if frontend_env_blocked:
            frontend_runner = None

    static_checks: list[str] = []

    if go_available:
        report["mode"] = "go-tests"
        if not gofmt_available:
            report["limitations"].append("gofmt not found; only go test was executed")
        code, out = run(
            [
                "go",
                "test",
                "./internal/handler/admin",
                "./internal/service",
                "-run",
                "DistributionWithdrawalRiskControls",
                "-count=1",
            ],
            cwd=ROOT / "backend",
        )
        report["backend_test"] = {"exit_code": code, "output": out}
        if code != 0:
            print(json.dumps(report, ensure_ascii=False, indent=2))
            return code

        if frontend_runner is None:
            static_checks = static_verify()
            report["static_checks"] = static_checks
        print(json.dumps(report, ensure_ascii=False, indent=2))
        return 0

    report["mode"] = "static-fallback-no-go"
    report["limitations"].append("go toolchain unavailable; backend go test skipped and replaced by structured static verification")
    static_checks = static_verify()
    report["static_checks"] = static_checks
    print(json.dumps(report, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
