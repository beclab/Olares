/*

 Copyright 2021 The KubeSphere Authors.

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

package authentication

import (
	"time"

	_ "kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider/aliyunidaas"
	_ "kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider/cas"
	_ "kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider/github"
	_ "kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider/ldap"
	_ "kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider/oidc"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
)

type Options struct {
	// AuthenticateRateLimiter defines under which circumstances we will block user.
	// A user will be blocked if his/her failed login attempt reaches AuthenticateRateLimiterMaxTries in
	// AuthenticateRateLimiterDuration for about AuthenticateRateLimiterDuration. For example,
	//   AuthenticateRateLimiterMaxTries: 5
	//   AuthenticateRateLimiterDuration: 10m
	// A user will be blocked for 10m if he/she logins with incorrect credentials for at least 5 times in 10m.
	AuthenticateRateLimiterMaxTries int           `json:"authenticateRateLimiterMaxTries" yaml:"authenticateRateLimiterMaxTries"`
	AuthenticateRateLimiterDuration time.Duration `json:"authenticateRateLimiterDuration" yaml:"authenticateRateLimiterDuration"`
	// Token verification maximum time difference, default to 10s.
	// You should consider allowing a clock skew when checking the time-based values.
	// This should be values of a few seconds, and we don’t recommend using more than 30 seconds for this purpose,
	// as this would rather indicate problems with the server, rather than a common clock skew.
	MaximumClockSkew time.Duration `json:"maximumClockSkew" yaml:"maximumClockSkew"`
	// retention login history, records beyond this amount will be deleted
	LoginHistoryRetentionPeriod time.Duration `json:"loginHistoryRetentionPeriod" yaml:"loginHistoryRetentionPeriod"`
	// retention login history, records beyond this amount will be deleted
	// LoginHistoryMaximumEntries restricts for all kubesphere accounts and must be greater than AuthenticateRateLimiterMaxTries
	LoginHistoryMaximumEntries int `json:"loginHistoryMaximumEntries,omitempty" yaml:"loginHistoryMaximumEntries,omitempty"`
	// allow multiple users login from different location at the same time
	MultipleLogin bool `json:"multipleLogin" yaml:"multipleLogin"`
	// secret to sign jwt token
	JwtSecret string `json:"-" yaml:"jwtSecret"`
	// OAuthOptions defines options needed for integrated oauth plugins
	OAuthOptions *oauth.Options `json:"oauthOptions" yaml:"oauthOptions"`
	// KubectlImage is the image address we use to create kubectl pod for users who have admin access to the cluster.
	KubectlImage string `json:"kubectlImage" yaml:"kubectlImage"`
}

func NewOptions() *Options {
	return &Options{
		AuthenticateRateLimiterMaxTries: 5,
		AuthenticateRateLimiterDuration: time.Minute * 30,
		MaximumClockSkew:                10 * time.Second,
		LoginHistoryRetentionPeriod:     time.Hour * 24 * 7,
		LoginHistoryMaximumEntries:      100,
		OAuthOptions:                    oauth.NewOptions(),
		MultipleLogin:                   false,
		JwtSecret:                       "",
		KubectlImage:                    "kubesphere/kubectl:v1.22.0",
	}
}
