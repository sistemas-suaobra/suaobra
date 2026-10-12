package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrmLeadsQueryLoadsLatestMoveOnce(t *testing.T) {
	raw, err := templates.ReadFile("templates/crm/crm_leads.sql")
	assert.NoError(t, err)

	sql := strings.ToLower(string(raw))
	assert.NotContains(t, sql, "select activity.id")
	assert.Contains(t, sql, "row_number() over")
	assert.Contains(t, sql, "partition by activity.lead_id")
	assert.Contains(t, sql, "{:stage_id} in ('', 'all')")
}
