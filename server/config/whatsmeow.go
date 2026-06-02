package config

import "strings"

type WhatsMeowConfig struct {
	BaseURL     string
	APIKey      string
	WebhookBase string
}

// normalizeWuzAPIBaseURL garante uma única barra final para concatenação com paths (ex.: admin/users).
func normalizeWuzAPIBaseURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return u
	}
	return strings.TrimRight(u, "/") + "/"
}

func NewWhatsMeowConfig() WhatsMeowConfig {
	return WhatsMeowConfig{
		BaseURL: normalizeWuzAPIBaseURL(EnvOr("WHATSMEOW_URL", "http://31.97.16.33:8084")),
		APIKey:  EnvOr("WHATSMEOW_APIKEY", "9F3C7A1E5B2D8C4F6A7E9D1B3C5F8A2D"),
		WebhookBase: EnvOr("WHATSMEOW_WEBHOOK_BASE", "https://api-hml.suaobra.com.br"),
	}
}

func (c WhatsMeowConfig) WebhookURL() string {
	base := strings.TrimRight(c.WebhookBase, "/")
	if base == "" {
		return ""
	}
	return base + "/webhooks/whatsmeow"
}