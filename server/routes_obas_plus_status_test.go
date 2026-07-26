package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeStatusCond_UsesJoinedLeadNotSubquery(t *testing.T) {
	cond := makeStatusCond([]string{"todos"})

	assert.Contains(t, cond, "coalesce(l.excluded_at, '') = ''")
	assert.NotContains(t, strings.ToLower(cond), "select obra_id from main.lead")
	assert.NotContains(t, strings.ToLower(cond), " not in (")
}

func TestMakeStatusCond_VisitadaFavorita(t *testing.T) {
	visitada := makeStatusCond([]string{"ja-visitada"})
	assert.Contains(t, visitada, "l.visited_at > ''")
	assert.NotContains(t, strings.ToLower(visitada), "select obra_id")

	favorita := makeStatusCond([]string{"favorita"})
	assert.Contains(t, favorita, "l.favorited_at > ''")

	excluida := makeStatusCond([]string{"excluida"})
	assert.Contains(t, excluida, "l.excluded_at > ''")
	assert.NotContains(t, excluida, "coalesce(l.excluded_at, '') = ''")
}
