package integration

import (
	"olares.com/backup-server/pkg/apiserver/response"
	"olares.com/backup-server/pkg/util"
)

type SpaceToken struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	OlaresDid   string `json:"olares_did"`
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
	Available   bool   `json:"available"`
}

func (s *SpaceToken) Expired() bool {
	return util.IsTimestampNearingExpiration(s.ExpiresAt)
}

type IntegrationToken struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Endpoint  string `json:"endpoint"`
	Bucket    string `json:"bucket"`
	Available bool   `json:"available"`
}

type accountResponse struct {
	response.Header
	Data *accountResponseData `json:"data,omitempty"`
}

type accountResponseData struct {
	Name    string                  `json:"name"`
	Type    string                  `json:"type"`
	RawData *accountResponseRawData `json:"rawData"`
}

type accountResponseRawData struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
	CreateAt     int64  `json:"createAt"`
	Available    bool   `json:"available"`
	Endpoint     string `json:"endpoint"`
	Bucket       string `json:"bucket"`
	UserId       string `json:"userId"`
	CloudUrl     string `json:"cloudUrl"`
}

type accountsResponse struct {
	response.Header
	Data []*accountsResponseData `json:"data,omitempty"`
}

type accountsResponseData struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Available bool   `json:"available"`
	CreateAt  int64  `json:"createAt"`
	ExpiresAt int64  `json:"expiresAt"`
}
