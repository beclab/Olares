package callerjwt

import "github.com/prometheus/client_golang/prometheus"

var callerJWTIssueFailTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "olares_caller_jwt_issue_fail_total",
		Help: "Caller JWT issue failures by application namespace (e.g. missing spec.appid on Shared apps)",
	},
	[]string{"namespace"},
)

func init() {
	prometheus.MustRegister(callerJWTIssueFailTotal)
}

// RecordIssueFail increments the Shared/caller JWT issue failure counter.
func RecordIssueFail(namespace string) {
	if namespace == "" {
		namespace = "unknown"
	}
	callerJWTIssueFailTotal.WithLabelValues(namespace).Inc()
}
