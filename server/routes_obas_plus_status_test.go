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

func TestMakeStatusCond_EmAndamento(t *testing.T) {
	cond := strings.ToLower(makeStatusCond([]string{"em-andamento"}))

	assert.Contains(t, cond, "execucao")
	assert.Contains(t, cond, "date(cop.start_date) <= date('now')")
	assert.Contains(t, cond, "date(cop.end_date) >= date('now')")
	assert.NotContains(t, cond, "current_date")
}

func TestMakeOrderSQL(t *testing.T) {
	desc, ok := makeOrderSQL("first_listing_date-desc,start_date-desc")
	assert.True(t, ok)
	assert.Equal(t, "first_listing_date DESC, start_date DESC, obra_number DESC", desc)

	asc, ok := makeOrderSQL("first_listing_date-asc,start_date-asc")
	assert.True(t, ok)
	assert.Equal(t, "first_listing_date ASC, start_date ASC, obra_number ASC", asc)

	sizeDesc, ok := makeOrderSQL("size-desc")
	assert.True(t, ok)
	assert.Equal(t, "size DESC, obra_number DESC", sizeDesc)

	_, ok = makeOrderSQL("first_listing_date-desc")
	assert.False(t, ok)
}

func TestMakeEtapaCond(t *testing.T) {
	todos := makeEtapaCond([]string{"todos"})
	assert.Equal(t, "1=1", todos)

	inicio := strings.ToLower(makeEtapaCond([]string{"inicio"}))
	assert.Contains(t, inicio, "not like '%projeto%'")
	assert.Contains(t, inicio, "date('now') < date(cop.start_date)")
	assert.Contains(t, inicio, "<= 0.33")

	estrutura := makeEtapaCond([]string{"estrutura"})
	assert.Contains(t, estrutura, "> 0.33")
	assert.Contains(t, estrutura, "<= 0.66")

	acabamento := makeEtapaCond([]string{"acabamento"})
	assert.Contains(t, acabamento, "> 0.66")

	finalizada := makeEtapaCond([]string{"finalizada"})
	assert.Contains(t, finalizada, "date('now') > date(cop.end_date)")

	multi := makeEtapaCond([]string{"inicio", "estrutura"})
	assert.Contains(t, multi, " or ")
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
