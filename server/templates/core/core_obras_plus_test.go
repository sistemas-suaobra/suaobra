package coretemplate_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sqlTemplatePath(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)

	// server/templates/core/core_obras_plus_test.go -> core_obras_plus.sql
	return filepath.Join(filepath.Dir(file), "core_obras_plus.sql")
}

func TestCoreObrasPlusSQL_ContactFilter_Smoke(t *testing.T) {
	raw, err := os.ReadFile(sqlTemplatePath(t))
	require.NoError(t, err)

	sql := string(raw)

	assert.Contains(t, sql, "owner_enviado_em")
	assert.Contains(t, sql, "professional_enviado_em")
	assert.Contains(t, sql, "campanha_destinatarios")
}

func TestCoreObrasPlusSQL_ContactFilter_BlocksActiveStatuses(t *testing.T) {
	raw, err := os.ReadFile(sqlTemplatePath(t))
	require.NoError(t, err)

	sql := string(raw)

	// Não deve usar apenas ENVIADO (bug antigo)
	assert.NotContains(t, sql, "contato_tipo = 'OWNER' and status = 'ENVIADO'")
	assert.NotContains(t, sql, "contato_tipo = 'PROFISSIONAL' and status = 'ENVIADO'")

	// Deve excluir apenas falhas e ignorados
	assert.Contains(t, sql, "status not in ('FALHOU', 'IGNORADO')")
}

func TestCoreObrasPlusSQL_ContactFilter_PerRecipientType(t *testing.T) {
	raw, err := os.ReadFile(sqlTemplatePath(t))
	require.NoError(t, err)

	sql := strings.ToUpper(string(raw))

	ownerIdx := strings.Index(sql, "OWNER")
	profIdx := strings.Index(sql, "PROFISSIONAL")
	require.Greater(t, ownerIdx, -1)
	require.Greater(t, profIdx, -1)
	assert.NotEqual(t, ownerIdx, profIdx)
}
