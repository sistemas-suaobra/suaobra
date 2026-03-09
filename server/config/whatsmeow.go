package config

import "strings"

type WhatsMeowConfig struct {
	BaseURL     string
	APIKey      string
	WebhookBase string
}

func NewWhatsMeowConfig() WhatsMeowConfig {
	return WhatsMeowConfig{
		BaseURL:     EnvOr("WHATSMEOW_URL", "https://wuzapi-server-production.up.railway.app"),
		APIKey:      EnvOr("WHATSMEOW_APIKEY", "9F3C7A1E5B2D8C4F6A7E9D1B3C5F8A2D"),
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