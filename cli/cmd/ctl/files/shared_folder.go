package files

import (
	"encoding/json"
	"fmt"
	"strings"
)

// sharedFolderConflict reports whether a response is the files
// backend refusing to rename, move or delete a folder that is shared,
// or holds one: HTTP 409 with the shares in the body. It returns a
// message naming them and the flag that lets the change go ahead.
// A 409 without a share list (a name clash) is not one.
func sharedFolderConflict(status int, body string) (string, bool) {
	if status != 409 {
		return "", false
	}
	var resp struct {
		Message string `json:"message"`
		Shares  []struct {
			ShareType string `json:"share_type"`
			Owner     string `json:"owner"`
			Path      string `json:"path"`
			Name      string `json:"name"`
		} `json:"shares"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil || resp.Shares == nil {
		return "", false
	}
	const max = 5
	var list []string
	for i, sh := range resp.Shares {
		if i == max {
			list = append(list, fmt.Sprintf("and %d more", len(resp.Shares)-max))
			break
		}
		name := sh.Name
		if name == "" {
			name = sh.Path
		}
		kind := sh.ShareType
		if kind == "external" {
			kind = "public"
		}
		list = append(list, fmt.Sprintf("%s (%s, by %s)", name, kind, sh.Owner))
	}
	return fmt.Sprintf("the folder, or a folder in it, is shared: %s (HTTP 409). "+
		"Re-run with --unshare to cancel these shares and go ahead.", strings.Join(list, ", ")), true
}
