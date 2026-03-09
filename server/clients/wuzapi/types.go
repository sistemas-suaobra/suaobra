package wuzapi

type SessionQRResp struct {
	Code int `json:"code"`
	Data struct {
		QRCode string `json:"QRCode"`
	} `json:"data"`
}

type CreateAdminUserReq struct {
	Name    string `json:"name"`
	Token   string `json:"token"`
	Webhook string `json:"webhook"`
	Events  string `json:"events"`
	History int    `json:"history"`

	ProxyConfig struct {
		Enabled  bool   `json:"enabled"`
		ProxyURL string `json:"proxyURL"`
	} `json:"proxyConfig"`

	S3Config struct {
		Enabled       bool   `json:"enabled"`
		Endpoint      string `json:"endpoint"`
		Region        string `json:"region"`
		Bucket        string `json:"bucket"`
		AccessKey     string `json:"accessKey"`
		SecretKey     string `json:"secretKey"`
		PathStyle     bool   `json:"pathStyle"`
		PublicURL     string `json:"publicURL"`
		MediaDelivery string `json:"mediaDelivery"`
		RetentionDays int    `json:"retentionDays"`
	} `json:"s3Config"`
}

type CreateAdminUserResp struct {
	Code    int  `json:"code"`
	Success bool `json:"success"`
	Data    []struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Webhook    string `json:"webhook"`
		Events     string `json:"events"`
		Token      string `json:"token"`
		JID        string `json:"jid"`
		LoggedIn   bool   `json:"loggedIn"`
		Connected  bool   `json:"connected"`
		Expiration int    `json:"expiration"`
	} `json:"data"`
}

// AdminUserInfo representa um user retornado por GET /admin/users/:id
type AdminUserInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Token     string `json:"token"`
	Webhook   string `json:"webhook"`
	Events    string `json:"events"`
	Connected bool   `json:"connected"`
	LoggedIn  bool   `json:"loggedIn"`
	JID       string `json:"jid"`
	QRCode    string `json:"qrcode"`
}

type AdminUserResp struct {
	Code    int             `json:"code"`
	Data    []AdminUserInfo `json:"data"`
	Success bool            `json:"success"`
}

type AdminUsersResp = AdminUserResp