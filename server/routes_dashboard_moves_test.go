package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDashboardMovesUsesSavedLeadName(t *testing.T) {
	raw, err := templates.ReadFile("templates/dashboard/dashboard_moves.sql")
	assert.NoError(t, err)

	sql := strings.ToLower(string(raw))
	assert.Contains(t, sql, "lead.properties -> 'obra' ->> 'owner'")
	assert.Contains(t, sql, "cop.professional")
	assert.Contains(t, sql, "$.contacts[0].name")
}
