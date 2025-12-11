package rediscluster

import (
	"fmt"
	"regexp"
	"testing"
)

func TestUserAuthSection(t *testing.T) {
	t.Log(GetUserAuthSections([][]string{
		{"pwd1"},
		{"pwd2"},
	}))
}

func TestConfigUpdate(t *testing.T) {
	config := `
################################### GENERAL ####################################
Name PredixyServer
Bind 0.0.0.0:6379
WorkerThreads 2
ClientTimeout 300
LogVerbSample 0
LogDebugSample 1
LogInfoSample 10000
LogNoticeSample 1
LogWarnSample 1
LogErrorSample 1
################################### AUTHORITY ##################################
################################### SERVERS ####################################
ClusterServerPool {
	Password $password
	MasterReadPriority 60
	StaticSlaveReadPriority 50
	DynamicSlaveReadPriority 50
	RefreshInterval 10
	ServerTimeout 1
	ServerFailureLimit 10
	ServerRetryTimeout 1
	KeepAlive 120
	Servers {
		+ drc-redis-cluster-0-0.redis-cluster-0:6379
		+ drc-redis-cluster-0-1.redis-cluster-0:6379
	}
}
	`
	passwords := [][]string{
		{"pwd1111"},
		{"u2222ser2"},
	}
	configNew := regexp.MustCompile(AuthSectionExpr).ReplaceAllStringFunc(config, func(string) string {
		return AuthSectionExprPrev + "\n" + GetUserAuthSections(passwords) + "\n" + AuthSectionExprNext
	})

	fmt.Println(AuthSectionExpr)
	fmt.Println(configNew)

	passwords = [][]string{
		{"pwd1111"},
	}

	configNew = regexp.MustCompile(AuthSectionExpr).ReplaceAllStringFunc(configNew, func(string) string {
		return AuthSectionExprPrev + "\n" + GetUserAuthSections(passwords) + "\n" + AuthSectionExprNext
	})

	fmt.Println(configNew)
}
