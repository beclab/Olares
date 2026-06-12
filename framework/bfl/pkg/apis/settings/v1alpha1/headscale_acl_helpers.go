package v1alpha1

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// ACLProto lists protocols accepted by the headscale ACL. If the ACL proto
// field is empty, headscale allows ICMPv4, ICMPv6, TCP, and UDP.
var ACLProto = sets.NewString(
	"", "igmp", "ipv4", "ip-in-ip", "tcp", "egp", "igp", "udp",
	"gre", "esp", "ah", "sctp", "icmp",
)

const expectedTokenItems = 2

var (
	ErrInvalidAction     = errors.New("invalid action")
	ErrInvalidPortFormat = errors.New("invalid port format")
)

// CheckTailScaleACLs validates the supplied ACL entries and fills in defaults.
func CheckTailScaleACLs(acls []appv1.ACL) error {
	if len(acls) == 0 {
		return nil
	}
	for i := range acls {
		acls[i].Action = "accept"
		acls[i].Src = []string{"*"}
	}
	for _, acl := range acls {
		if err := parseProtocol(acl.Proto); err != nil {
			return err
		}
		for _, dest := range acl.Dst {
			if _, _, err := parseDestination(dest); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseProtocol(protocol string) error {
	if ACLProto.Has(protocol) {
		return nil
	}
	return fmt.Errorf("unsupported protocol: %v", protocol)
}

// parseDestination from
// https://github.com/juanfont/headscale/blob/770f3dcb9334adac650276dcec90cd980af53c6e/hscontrol/policy/acls.go#L475
func parseDestination(dest string) (string, string, error) {
	var tokens []string

	// Check if there is a IPv4/6:Port combination, IPv6 has more than
	// three ":".
	tokens = strings.Split(dest, ":")
	if len(tokens) < expectedTokenItems || len(tokens) > 3 {
		port := tokens[len(tokens)-1]

		maybeIPv6Str := strings.TrimSuffix(dest, ":"+port)

		filteredMaybeIPv6Str := maybeIPv6Str
		if strings.Contains(maybeIPv6Str, "/") {
			networkParts := strings.Split(maybeIPv6Str, "/")
			filteredMaybeIPv6Str = networkParts[0]
		}

		if maybeIPv6, err := netip.ParseAddr(filteredMaybeIPv6Str); err != nil && !maybeIPv6.Is6() {
			return "", "", fmt.Errorf(
				"failed to parse destination, tokens %v: %w",
				tokens,
				ErrInvalidPortFormat,
			)
		} else {
			tokens = []string{maybeIPv6Str, port}
		}
	}

	var alias string
	// We can have here stuff like:
	// git-server:*
	// 192.168.1.0/24:22
	// fd7a:115c:a1e0::2:22
	// fd7a:115c:a1e0::2/128:22
	// tag:montreal-webserver:80,443
	// tag:api-server:443
	// example-host-1:*
	if len(tokens) == expectedTokenItems {
		alias = tokens[0]
	} else {
		alias = fmt.Sprintf("%s:%s", tokens[0], tokens[1])
	}

	return alias, tokens[len(tokens)-1], nil
}
