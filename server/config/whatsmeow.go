package config

import "strings"

type WhatsMeowConfig struct {
	BaseURL     string
	APIKey      string
	WebhookBase string
}

func NewWhatsMeowConfig() WhatsMeowConfig {
	return WhatsMeowConfig{
		BaseURL:     EnvOr("WHATSMEOW_URL", "http://whatsmeow:8081"),
		APIKey:      EnvOr("WHATSMEOW_APIKEY", ""),
		WebhookBase: EnvOr("WHATSMEOW_WEBHOOK_BASE", ""),
	}
}

func (c WhatsMeowConfig) WebhookURL() string {
	base := strings.TrimRight(c.WebhookBase, "/")
	if base == "" {
		return ""
	}
	return base + "/webhooks/whatsmeow"
}