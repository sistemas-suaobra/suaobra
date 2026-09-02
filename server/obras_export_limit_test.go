package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBrtDayStart(t *testing.T) {
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	now := time.Date(2026, 9, 2, 15, 30, 0, 0, loc)

	start := brtDayStart(now)

	assert.Equal(t, 2026, start.Year())
	assert.Equal(t, time.September, start.Month())
	assert.Equal(t, 2, start.Day())
	assert.Equal(t, 0, start.Hour())
	assert.Equal(t, 0, start.Minute())
}

func TestCapObrasExportRows_WithinDailyLimit(t *testing.T) {
	allowed, err := capObrasExportRows("", 100, time.Now())
	assert.NoError(t, err)
	assert.Equal(t, 100, allowed)
}
