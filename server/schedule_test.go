package server

import (
	"strings"
	"testing"
	"time"

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

func TestCrmRemindersSQL_RequiresPositiveAbsoluteAlertAt(t *testing.T) {
	raw, err := templates.ReadFile("templates/crm/crm_reminders.sql")
	require.NoError(t, err)
	sql := strings.ToLower(string(raw))

	// Data personalizada: alert_at absoluto em ms; ignora 0/vazio
	assert.Contains(t, sql, "cast((lead.properties -> 'alert_at') as real) > 0")
	assert.Contains(t, sql, "datetime((lead.properties -> 'alert_at') / 1000, 'unixepoch')")
}

func TestFormatReminderAtBR(t *testing.T) {
	// 2026-08-17 18:32:18 -03 = 2026-08-17 21:32:18 UTC
	sec := time.Date(2026, 8, 17, 21, 32, 18, 0, time.UTC).Unix()
	got := formatReminderAtBR(sec)
	assert.Equal(t, "17/08/2026 18:32", got)

	assert.Equal(t, "", formatReminderAtBR(0))
	assert.Equal(t, "", formatReminderAtBR(nil))
}
