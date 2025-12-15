package lldap

type User struct {
	Id          string   `json:"id"`
	Email       string   `json:"email"`
	DisplayName string   `json:"name"`
	Groups      []string `json:"groupName"`
}

type UserSyncer interface {
	GetProviderName() string
	Sync() ([]User, error)
}
