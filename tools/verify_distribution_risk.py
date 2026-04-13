#!/usr/bin/env python3
"""Minimal regression verification for distribution withdrawal risk controls.

Behavior:
1) Always runs frontend targeted unit tests for withdrawal error mapping.
2) If Go toolchain is available, runs focused backend unit tests.
3) If Go is unavailable, performs structured static validation on key files.
"""

from __future__ import annotations

import json
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def run(cmd: list[str], cwd: Path | None = None) -> tuple[int, str]:
    proc = subprocess.run(cmd, cwd=cwd or ROOT, text=True, capture_output=True)
    out = (proc.stdout or "") + (proc.stderr or "")
    return proc.returncode, out.strip()


def require_contains(path: Path, needle: str, title: str) -> None:
    content = path.read_text(encoding="utf-8")
    if needle not in content:
        raise AssertionError(f"{title}: missing `{needle}` in {path.relative_to(ROOT)}")


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
    require_contains(setting_handler_test, "TestSettingHandler_UpdateSettings_DistributionWithdrawalRiskControls_ClampNegative", "P0 handler regression")
    require_contains(setting_handler_test, "TestSettingHandler_GetSettings_DistributionWithdrawalRiskControls_ClampNegative", "P0 handler read clamp regression")
    require_contains(setting_service_test, "TestSettingService_UpdateSettings_DistributionWithdrawalRiskControls_ClampNegative", "P0 service update clamp regression")
    require_contains(setting_service_test, "TestSettingService_GetAllSettings_DistributionWithdrawalRiskControls_ClampNegative", "P0 service read clamp regression")
    checks.append("backend tests contain handler/service risk clamp regressions")

    require_contains(setting_handler, "DistributionWithdrawalRiskThreshold", "P0 handler fields")
    require_contains(setting_handler, "if req.DistributionWithdrawalRiskThreshold != nil && *req.DistributionWithdrawalRiskThreshold < 0", "P0 handler clamp")
    require_contains(setting_service, "if settings.DistributionWithdrawalDailyLimitAmount < 0", "P0 service clamp")
    checks.append("backend implementation contains risk settings + negative clamp logic")

    # P1: frontend reason mapping and i18n message keys
    require_contains(withdrawal_util, "DISTRIBUTION_WITHDRAWAL_COOLDOWN", "P1 cooldown mapping")
    require_contains(withdrawal_util, "DISTRIBUTION_WITHDRAWAL_DAILY_LIMIT", "P1 daily count mapping")
    require_contains(withdrawal_util, "DISTRIBUTION_WITHDRAWAL_DAILY_AMOUNT_LIMIT", "P1 daily amount mapping")
    require_contains(zh_locale, "withdrawalErrors", "P1 zh locale section")
    require_contains(zh_locale, "dailyLimitCount", "P1 zh daily count text")
    require_contains(zh_locale, "dailyLimitAmount", "P1 zh daily amount text")
    require_contains(en_locale, "withdrawalErrors", "P1 en locale section")
    require_contains(en_locale, "dailyLimitCount", "P1 en daily count text")
    require_contains(en_locale, "dailyLimitAmount", "P1 en daily amount text")
    checks.append("frontend mapping + zh/en readable messages exist for cooldown/count/amount limits")

    return checks


def main() -> int:
    report: dict[str, object] = {
        "frontend_test": None,
        "backend_test": None,
        "go_available": bool(shutil.which("go")),
        "mode": "",
        "static_checks": [],
    }

    code, out = run(["npm", "--prefix", "frontend", "run", "test:run", "--", "src/utils/__tests__/distributionWithdrawalError.spec.ts"])
    report["frontend_test"] = {"exit_code": code, "output": out}
    if code != 0:
        print(json.dumps(report, ensure_ascii=False, indent=2))
        return 1

    if shutil.which("go"):
        report["mode"] = "go-tests"
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
    checks = static_verify()
    report["static_checks"] = checks
    print(json.dumps(report, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
