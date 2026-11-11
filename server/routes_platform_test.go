package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeCityIDsKeepsMultipleStates(t *testing.T) {
	got := normalizeCityIDs([]string{"sp-*", "PR-*", "SP-CAMPINAS", "sp-*", "xx", "GO-GOIANIA"})
	assert.Equal(t, []string{"GO-GOIANIA", "PR-*", "SP-*", "SP-CAMPINAS"}, got)
}

func TestParseCitiesFieldFromJSON(t *testing.T) {
	got := parseCitiesField(`["SP-*","PR-*"]`)
	assert.Equal(t, []string{"PR-*", "SP-*"}, got)
}
