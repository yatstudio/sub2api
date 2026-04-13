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

ROOT = Path(__file__).resolve().parents[1]


def run(cmd: list[str], cwd: Path | None = None) -> tuple[int, str]:
    proc = subprocess.run(cmd, cwd=cwd or ROOT, text=True, capture_output=True)
    out = (proc.stdout or "") + (proc.stderr or "")
    return proc.returncode, out.strip()


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


def static_verify() -> list[str]:
    checks: list[str] = []

    setting_handler_test = ROOT / "backend/internal/handler/admin/setting_handler_distribution_risk_test.go"
    setting_service_test = ROOT / "backend/internal/service/setting_service_update_test.go"
    setting_handler = ROOT / "backend/internal/handler/admin/setting_handler.go"
    setting_service = ROOT / "backend/internal/service/setting_service.go"
    withdrawal_util = ROOT / "frontend/src/utils/distributionWithdrawalError.ts"
    zh_locale = ROOT / "frontend/src/i18n/locales/zh.ts"
    en_locale = ROOT / "frontend/src/i18n/locales/en.ts"

    # P0: handler + service coverage and negative clamp signals
    require_all(
        setting_handler_test,
        [
            ("TestSettingHandler_GetSettings_IncludesDistributionWithdrawalRiskControls", "P0 handler read persisted regression"),
            ("TestSettingHandler_GetSettings_DistributionWithdrawalRiskControls_ClampNegative", "P0 handler read clamp regression"),
            ("TestSettingHandler_UpdateSettings_DistributionWithdrawalRiskControls_Persisted", "P0 handler write persisted regression"),
            ("TestSettingHandler_UpdateSettings_DistributionWithdrawalRiskControls_ClampNegative", "P0 handler write clamp regression"),
            ("TestGetChangedSettingKeys_DistributionWithdrawalRiskControls_IncludeAllFourFields", "P0 handler audit changed-keys include all four controls"),
            ("TestGetChangedSettingKeys_DistributionWithdrawalRiskControls_UnchangedNotIncluded", "P0 handler audit changed-keys ignore unchanged controls"),
        ],
    )
    require_all(
        setting_service_test,
        [
            ("TestSettingService_UpdateSettings_DistributionWithdrawalRiskControls_Persisted", "P0 service write persisted regression"),
            ("TestSettingService_UpdateSettings_DistributionWithdrawalRiskControls_ClampNegative", "P0 service write clamp regression"),
            ("TestSettingService_GetAllSettings_DistributionWithdrawalRiskControls_ReadPersisted", "P0 service read persisted regression"),
            ("TestSettingService_GetAllSettings_DistributionWithdrawalRiskControls_ClampNegative", "P0 service read clamp regression"),
        ],
    )
    checks.append("backend tests cover handler/service read+write for all 4 risk controls (including negative clamp)")

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

    # P1: frontend reason mapping and i18n message keys
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
    require_distribution_withdrawal_error_locale_keys(zh_locale, "zh")
    require_distribution_withdrawal_error_locale_keys(en_locale, "en")
    checks.append("frontend mapping + zh/en distribution.withdrawalErrors keys exist for cooldown/count/amount limits")

    return checks


def main() -> int:
    go_available = bool(shutil.which("go"))
    gofmt_available = bool(shutil.which("gofmt"))

    report: dict[str, object] = {
        "frontend_test": None,
        "backend_test": None,
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
    frontend_results: list[dict[str, object]] = []
    for spec in frontend_specs:
        code, out = run(["npm", "--prefix", "frontend", "run", "test:run", "--", spec])
        frontend_results.append({"spec": spec, "exit_code": code, "output": out})
        if code != 0:
            report["frontend_test"] = frontend_results
            print(json.dumps(report, ensure_ascii=False, indent=2))
            return 1

    report["frontend_test"] = frontend_results

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
        print(json.dumps(report, ensure_ascii=False, indent=2))
        return code

    report["mode"] = "static-fallback-no-go"
    report["limitations"].append("go toolchain unavailable; backend go test skipped and replaced by structured static verification")
    checks = static_verify()
    report["static_checks"] = checks
    print(json.dumps(report, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
