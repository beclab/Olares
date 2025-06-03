package utils

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/hirochachacha/go-smb2"
	"k8s.io/klog/v2"
)

func ListSambaSharenames(ctx context.Context, server string, username string, password string) ([]string, error) {
	// check server is valid
	if server == "" {
		klog.Error("server is empty")
		return nil, errors.New("server is empty")
	}

	serverToken := strings.Split(server, ":")
	if len(serverToken) < 2 {
		serverToken = append(serverToken, "445") // default samba port is 445
	}

	dialServer := strings.Join(serverToken, ":")
	conn, err := net.DialTimeout("tcp", dialServer, 10*time.Second)
	if err != nil {
		klog.Error("connect to samba server error, ", err)
		return nil, err
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     username,
			Password: password,
		},
	}

	s, err := d.DialContext(ctx, conn)
	if err != nil {
		klog.Error("dial samba server with username and password error, ", err)
		return nil, err
	}
	defer s.Logoff()

	names, err := s.ListSharenames()
	if err != nil {
		klog.Error("list samba server share names error, ", err)
	}

	systemShare := func(name string) bool {
		return !strings.HasSuffix(name, "$")
	}

	filters := []func(string) bool{systemShare}

	var filteredNames []string
	for _, name := range names {
		for _, filter := range filters {
			if filter(name) {
				filteredNames = append(filteredNames, name)
				break
			}
		}
	}

	return filteredNames, err
}
