package infisical

type ErrorMessage struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Error      string `json:"error"`
}

type Header struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type InfisicalResp struct {
	Header
	Data InfisicalData `json:"data"`
}

type InfisicalListResp struct {
	Header
	Data []*InfisicalData `json:"data"`
}

type InfisicalData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Env   string `json:"env"`
}

// cookie
type CookieData struct {
	Domain  string           `json:"domain"`
	Account string           `json:"account"`
	Records []*CookieRecords `json:"records"`
}

type CookieRecords struct {
	Domain   string  `json:"domain"`
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Expires  float64 `json:"expires"`
	Path     string  `json:"path"`
	Secure   bool    `json:"secure"`
	HttpOnly bool    `json:"httpOnly"`
	SameSite string  `json:"sameSite"`
	Other    string  `json:"other"`
}

// account
type AccountDataRaw struct {
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	Endpoint     string `json:"endpoint"`
	Bucket       string `json:"bucket"`
	UserId       string `json:"userid"`
	Available    bool   `json:"available"`
	CreateAt     int64  `json:"create_at"`
	Scope        string `json:"scope"`
	IdToken      string `json:"id_token"`
	ClientId     string `json:"client_id"`
}
