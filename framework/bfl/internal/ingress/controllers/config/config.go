/*
Copyright 2016 The Kubernetes Authors.

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

package config

import (
	"math"
	"os"
	"runtime"
	"strconv"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const (
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size
	// Sets the maximum allowed size of the client request body
	bodySize = "1m"

	// http://nginx.org/en/docs/ngx_core_module.html#error_log
	// Configures logging level [debug | info | notice | warn | error | crit | alert | emerg]
	// Log levels above are listed in the order of increasing severity
	errorLevel = "notice"

	gzipTypes = "application/atom+xml application/javascript application/x-javascript application/json application/rss+xml application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon text/css text/javascript text/plain text/x-component application/x-mpegURL"

	logFormatUpstream = `$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_length $request_time [$proxy_upstream_name] [$proxy_alternative_upstream_name] $upstream_addr $upstream_response_length $upstream_response_time $upstream_status $req_id`

	logFormatMain = `'$remote_addr - $remote_user [$time_local] "$http_host" "$request" ' '$status $body_bytes_sent "$http_referer" ' '"$http_user_agent" $request_length $request_time "$http_x_forwarded_for" ' '$upstream_addr $upstream_status $upstream_bytes_sent $upstream_response_time'`

	logFormatStream = `[$remote_addr] [$time_local] $protocol $status $bytes_sent $bytes_received $session_time`

	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size
	// Sets the size of the buffer used for sending data.
	// 4k helps NGINX to improve TLS Time To First Byte (TTTFB)
	// https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/
	sslBufferSize = "4k"

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by the OpenSSL library
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	sslCiphers = "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384"

	// SSL enabled protocols to use
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols
	sslProtocols = "TLSv1.2 TLSv1.3"

	// Time during which a client may reuse the session parameters stored in a cache.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout
	sslSessionTimeout = "10m"

	// Size of the SSL shared cache between all worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	sslSessionCacheSize = "10m"

	// Parameters for a shared memory zone that will keep states for various keys.
	// http://nginx.org/en/docs/http/ngx_http_limit_conn_module.html#limit_conn_zone
	defaultLimitConnZoneVariable = "$binary_remote_addr"

	proxyConnectTimeout = "30s"

	proxyReadTimeout = "1800s"

	proxySendTimeout = "60s"

	otelAgentLib = "/opt/opentelemetry-webserver/agent/WebServerModule/Nginx/1.25.3/ngx_http_opentelemetry_module.so"
)

// Configuration represents the content of nginx.conf file
type Configuration struct {
	// Sets the name of the configmap that contains the headers to pass to the client
	AddHeaders string `json:"add-headers,omitempty"`

	// AccessLogPath sets the path of the access logs for both http and stream contexts if enabled
	// http://nginx.org/en/docs/http/ngx_http_log_module.html#access_log
	// http://nginx.org/en/docs/stream/ngx_stream_log_module.html#access_log
	// By default access logs go to /var/log/nginx/access.log
	AccessLogPath string `json:"access-log-path,omitempty"`

	// WorkerCPUAffinity bind nginx worker processes to CPUs this will improve response latency
	// http://nginx.org/en/docs/ngx_core_module.html#worker_cpu_affinity
	// By default this is disabled
	WorkerCPUAffinity string `json:"worker-cpu-affinity,omitempty"`
	// ErrorLogPath sets the path of the error logs
	// http://nginx.org/en/docs/ngx_core_module.html#error_log
	// By default error logs go to /var/log/nginx/error.log
	ErrorLogPath string `json:"error-log-path,omitempty"`

	// ClientHeaderBufferSize allows to configure a custom buffer
	// size for reading client request header
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_header_buffer_size
	ClientHeaderBufferSize string `json:"client-header-buffer-size"`

	// Defines a timeout for reading client request header, in seconds
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_header_timeout
	ClientHeaderTimeout int `json:"client-header-timeout,omitempty"`

	// Sets buffer size for reading client request body
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_buffer_size
	ClientBodyBufferSize string `json:"client-body-buffer-size,omitempty"`

	// Defines a timeout for reading client request body, in seconds
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_timeout
	ClientBodyTimeout int `json:"client-body-timeout,omitempty"`

	// EnableUnderscoresInHeaders enables underscores in header names
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#underscores_in_headers
	// By default this is disabled
	EnableUnderscoresInHeaders bool `json:"enable-underscores-in-headers"`

	// IgnoreInvalidHeaders set if header fields with invalid names should be ignored
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#ignore_invalid_headers
	// By default this is enabled
	IgnoreInvalidHeaders bool `json:"ignore-invalid-headers"`

	// http://nginx.org/en/docs/ngx_core_module.html#error_log
	// Configures logging level [debug | info | notice | warn | error | crit | alert | emerg]
	// Log levels above are listed in the order of increasing severity
	ErrorLogLevel string `json:"error-log-level,omitempty"`

	// Time during which a keep-alive client connection will stay open on the server side.
	// The zero value disables keep-alive client connections
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout
	KeepAlive int `json:"keep-alive,omitempty"`

	// Sets the maximum number of requests that can be served through one keep-alive connection.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_requests
	KeepAliveRequests int `json:"keep-alive-requests,omitempty"`

	// LargeClientHeaderBuffers Sets the maximum number and size of buffers used for reading
	// large client request header.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#large_client_header_buffers
	// Default: 4 8k
	LargeClientHeaderBuffers string `json:"large-client-header-buffers"`

	// If disabled, a worker process will accept one new connection at a time.
	// Otherwise, a worker process will accept all new connections at a time.
	// http://nginx.org/en/docs/ngx_core_module.html#multi_accept
	// Default: true
	EnableMultiAccept bool `json:"enable-multi-accept,omitempty"`

	// Maximum number of simultaneous connections that can be opened by each worker process
	// http://nginx.org/en/docs/ngx_core_module.html#worker_connections
	MaxWorkerConnections int `json:"max-worker-connections,omitempty"`

	// Maximum number of files that can be opened by each worker process.
	// http://nginx.org/en/docs/ngx_core_module.html#worker_rlimit_nofile
	MaxWorkerOpenFiles int `json:"max-worker-open-files,omitempty"`

	// Sets the bucket size for the map variables hash tables.
	// Default value depends on the processor’s cache line size.
	// http://nginx.org/en/docs/http/ngx_http_map_module.html#map_hash_bucket_size
	MapHashBucketSize int `json:"map-hash-bucket-size,omitempty"`

	// Maximum size of the server names hash tables used in server names, map directive’s values,
	// MIME types, names of request header strings, etcd.
	// http://nginx.org/en/docs/hash.html
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_max_size
	ServerNameHashMaxSize int `json:"server-name-hash-max-size,omitempty"`

	// Size of the bucket for the server names hash tables
	// http://nginx.org/en/docs/hash.html
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_names_hash_bucket_size
	ServerNameHashBucketSize int `json:"server-name-hash-bucket-size,omitempty"`

	// Size of the bucket for the proxy headers hash tables
	// http://nginx.org/en/docs/hash.html
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_headers_hash_max_size
	ProxyHeadersHashMaxSize int `json:"proxy-headers-hash-max-size,omitempty"`

	// Maximum size of the bucket for the proxy headers hash tables
	// http://nginx.org/en/docs/hash.html
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_headers_hash_bucket_size
	ProxyHeadersHashBucketSize int `json:"proxy-headers-hash-bucket-size,omitempty"`

	// Enables or disables emitting nginx version in error messages and in the “Server” response header field.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#server_tokens
	// Default: false
	ShowServerTokens bool `json:"server-tokens"`

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by
	// the OpenSSL library
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	SSLCiphers string `json:"ssl-ciphers,omitempty"`

	// SSL enabled protocols to use
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_protocols
	SSLProtocols string `json:"ssl-protocols,omitempty"`

	// Enables or disables the use of shared SSL cache among worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	SSLSessionCache bool `json:"ssl-session-cache,omitempty"`

	// Size of the SSL shared cache between all worker processes.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_cache
	SSLSessionCacheSize string `json:"ssl-session-cache-size,omitempty"`

	// Enables or disables session resumption through TLS session tickets.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets
	SSLSessionTickets bool `json:"ssl-session-tickets,omitempty"`

	// Sets the secret key used to encrypt and decrypt TLS session tickets.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_tickets
	// By default, a randomly generated key is used.
	// Example: openssl rand 80 | openssl enc -A -base64
	SSLSessionTicketKey string `json:"ssl-session-ticket-key,omitempty"`

	// Time during which a client may reuse the session parameters stored in a cache.
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_session_timeout
	SSLSessionTimeout string `json:"ssl-session-timeout,omitempty"`

	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size
	// Sets the size of the buffer used for sending data.
	// 4k helps NGINX to improve TLS Time To First Byte (TTTFB)
	// https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/
	SSLBufferSize string `json:"ssl-buffer-size,omitempty"`

	// Enables or disables the use of the PROXY protocol to receive client connection
	// (real IP address) information passed through proxy servers and load balancers
	// such as HAproxy and Amazon Elastic Load Balancer (ELB).
	// https://www.nginx.com/resources/admin-guide/proxy-protocol/
	UseProxyProtocol bool `json:"use-proxy-protocol,omitempty"`

	// When use-proxy-protocol is enabled, sets the maximum time the connection handler will wait
	// to receive proxy headers.
	// Example '60s'
	ProxyProtocolHeaderTimeout time.Duration `json:"proxy-protocol-header-timeout,omitempty"`

	// Enables or disables the use of the nginx module that compresses responses using the "gzip" method
	// http://nginx.org/en/docs/http/ngx_http_gzip_module.html
	UseGzip bool `json:"use-gzip,omitempty"`

	// gzip Compression Level that will be used
	GzipLevel int `json:"gzip-level,omitempty"`

	// Minimum length of responses to be sent to the client before it is eligible
	// for gzip compression, in bytes.
	GzipMinLength int `json:"gzip-min-length,omitempty"`

	// MIME types in addition to "text/html" to compress. The special value “*” matches any MIME type.
	// Responses with the “text/html” type are always compressed if UseGzip is enabled
	GzipTypes string `json:"gzip-types,omitempty"`

	// Enables or disables the HTTP/2 support in secure connections
	// http://nginx.org/en/docs/http/ngx_http_v2_module.html
	// Default: true
	UseHTTP2 bool `json:"use-http2,omitempty"`

	// https://nginx.org/en/docs/http/ngx_http_v2_module.html#http2_max_field_size
	// HTTP2MaxFieldSize Limits the maximum size of an HPACK-compressed request header field
	HTTP2MaxFieldSize string `json:"http2-max-field-size,omitempty"`

	// https://nginx.org/en/docs/http/ngx_http_v2_module.html#http2_max_header_size
	// HTTP2MaxHeaderSize Limits the maximum size of the entire request header list after HPACK decompression
	HTTP2MaxHeaderSize string `json:"http2-max-header-size,omitempty"`

	// http://nginx.org/en/docs/http/ngx_http_v2_module.html#http2_max_requests
	// HTTP2MaxRequests Sets the maximum number of requests (including push requests) that can be served
	// through one HTTP/2 connection, after which the next client request will lead to connection closing
	// and the need of establishing a new connection.
	HTTP2MaxRequests int `json:"http2-max-requests,omitempty"`

	// http://nginx.org/en/docs/http/ngx_http_v2_module.html#http2_max_concurrent_streams
	// Sets the maximum number of concurrent HTTP/2 streams in a connection.
	HTTP2MaxConcurrentStreams int `json:"http2-max-concurrent-streams,omitempty"`

	// Defines the number of worker processes. By default auto means number of available CPU cores
	// http://nginx.org/en/docs/ngx_core_module.html#worker_processes
	WorkerProcesses string `json:"worker-processes,omitempty"`

	// Defines a timeout for a graceful shutdown of worker processes
	// http://nginx.org/en/docs/ngx_core_module.html#worker_shutdown_timeout
	WorkerShutdownTimeout string `json:"worker-shutdown-timeout,omitempty"`

	// Sets the bucket size for the variables hash table.
	// http://nginx.org/en/docs/http/ngx_http_map_module.html#variables_hash_bucket_size
	VariablesHashBucketSize int `json:"variables-hash-bucket-size,omitempty"`

	// Sets the maximum size of the variables hash table.
	// http://nginx.org/en/docs/http/ngx_http_map_module.html#variables_hash_max_size
	VariablesHashMaxSize int `json:"variables-hash-max-size,omitempty"`

	// Modifies the HTTP version the proxy uses to interact with the backend.
	// http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_http_version
	ProxyHTTPVersion string `json:"proxy-http-version"`

	// Sets the ipv4 addresses on which the server will accept requests.
	BindAddressIpv4 []string `json:"bind-address-ipv4,omitempty"`

	// Sets the ipv6 addresses on which the server will accept requests.
	BindAddressIpv6 []string `json:"bind-address-ipv6,omitempty"`

	// Sets whether to use incoming X-Forwarded headers.
	UseForwardedHeaders bool `json:"use-forwarded-headers"`

	// Sets whether to enable the real ip module
	EnableRealIp bool `json:"enable-real-ip"`

	// Sets the header field for identifying the originating IP address of a client
	// Default is X-Forwarded-For
	ForwardedForHeader string `json:"forwarded-for-header,omitempty"`

	// Append the remote address to the X-Forwarded-For header instead of replacing it
	// Default: false
	ComputeFullForwardedFor bool `json:"compute-full-forwarded-for,omitempty"`

	// Adds an X-Original-Uri header with the original request URI to the backend request
	// Default: true
	ProxyAddOriginalURIHeader bool `json:"proxy-add-original-uri-header"`

	// ReusePort instructs NGINX to create an individual listening socket for
	// each worker process (using the SO_REUSEPORT socket option), allowing a
	// kernel to distribute incoming connections between worker processes
	// Default: true
	ReusePort bool `json:"reuse-port"`

	// HideHeaders sets additional header that will not be passed from the upstream
	// server to the client response
	// Default: empty
	HideHeaders []string `json:"hide-headers"`

	// Checksum contains a checksum of the configmap configuration
	Checksum string `json:"-"`

	// Block all requests from given IPs
	BlockCIDRs []string `json:"block-cidrs"`

	// Block all requests with given User-Agent headers
	BlockUserAgents []string `json:"block-user-agents"`

	// Block all requests with given Referer headers
	BlockReferers []string `json:"block-referers"`

	// DefaultType Sets the default MIME type of a response.
	// http://nginx.org/en/docs/http/ngx_http_core_module.html#default_type
	// Default: text/html
	DefaultType string `json:"default-type"`

	// GlobalRateLimitMemcachedHost configures memcached host.
	GlobalRateLimitMemcachedHost string `json:"global-rate-limit-memcached-host"`

	// GlobalRateLimitMemcachedPort configures memcached port.
	GlobalRateLimitMemcachedPort int `json:"global-rate-limit-memcached-port"`

	// GlobalRateLimitMemcachedConnectTimeout configures timeout when connecting to memcached.
	// The unit is millisecond.
	GlobalRateLimitMemcachedConnectTimeout int `json:"global-rate-limit-memcached-connect-timeout"`

	// GlobalRateLimitMemcachedMaxIdleTimeout configured how long connections
	// should be kept alive in idle state. The unit is millisecond.
	GlobalRateLimitMemcachedMaxIdleTimeout int `json:"global-rate-limit-memcached-max-idle-timeout"`

	// GlobalRateLimitMemcachedPoolSize configures how many connections
	// should be kept alive in the pool.
	// Note that this is per NGINX worker. Make sure your memcached server can
	// handle `MemcachedPoolSize * <nginx worker count> * <nginx replica count>`
	// simultaneous connections.
	GlobalRateLimitMemcachedPoolSize int `json:"global-rate-limit-memcached-pool-size"`

	// GlobalRateLimitStatucCode determines the HTTP status code to return
	// when limit is exceeding during global rate limiting.
	GlobalRateLimitStatucCode int `json:"global-rate-limit-status-code"`

	LogHTTPProxyFormat string `json:"log-http-proxy-format"`

	// Proxy timeout
	ProxyConnectTimeout string `json:"proxy-connect-timeout"`
	ProxyReadTimeout    string `json:"proxy-read-timeout"`
	ProxySendTimeout    string `json:"proxy-send-timeout"`

	EnableOtel bool `json:"enable-otel"`
}

// NewDefault returns the default nginx configuration
func NewDefault() Configuration {
	defIPCIDR := make([]string, 0)
	defBindAddress := make([]string, 0)
	defBlockEntity := make([]string, 0)
	defNginxStatusIpv4Whitelist := make([]string, 0)
	defNginxStatusIpv6Whitelist := make([]string, 0)

	defIPCIDR = append(defIPCIDR, "0.0.0.0/0")
	defNginxStatusIpv4Whitelist = append(defNginxStatusIpv4Whitelist, "127.0.0.1")
	defNginxStatusIpv6Whitelist = append(defNginxStatusIpv6Whitelist, "::1")
	defProxyDeadlineDuration := time.Duration(5) * time.Second

	cfg := Configuration{
		AccessLogPath:              "/var/log/nginx/access.log",
		WorkerCPUAffinity:          "",
		ErrorLogPath:               "/var/log/nginx/error.log",
		BlockCIDRs:                 defBlockEntity,
		BlockUserAgents:            defBlockEntity,
		BlockReferers:              defBlockEntity,
		ClientHeaderBufferSize:     "10k",
		ClientHeaderTimeout:        60,
		ClientBodyBufferSize:       "10m",
		ClientBodyTimeout:          60,
		EnableUnderscoresInHeaders: false,
		ErrorLogLevel:              errorLevel,
		UseForwardedHeaders:        true,
		EnableRealIp:               false,
		ForwardedForHeader:         "X-Forwarded-For",
		ComputeFullForwardedFor:    true,
		ProxyAddOriginalURIHeader:  false,
		HTTP2MaxFieldSize:          "4k",
		HTTP2MaxHeaderSize:         "16k",
		HTTP2MaxRequests:           1000,
		HTTP2MaxConcurrentStreams:  128,
		IgnoreInvalidHeaders:       true,
		GzipLevel:                  1,
		GzipMinLength:              256,
		GzipTypes:                  gzipTypes,
		KeepAlive:                  75,
		KeepAliveRequests:          1000,
		LargeClientHeaderBuffers:   "6 10k",
		EnableMultiAccept:          true,
		MaxWorkerConnections:       16384,
		MaxWorkerOpenFiles:         65535,
		MapHashBucketSize:          64,
		ProxyProtocolHeaderTimeout: defProxyDeadlineDuration,
		ServerNameHashBucketSize:   1024,
		ServerNameHashMaxSize:      4096,
		ProxyHeadersHashMaxSize:    512,
		ProxyHeadersHashBucketSize: 64,
		ShowServerTokens:           false,
		SSLBufferSize:              sslBufferSize,
		UseProxyProtocol:           true,
		SSLCiphers:                 sslCiphers,
		SSLProtocols:               sslProtocols,
		SSLSessionCache:            true,
		SSLSessionCacheSize:        sslSessionCacheSize,
		SSLSessionTickets:          false,
		SSLSessionTimeout:          sslSessionTimeout,
		UseGzip:                    true,
		WorkerShutdownTimeout:      "240s",
		VariablesHashBucketSize:    256,
		VariablesHashMaxSize:       2048,
		UseHTTP2:                   true,
		BindAddressIpv4:            defBindAddress,
		BindAddressIpv6:            defBindAddress,
		DefaultType:                "text/html",
		LogHTTPProxyFormat:         logFormatMain,
		ProxyConnectTimeout:        proxyConnectTimeout,
		ProxySendTimeout:           proxySendTimeout,
		ProxyReadTimeout:           proxyReadTimeout,
		EnableOtel:                 false,
	}
	workerProcesses := int(math.Ceil(float64(runtime.NumCPU() / 2)))
	if workerProcesses < 1 {
		workerProcesses = 1
	}
	cfg.WorkerProcesses = strconv.Itoa(workerProcesses)

	if klog.V(5).Enabled() {
		cfg.ErrorLogLevel = "debug"
	}

	if _, err := os.Stat(otelAgentLib); err == nil {
		cfg.EnableOtel = true
	}

	return cfg
}

// TemplateConfig contains the nginx configuration to render the file nginx.conf
type TemplateConfig struct {
	ProxySetHeaders         map[string]string
	AddHeaders              map[string]string
	BacklogSize             int
	HealthzURI              string
	IsIPV6Enabled           bool
	IsSSLPassthroughEnabled bool
	ListenPorts             *ListenPorts
	PublishService          *apiv1.Service
	EnableMetrics           bool
	MaxmindEditionFiles     []string
	MonitorMaxBatchSize     int

	Cfg Configuration

	Servers             []Server
	CustomDomainServers []CustomServer
	StreamServers       []StreamServer

	PID        string
	StatusPath string
	StatusPort int
	StreamPort int

	SSLCertificatePath    string
	SSLCertificateKeyPath string

	UserName        string
	IsEphemeralUser bool
	UserZone        string
	RealIpFrom      []string
}

type Server struct {
	Hostname string   `json:"hostname" yaml:"hostname"`
	Aliases  []string `json:"aliases" yaml:"aliases"`

	Port int `json:"port" yaml:"port"`

	EnableSSL bool `json:"ssl-enable" yaml:"enableSSL"`

	Locations []Location `json:"locations,omitempty"`

	EnableAuth            bool   `json:"enableAuth"`
	EnableOIDC            bool   `json:"enableOIDC"`
	EnableWindowPushState bool   `json:"enableWindowPushState"`
	Language              string `json:"language"`
}

type StreamServer struct {
	Protocol  string `json:"protocol" yaml:"protocol"`
	Port      int32  `json:"port" yaml:"port"`
	ProxyPass string `json:"proxyPass" yaml:"proxyPass"`
}

type CustomServer struct {
	Server
	SslCertPath string `json:"sslCertPath"`
	SslKeyPath  string `json:"sslKeyPath"`
}

type Location struct {
	Prefix      string   `json:"prefix" yaml:"prefix"`
	ProxyPass   string   `json:"proxy-pass" yaml:"proxyPass"`
	Additionals []string `json:"additionals,omitempty" yaml:"additionals,omitempty"`
	DirectProxy bool     `json:"direct-proxy,omitempty" yaml:"directProxy,omitempty"`
}

// ListenPorts describe the ports required to run the
// NGINX Ingress controller
type ListenPorts struct {
	HTTP     int
	HTTPS    int
	Health   int
	Default  int
	SSLProxy int
}
