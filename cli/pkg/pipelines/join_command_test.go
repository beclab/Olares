package pipelines

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestBuildJoinCommand(t *testing.T) {
	authPayload := encodeMasterAuthInfo(&common.MasterHostConfig{
		MasterHost:        "192.168.1.15",
		MasterSSHUser:     "olares",
		MasterSSHPassword: "secret-password",
		MasterSSHPort:     2222,
	})
	command := buildJoinCommand(workerCommandSpec{
		Version:        "1.12.7",
		CDNService:     "https://cdn.olares.cn/",
		MasterAuthInfo: authPayload,
	})

	assert.Equal(t,
		"export MASTER_AUTH_INFO='"+authPayload+"' "+
			"OLARES_SYSTEM_CDN_SERVICE='https://cdn.olares.cn' && "+
			"curl -fsSL 'https://cdn.olares.cn/joincluster-v1.12.7.sh' | bash",
		command)
	assert.NotContains(t, command, "secret-password")
	// the version is in the URL, so passing it through the environment as well
	// would just be another thing that can disagree
	assert.NotContains(t, command, "VERSION=")
}

// The regional CDN decides where every download comes from, so it must be
// carried even when it happens to equal the compiled-in default: the worker
// cannot be relied on to guess it.
func TestBuildJoinCommandAlwaysCarriesCDN(t *testing.T) {
	command := buildJoinCommand(workerCommandSpec{
		Version:        "1.12.7",
		CDNService:     "https://cdn.olares.com",
		MasterAuthInfo: "payload",
	})
	assert.Contains(t, command, "OLARES_SYSTEM_CDN_SERVICE='https://cdn.olares.com'")
}

// The script the worker fetches must be the one released with the cluster's
// version, so the version has to reach the URL and not only the environment.
func TestBuildJoinCommandPinsScriptToVersion(t *testing.T) {
	command := buildJoinCommand(workerCommandSpec{
		Version:        "1.12.7-20260728",
		CDNService:     "https://cdn.olares.cn",
		MasterAuthInfo: "payload",
	})
	assert.Contains(t, command, "'https://cdn.olares.cn/joincluster-v1.12.7-20260728.sh'")
}

// --yes must not dial the master, and must leave the password out so the worker
// supplies one; a password without --yes is verified, which needs a reachable
// master and therefore cannot be covered here.
func TestResolveMasterSSHAccessNonInteractive(t *testing.T) {
	detected := func() *common.MasterHostConfig {
		return &common.MasterHostConfig{
			MasterHost:    "192.168.31.16",
			MasterSSHUser: "olares",
			MasterSSHPort: 22,
		}
	}

	t.Run("yes embeds no password", func(t *testing.T) {
		withMasterPassword(t, "")
		cfg := detected()
		assert.NoError(t, resolveMasterSSHAccess(cfg, JoinCommandOptions{AssumeYes: true}))
		assert.Empty(t, cfg.MasterSSHPassword)
		assert.Equal(t, "olares", cfg.MasterSSHUser)
	})

	t.Run("yes with a password embeds it unverified", func(t *testing.T) {
		withMasterPassword(t, "from-flag")
		cfg := detected()
		assert.NoError(t, resolveMasterSSHAccess(cfg, JoinCommandOptions{AssumeYes: true}))
		assert.Equal(t, "from-flag", cfg.MasterSSHPassword)
	})
}

// A payload without a password is valid: the worker is expected to supply one.
func TestMasterAuthInfoAllowsEmptyPassword(t *testing.T) {
	decoded, err := decodeMasterAuthInfo(encodeMasterAuthInfo(&common.MasterHostConfig{
		MasterHost:    "192.168.31.16",
		MasterSSHUser: "olares",
		MasterSSHPort: 22,
	}))
	assert.NoError(t, err)
	assert.Equal(t, "192.168.31.16", decoded.MasterHost)
	assert.Equal(t, "olares", decoded.MasterSSHUser)
	assert.Empty(t, decoded.MasterSSHPassword)
}

func TestShellQuote(t *testing.T) {
	assert.Equal(t, `'node-1'`, shellQuote("node-1"))
	assert.Equal(t, `'a'"'"'b'`, shellQuote("a'b"))
}

func TestValidateMasterHost(t *testing.T) {
	assert.NoError(t, validateMasterHost("192.168.31.16"))
	assert.EqualError(t, validateMasterHost("127.0.0.1"),
		`the detected master address "127.0.0.1" is not a reachable IPv4 address; set OS_LOCALIP to the LAN IPv4 address workers can reach`)
	assert.Error(t, validateMasterHost(""))
	assert.Error(t, validateMasterHost("fe80::1"))
}

func TestValidateCDNService(t *testing.T) {
	assert.NoError(t, validateCDNService("https://cdn.olares.cn"))
	assert.NoError(t, validateCDNService("http://192.168.31.2:8080"))
	assert.NoError(t, validateCDNService(""))
	assert.EqualError(t, validateCDNService("not-a-url"),
		`CDN service "not-a-url" is invalid; use an http:// or https:// URL`)
	assert.Error(t, validateCDNService("ftp://cdn.olares.cn"))
}

func TestMasterAuthInfoRoundTrip(t *testing.T) {
	payload := encodeMasterAuthInfo(&common.MasterHostConfig{
		MasterHost:        "192.168.31.16",
		MasterSSHUser:     "olares",
		MasterSSHPassword: "p@ss:word\nwith symbols",
		MasterSSHPort:     2222,
	})
	decoded, err := decodeMasterAuthInfo(payload)
	assert.NoError(t, err)
	assert.Equal(t, "192.168.31.16", decoded.MasterHost)
	assert.Equal(t, "olares", decoded.MasterSSHUser)
	assert.Equal(t, "p@ss:word\nwith symbols", decoded.MasterSSHPassword)
	assert.Equal(t, 2222, decoded.MasterSSHPort)

	_, err = decodeMasterAuthInfo("not-base64")
	assert.ErrorContains(t, err, "decode MASTER_AUTH_INFO")

	_, err = decodeMasterAuthInfo(encodeMasterAuthInfo(&common.MasterHostConfig{
		MasterSSHUser: "olares",
		MasterSSHPort: 22,
	}))
	assert.ErrorContains(t, err, "missing the master address or SSH user")
}

func withMasterPassword(t *testing.T, password string) {
	t.Helper()
	saved := viper.Get(common.FlagMasterSSHPassword)
	viper.Set(common.FlagMasterSSHPassword, password)
	t.Cleanup(func() { viper.Set(common.FlagMasterSSHPassword, saved) })
}
