package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDashboardFunnelCountsVendaMaisStages(t *testing.T) {
	raw, err := templates.ReadFile("templates/dashboard/dashboard_funnel.sql")
	assert.NoError(t, err)

	sql := strings.ToLower(string(raw))
	assert.Contains(t, sql, "lead_activity")
	assert.Contains(t, sql, "%propost%")
	assert.Contains(t, sql, "%vend%")
	assert.Contains(t, sql, "opportunity_cnt")
	assert.Contains(t, sql, "sold_cnt")
}
