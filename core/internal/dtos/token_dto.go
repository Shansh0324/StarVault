package dtos

type TokenRequest struct {
	AppID     string `json:"appId"`
	AppSecret string `json:"appSecret"`
	Scope     string `json:"scope"`
}

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn"` // in seconds
}

type RevokeTokenRequest struct {
	Token string `json:"token"`
}
