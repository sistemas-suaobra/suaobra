package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrmRemindersSQL_AlertedFilterIsTextSafe(t *testing.T) {
	raw, err := templates.ReadFile("templates/crm/crm_reminders.sql")
	require.NoError(t, err)
	sql := strings.ToLower(string(raw))

	// Comparação via ->> (texto) evita falha com JSON boolean
	assert.Contains(t, sql, "properties ->> 'alerted'")
	assert.Contains(t, sql, "in ('false', '0', '')")
	assert.Contains(t, sql, "alert_at")
}

func TestCrmRemindersSQL_HasOwnerAndManagerFallbacks(t *testing.T) {
	raw, err := templates.ReadFile("templates/crm/crm_reminders.sql")
	require.NoError(t, err)
	sql := strings.ToLower(string(raw))

	assert.Contains(t, sql, "union all")
	assert.Contains(t, sql, "user.manager = true")
	assert.Contains(t, sql, "lead.owner_id")
}
