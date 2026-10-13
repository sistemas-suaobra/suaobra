package server

import (
	"testing"

	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/stretchr/testify/assert"
)

func TestActiveFromProperties(t *testing.T) {
	assert.True(t, activeFromProperties(nil))
	assert.True(t, activeFromProperties(map[string]any{"name": "Ana"}))
	assert.True(t, activeFromProperties(`{"name":"Ana"}`))
	assert.False(t, activeFromProperties(map[string]any{"active": false}))
	assert.False(t, activeFromProperties(`{"name":"Ana","active":false}`))
	assert.False(t, activeFromProperties(types.JsonRaw(`{"active":false}`)))
	assert.True(t, activeFromProperties(types.JsonRaw(`{"active":true}`)))
}
