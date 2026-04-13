//go:build unit

package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type settingHandlerRepoStub struct {
	settings map[string]string
}

func (s *settingHandlerRepoStub) Get(ctx context.Context, key string) (*service.Setting, error) {
	panic("unexpected Get call")
}

func (s *settingHandlerRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if s.settings == nil {
		return "", nil
	}
	return s.settings[key], nil
}

func (s *settingHandlerRepoStub) Set(ctx context.Context, key, value string) error {
	if s.settings == nil {
		s.settings = map[string]string{}
	}
	s.settings[key] = value
	return nil
}

func (s *settingHandlerRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		if v, ok := s.settings[k]; ok {
			out[k] = v
		}
	}
	return out, nil
}

func (s *settingHandlerRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	if s.settings == nil {
		s.settings = make(map[string]string, len(settings))
	}
	for k, v := range settings {
		s.settings[k] = v
	}
	return nil
}

func (s *settingHandlerRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	out := make(map[string]string, len(s.settings))
	for k, v := range s.settings {
		out[k] = v
	}
	return out, nil
}

func (s *settingHandlerRepoStub) Delete(ctx context.Context, key string) error {
	delete(s.settings, key)
	return nil
}

func newSettingHandlerForTest(repo *settingHandlerRepoStub) *SettingHandler {
	svc := service.NewSettingService(repo, &config.Config{})
	return NewSettingHandler(svc, nil, nil, nil, nil)
}

func TestSettingHandler_GetSettings_IncludesDistributionWithdrawalRiskControls(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{settings: map[string]string{
		service.SettingKeyDistributionWithdrawalRiskThreshold:   "1234.50000000",
		service.SettingKeyDistributionWithdrawalCooldownDays:    "3",
		service.SettingKeyDistributionWithdrawalDailyLimitCount: "4",
		service.SettingKeyDistributionWithdrawalDailyLimitAmount: "999.25000000",
	}}
	h := newSettingHandlerForTest(repo)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)

	h.GetSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp struct {
		Data struct {
			RiskThreshold   float64 `json:"distribution_withdrawal_risk_threshold"`
			CooldownDays    int     `json:"distribution_withdrawal_cooldown_days"`
			DailyLimitCount int     `json:"distribution_withdrawal_daily_limit_count"`
			DailyLimitAmt   float64 `json:"distribution_withdrawal_daily_limit_amount"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 1234.5, resp.Data.RiskThreshold)
	require.Equal(t, 3, resp.Data.CooldownDays)
	require.Equal(t, 4, resp.Data.DailyLimitCount)
	require.Equal(t, 999.25, resp.Data.DailyLimitAmt)
}

func TestSettingHandler_GetSettings_DistributionWithdrawalRiskControls_ClampNegative(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{settings: map[string]string{
		service.SettingKeyDistributionWithdrawalRiskThreshold:   "-1",
		service.SettingKeyDistributionWithdrawalCooldownDays:    "-2",
		service.SettingKeyDistributionWithdrawalDailyLimitCount: "-3",
		service.SettingKeyDistributionWithdrawalDailyLimitAmount: "-4",
	}}
	h := newSettingHandlerForTest(repo)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)

	h.GetSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp struct {
		Data struct {
			RiskThreshold   float64 `json:"distribution_withdrawal_risk_threshold"`
			CooldownDays    int     `json:"distribution_withdrawal_cooldown_days"`
			DailyLimitCount int     `json:"distribution_withdrawal_daily_limit_count"`
			DailyLimitAmt   float64 `json:"distribution_withdrawal_daily_limit_amount"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, float64(0), resp.Data.RiskThreshold)
	require.Equal(t, 0, resp.Data.CooldownDays)
	require.Equal(t, 0, resp.Data.DailyLimitCount)
	require.Equal(t, float64(0), resp.Data.DailyLimitAmt)
}

func TestSettingHandler_GetSettings_DistributionWithdrawalRiskControls_InvalidRawValuesFallbackToDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{settings: map[string]string{
		service.SettingKeyDistributionWithdrawalRiskThreshold:   "invalid-number",
		service.SettingKeyDistributionWithdrawalCooldownDays:    "invalid-days",
		service.SettingKeyDistributionWithdrawalDailyLimitCount: "invalid-count",
		service.SettingKeyDistributionWithdrawalDailyLimitAmount: "invalid-amount",
	}}
	h := newSettingHandlerForTest(repo)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)

	h.GetSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp struct {
		Data struct {
			RiskThreshold   float64 `json:"distribution_withdrawal_risk_threshold"`
			CooldownDays    int     `json:"distribution_withdrawal_cooldown_days"`
			DailyLimitCount int     `json:"distribution_withdrawal_daily_limit_count"`
			DailyLimitAmt   float64 `json:"distribution_withdrawal_daily_limit_amount"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 1000.0, resp.Data.RiskThreshold)
	require.Equal(t, 0, resp.Data.CooldownDays)
	require.Equal(t, 1, resp.Data.DailyLimitCount)
	require.Equal(t, 10000.0, resp.Data.DailyLimitAmt)
}

func TestSettingHandler_UpdateSettings_DistributionWithdrawalRiskControls_Persisted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{settings: map[string]string{}}
	h := newSettingHandlerForTest(repo)

	payload := map[string]any{
		"distribution_withdrawal_risk_threshold":    888.125,
		"distribution_withdrawal_cooldown_days":     7,
		"distribution_withdrawal_daily_limit_count": 5,
		"distribution_withdrawal_daily_limit_amount": 4500.75,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "888.12500000", repo.settings[service.SettingKeyDistributionWithdrawalRiskThreshold])
	require.Equal(t, "7", repo.settings[service.SettingKeyDistributionWithdrawalCooldownDays])
	require.Equal(t, "5", repo.settings[service.SettingKeyDistributionWithdrawalDailyLimitCount])
	require.Equal(t, "4500.75000000", repo.settings[service.SettingKeyDistributionWithdrawalDailyLimitAmount])

	var resp struct {
		Data struct {
			RiskThreshold   float64 `json:"distribution_withdrawal_risk_threshold"`
			CooldownDays    int     `json:"distribution_withdrawal_cooldown_days"`
			DailyLimitCount int     `json:"distribution_withdrawal_daily_limit_count"`
			DailyLimitAmt   float64 `json:"distribution_withdrawal_daily_limit_amount"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 888.125, resp.Data.RiskThreshold)
	require.Equal(t, 7, resp.Data.CooldownDays)
	require.Equal(t, 5, resp.Data.DailyLimitCount)
	require.Equal(t, 4500.75, resp.Data.DailyLimitAmt)
}

func TestSettingHandler_UpdateSettings_DistributionWithdrawalRiskControls_ClampNegative(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{settings: map[string]string{
		service.SettingKeyDistributionWithdrawalRiskThreshold:   "10",
		service.SettingKeyDistributionWithdrawalCooldownDays:    "1",
		service.SettingKeyDistributionWithdrawalDailyLimitCount: "1",
		service.SettingKeyDistributionWithdrawalDailyLimitAmount: "100",
	}}
	h := newSettingHandlerForTest(repo)

	payload := map[string]any{
		"distribution_withdrawal_risk_threshold":    -1,
		"distribution_withdrawal_cooldown_days":     -2,
		"distribution_withdrawal_daily_limit_count": -3,
		"distribution_withdrawal_daily_limit_amount": -4,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "0.00000000", repo.settings[service.SettingKeyDistributionWithdrawalRiskThreshold])
	require.Equal(t, "0", repo.settings[service.SettingKeyDistributionWithdrawalCooldownDays])
	require.Equal(t, "0", repo.settings[service.SettingKeyDistributionWithdrawalDailyLimitCount])
	require.Equal(t, "0.00000000", repo.settings[service.SettingKeyDistributionWithdrawalDailyLimitAmount])

	var resp struct {
		Data struct {
			RiskThreshold   float64 `json:"distribution_withdrawal_risk_threshold"`
			CooldownDays    int     `json:"distribution_withdrawal_cooldown_days"`
			DailyLimitCount int     `json:"distribution_withdrawal_daily_limit_count"`
			DailyLimitAmt   float64 `json:"distribution_withdrawal_daily_limit_amount"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, float64(0), resp.Data.RiskThreshold)
	require.Equal(t, 0, resp.Data.CooldownDays)
	require.Equal(t, 0, resp.Data.DailyLimitCount)
	require.Equal(t, float64(0), resp.Data.DailyLimitAmt)
}

func TestSettingHandler_UpdateSettings_DistributionWithdrawalRiskControls_OmittedFieldsKeepPrevious(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{settings: map[string]string{
		service.SettingKeyDistributionWithdrawalRiskThreshold:   "66.60000000",
		service.SettingKeyDistributionWithdrawalCooldownDays:    "2",
		service.SettingKeyDistributionWithdrawalDailyLimitCount: "3",
		service.SettingKeyDistributionWithdrawalDailyLimitAmount: "123.45000000",
	}}
	h := newSettingHandlerForTest(repo)

	body := []byte(`{}`)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateSettings(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "66.60000000", repo.settings[service.SettingKeyDistributionWithdrawalRiskThreshold])
	require.Equal(t, "2", repo.settings[service.SettingKeyDistributionWithdrawalCooldownDays])
	require.Equal(t, "3", repo.settings[service.SettingKeyDistributionWithdrawalDailyLimitCount])
	require.Equal(t, "123.45000000", repo.settings[service.SettingKeyDistributionWithdrawalDailyLimitAmount])

	var resp struct {
		Data struct {
			RiskThreshold   float64 `json:"distribution_withdrawal_risk_threshold"`
			CooldownDays    int     `json:"distribution_withdrawal_cooldown_days"`
			DailyLimitCount int     `json:"distribution_withdrawal_daily_limit_count"`
			DailyLimitAmt   float64 `json:"distribution_withdrawal_daily_limit_amount"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, 66.6, resp.Data.RiskThreshold)
	require.Equal(t, 2, resp.Data.CooldownDays)
	require.Equal(t, 3, resp.Data.DailyLimitCount)
	require.Equal(t, 123.45, resp.Data.DailyLimitAmt)
}

func TestGetChangedSettingKeys_DistributionWithdrawalRiskControls_IncludeAllFourFields(t *testing.T) {
	before := &dto.SystemSettingsResponse{
		DistributionWithdrawalRiskThreshold:   10,
		DistributionWithdrawalCooldownDays:    1,
		DistributionWithdrawalDailyLimitCount: 2,
		DistributionWithdrawalDailyLimitAmount: 300,
	}
	after := &dto.SystemSettingsResponse{
		DistributionWithdrawalRiskThreshold:   11,
		DistributionWithdrawalCooldownDays:    3,
		DistributionWithdrawalDailyLimitCount: 4,
		DistributionWithdrawalDailyLimitAmount: 500,
	}

	changed := getChangedSettingKeys(before, after)
	require.Contains(t, changed, "distribution_withdrawal_risk_threshold")
	require.Contains(t, changed, "distribution_withdrawal_cooldown_days")
	require.Contains(t, changed, "distribution_withdrawal_daily_limit_count")
	require.Contains(t, changed, "distribution_withdrawal_daily_limit_amount")
}

func TestGetChangedSettingKeys_DistributionWithdrawalRiskControls_UnchangedNotIncluded(t *testing.T) {
	before := &dto.SystemSettingsResponse{
		DistributionWithdrawalRiskThreshold:   88.8,
		DistributionWithdrawalCooldownDays:    2,
		DistributionWithdrawalDailyLimitCount: 5,
		DistributionWithdrawalDailyLimitAmount: 456.78,
	}
	after := &dto.SystemSettingsResponse{
		DistributionWithdrawalRiskThreshold:   88.8,
		DistributionWithdrawalCooldownDays:    2,
		DistributionWithdrawalDailyLimitCount: 5,
		DistributionWithdrawalDailyLimitAmount: 456.78,
	}

	changed := getChangedSettingKeys(before, after)
	require.NotContains(t, changed, "distribution_withdrawal_risk_threshold")
	require.NotContains(t, changed, "distribution_withdrawal_cooldown_days")
	require.NotContains(t, changed, "distribution_withdrawal_daily_limit_count")
	require.NotContains(t, changed, "distribution_withdrawal_daily_limit_amount")
}
