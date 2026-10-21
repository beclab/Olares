/*
Copyright 2026 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package filters

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	k8srequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/rest"

	"kubesphere.io/kubesphere/pkg/apiserver/request"
)

type recordingResponder struct{ err error }

func (r *recordingResponder) Error(w http.ResponseWriter, _ *http.Request, err error) {
	r.err = err
	w.WriteHeader(http.StatusBadGateway)
}

// An exec upgrade proxied to kube-apiserver must reach it exactly once; each
// arrival runs the command in the container.
func TestWithKubeAPIServerUpgradesReachTheBackendOnce(t *testing.T) {
	var hits int32
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/exec") {
			http.NotFound(w, r)
			return
		}
		atomic.AddInt32(&hits, 1)
		conn, rw, err := w.(http.Hijacker).Hijack()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = rw.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n")
		_ = rw.Flush()
	}))
	defer backend.Close()

	responder := &recordingResponder{}
	handler := WithKubeAPIServer(http.NotFoundHandler(), &rest.Config{Host: backend.URL}, responder)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := &request.RequestInfo{RequestInfo: &k8srequest.RequestInfo{}, IsKubernetesRequest: true}
		handler.ServeHTTP(w, r.WithContext(request.WithRequestInfo(r.Context(), info)))
	}))
	defer proxy.Close()

	conn, err := net.Dial("tcp", strings.TrimPrefix(proxy.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, _ = conn.Write([]byte("GET /api/v1/namespaces/ns/pods/p/exec?command=true HTTP/1.1\r\nHost: proxy\r\nConnection: Upgrade\r\nUpgrade: websocket\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n\r\n"))
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("status %d, responder error %v", resp.StatusCode, responder.err)
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("kube-apiserver received the exec upgrade %d times, want 1", got)
	}
}
