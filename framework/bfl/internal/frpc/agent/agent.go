package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// Run consumes a controller-generated configuration. User and domain selection
// belong to BFL; an agent can also serve a future additional binding unchanged.
func Run(ctx context.Context, configPath, healthAddress string) error {
	cmd := exec.CommandContext(ctx, "/frpc", "-c", configPath)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	child := make(chan error, 1)
	go func() { child <- cmd.Wait() }()
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := Ready(r.Context()); err != nil {
			http.Error(w, "FRP proxies are not registered", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	server := &http.Server{Addr: healthAddress, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	serverErr := make(chan error, 1)
	go func() { serverErr <- server.ListenAndServe() }()
	defer server.Close()
	select {
	case err := <-child:
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("frpc exited: %v", err)
	case err := <-serverErr:
		_ = cmd.Process.Kill()
		<-child
		return err
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		<-child
		return nil
	}
}

func Ready(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:7400/api/status", nil)
	if err != nil {
		return err
	}
	resp, err := (&http.Client{Timeout: 2 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FRP status: %s", resp.Status)
	}
	var protocols map[string][]struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&protocols); err != nil {
		return err
	}
	count := 0
	for _, protocol := range []string{"http", "https"} {
		for _, proxy := range protocols[protocol] {
			if proxy.Status != "running" {
				return fmt.Errorf("FRP proxy is not running")
			}
			count++
		}
	}
	if count < 4 {
		return fmt.Errorf("FRP default routes are not registered")
	}
	return nil
}
