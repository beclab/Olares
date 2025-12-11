package nginx

import "time"

const (
	// ProfilerPort port used by the ingress controller to expose the Go Profiler when it is enabled.
	ProfilerPort = 10245

	// TemplatePath path of the NGINX template
	TemplatePath = "/etc/nginx/template/nginx.tmpl"

	// PID defines the location of the pid file used by NGINX
	PID = "/run/nginx.pid"

	// StatusPort port used by NGINX for the status server
	StatusPort = 10246

	// StreamPort defines the port used by NGINX for the NGINX stream configuration socket
	StreamPort = 10247

	// HealthPath defines the path used to define the health check location in NGINX
	HealthPath = "/healthz"

	// HealthCheckTimeout defines the time limit in seconds for a probe to health-check-path to succeed
	HealthCheckTimeout = 10 * time.Second

	// StatusPath defines the path used to expose the NGINX status page
	// http://nginx.org/en/docs/http/ngx_http_stub_status_module.html
	StatusPath = "/nginx_status"

	// PermReadWriteByUser define nginx configuration file perm
	PermReadWriteByUser = 0700
)
