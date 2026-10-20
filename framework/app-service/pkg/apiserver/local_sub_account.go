package apiserver

import "github.com/beclab/Olares/framework/app-service/pkg/users"

// Local installations have no cloud domain and bind a mnemonic after creation.
func subAccountOlaresName(username, domain string) users.OlaresName {
	if domain == "" || domain == "olares.local" {
		return users.OlaresName(username)
	}
	return users.NewOlaresName(username, domain)
}
