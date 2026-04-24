package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type stubGroupStatsDashboard struct {
	stats []usagestats.GroupStat
	err   error
}

func (s *stubGroupStatsDashboard) GetGroupStatsWithFilters(ctx context.Context, startTime, endTime time.Time, userID, apiKeyID, accountID, groupID int64, requestType *int16, stream *bool, billingType *int8) ([]usagestats.GroupStat, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.stats, nil
}

func (s *stubGroupStatsDashboard) GetGroupUsageSummary(ctx context.Context, todayStart time.Time) ([]usagestats.GroupUsageSummary, error) {
	return []usagestats.GroupUsageSummary{}, s.err
}

func TestGroupHandlerStats_IncludesUsageFromDashboardService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	adminSvc := newStubAdminService()
	dashboardSvc := &stubGroupStatsDashboard{
		stats: []usagestats.GroupStat{{
			GroupID:    2,
			Requests:   42,
			Cost:       12.34,
			ActualCost: 11.11,
		}},
	}
	groupHandler := NewGroupHandler(adminSvc, dashboardSvc, nil)
	router.GET("/api/v1/admin/groups/:id/stats", groupHandler.GetStats)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/groups/2/stats", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	data, ok := payload["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(1), data["total_api_keys"])
	require.Equal(t, float64(1), data["active_api_keys"])
	require.Equal(t, float64(42), data["total_requests"])
	require.Equal(t, 12.34, data["total_cost"])
}
