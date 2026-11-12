package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeFeedbackAttachments(t *testing.T) {
	got := normalizeFeedbackAttachments([]any{
		map[string]any{"name": "print.png", "kind": "foto", "key": "feedback/u1/a.png", "url": "/api/feedback/attachments?v=a"},
		map[string]any{"name": "  ", "kind": "foto"},
		map[string]any{"name": "nota.pdf"},
	})
	assert.Equal(t, []map[string]string{
		{"name": "print.png", "kind": "foto", "key": "feedback/u1/a.png", "url": "/api/feedback/attachments?v=a"},
		{"name": "nota.pdf", "kind": "arquivo"},
	}, got)
}

func TestParseFeedbackAttachmentsField(t *testing.T) {
	got := parseFeedbackAttachmentsField(`[{"name":"a.mp4","kind":"video"}]`)
	assert.Equal(t, []map[string]string{{"name": "a.mp4", "kind": "video"}}, got)
	assert.Empty(t, parseFeedbackAttachmentsField(""))
	assert.Empty(t, parseFeedbackAttachmentsField(nil))
}
