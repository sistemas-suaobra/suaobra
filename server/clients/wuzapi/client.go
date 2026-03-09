package wuzapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/flarco/g"
	"github.com/suaobra/suaobra-app/server/config"
)

type Client struct {
	cfg  config.WhatsMeowConfig
	http *http.Client
}

func NewClient(cfg config.WhatsMeowConfig) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: 25 * time.Second},
	}
}

func (c *Client) BaseURL() string {
	return c.cfg.BaseURL
}

func (c *Client) CreateAdminUser(name, token string) (CreateAdminUserResp, map[string]any, error) {
	var parsed CreateAdminUserResp

	if c.cfg.APIKey == "" {
		return parsed, nil, g.Error("WHATSMEOW_APIKEY not set")
	}

	reqBody := CreateAdminUserReq{
		Name:    name,
		Token:   token,
		Webhook: c.cfg.WebhookURL(),
		Events:  "All",
		History: 0,
	}
	reqBody.ProxyConfig.Enabled = false
	reqBody.S3Config.Enabled = false

	b, _ := json.Marshal(reqBody)

	url := c.cfg.BaseURL + "/admin/users"
	r, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return parsed, nil, err
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", c.cfg.APIKey)

	res, err := c.http.Do(r)
	if err != nil {
		return parsed, nil, err
	}
	defer res.Body.Close()

	var raw map[string]any
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return parsed, nil, g.Error("wuzapi returned non-json response (status=%d)", res.StatusCode)
	}

	rawBytes, _ := json.Marshal(raw)
	_ = json.Unmarshal(rawBytes, &parsed)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return parsed, raw, g.Error("wuzapi /admin/users failed status=%d raw=%v", res.StatusCode, raw)
	}

	if len(parsed.Data) == 0 || strings.TrimSpace(parsed.Data[0].ID) == "" {
		return parsed, raw, g.Error("wuzapi returned empty data.id: %v", raw)
	}

	item := parsed.Data[0]

	// wuzapi ignora "events" no POST /admin/users — forçamos via PUT
	if item.Events == "" || item.Events != "All" {
		_ = c.UpdateAdminUser(item.ID, map[string]any{
			"events":  "All",
			"webhook": c.cfg.WebhookURL(),
		})
		parsed.Data[0].Events = "All"
	}

	return parsed, raw, nil
}

func (c *Client) SessionConnect(userToken string) (map[string]any, error) {
	if userToken == "" {
		return nil, g.Error("empty user token")
	}

	body := map[string]any{"Immediate": true}
	b, _ := json.Marshal(body)

	url := c.cfg.BaseURL + "/session/connect"
	r, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	SetUserToken(r, userToken)

	res, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var raw map[string]any
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, g.Error("wuzapi returned non-json response (status=%d)", res.StatusCode)
	}
	return raw, nil
}

// GetAdminUser retorna os dados de um user wuzapi pelo ID (admin auth).
func (c *Client) GetAdminUser(userID string) (AdminUserInfo, error) {
	var info AdminUserInfo

	url := c.cfg.BaseURL + "/admin/users/" + userID
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return info, err
	}
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Authorization", c.cfg.APIKey)

	res, err := c.http.Do(r)
	if err != nil {
		return info, err
	}
	defer res.Body.Close()

	var parsed AdminUserResp
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return info, g.Error("wuzapi /admin/users/:id non-json (status=%d)", res.StatusCode)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return info, g.Error("wuzapi /admin/users/%s failed status=%d", userID, res.StatusCode)
	}
	if len(parsed.Data) == 0 {
		return info, g.Error("wuzapi /admin/users/%s returned empty data", userID)
	}
	return parsed.Data[0], nil
}

// UpdateAdminUser faz PUT no user — permite corrigir events/webhook após criação.
func (c *Client) UpdateAdminUser(userID string, fields map[string]any) error {
	b, _ := json.Marshal(fields)

	url := c.cfg.BaseURL + "/admin/users/" + userID
	r, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", c.cfg.APIKey)

	res, err := c.http.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return g.Error("wuzapi PUT /admin/users/%s failed status=%d", userID, res.StatusCode)
	}
	return nil
}

func (c *Client) SessionQR(userToken string) (SessionQRResp, map[string]any, error) {
	var parsed SessionQRResp

	if userToken == "" {
		return parsed, nil, g.Error("empty user token")
	}

	url := c.cfg.BaseURL + "/session/qr"
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return parsed, nil, err
	}

	r.Header.Set("Accept", "application/json")
	SetUserToken(r, userToken)

	res, err := c.http.Do(r)
	if err != nil {
		return parsed, nil, err
	}
	defer res.Body.Close()

	var raw map[string]any
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return parsed, nil, g.Error("wuzapi returned non-json response (status=%d)", res.StatusCode)
	}

	rawBytes, _ := json.Marshal(raw)
	_ = json.Unmarshal(rawBytes, &parsed)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return parsed, raw, g.Error("wuzapi /session/qr failed status=%d raw=%v", res.StatusCode, raw)
	}

	return parsed, raw, nil
}

// SendTextMessage envia uma mensagem de texto via WhatsApp.
func (c *Client) SendTextMessage(userToken, phone, body string) (map[string]any, error) {
	if userToken == "" {
		return nil, g.Error("empty user token")
	}

	reqBody := map[string]any{
		"Phone": phone,
		"Body":  body,
	}
	b, _ := json.Marshal(reqBody)

	url := c.cfg.BaseURL + "/chat/send/text"
	r, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	SetUserToken(r, userToken)

	res, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var raw map[string]any
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, g.Error("wuzapi returned non-json response (status=%d)", res.StatusCode)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return raw, g.Error("wuzapi /chat/send/text failed status=%d raw=%v", res.StatusCode, raw)
	}

	return raw, nil
}