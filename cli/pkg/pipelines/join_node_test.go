package pipelines

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestJoinHostnameProblem(t *testing.T) {
	assert.Empty(t, joinHostnameProblem("intel-a1b2c3", []string{"olares", "amd-d4e5f6"}))
	assert.Equal(t,
		`node name "intel-a1b2c3" already exists in the master cluster`,
		joinHostnameProblem("intel-a1b2c3", []string{"INTEL-A1B2C3"}),
	)
	assert.NotEmpty(t, joinHostnameProblem("not_valid", nil))
	assert.NotEmpty(t, joinHostnameProblem("", nil))
}

// An operator pointing this node at a cluster always outranks what the node
// remembers, but a bare rerun should resume rather than ask everything again.
func TestResolveMasterConnectionSources(t *testing.T) {
	payload := encodeMasterAuthInfo(&common.MasterHostConfig{
		MasterHost:        "192.168.31.16",
		MasterSSHUser:     "olares",
		MasterSSHPassword: "from-payload",
		MasterSSHPort:     2222,
	})

	t.Run("payload provides the whole connection", func(t *testing.T) {
		withJoinViper(t, map[string]any{common.FlagMasterAuthInfo: payload})
		arg := argWithRememberedMaster()
		assert.NoError(t, resolveMasterConnectionSources(arg))
		assert.Equal(t, "192.168.31.16", arg.MasterHost)
		assert.Equal(t, "olares", arg.MasterSSHUser)
		assert.Equal(t, "from-payload", arg.MasterSSHPassword)
		assert.Equal(t, 2222, arg.MasterSSHPort)
	})

	t.Run("explicit flags refine the payload", func(t *testing.T) {
		withJoinViper(t, map[string]any{
			common.FlagMasterAuthInfo:    payload,
			common.FlagMasterSSHUser:     "admin",
			common.FlagMasterSSHPassword: "from-flag",
		})
		arg := argWithRememberedMaster()
		assert.NoError(t, resolveMasterConnectionSources(arg))
		assert.Equal(t, "192.168.31.16", arg.MasterHost)
		assert.Equal(t, "admin", arg.MasterSSHUser)
		assert.Equal(t, "from-flag", arg.MasterSSHPassword)
	})

	t.Run("a payload for another master replaces what is remembered", func(t *testing.T) {
		withJoinViper(t, map[string]any{common.FlagMasterAuthInfo: payload})
		arg := argWithRememberedMaster()
		assert.NoError(t, resolveMasterConnectionSources(arg))
		assert.Equal(t, "192.168.31.16", arg.MasterHost)
		assert.Equal(t, "from-payload", arg.MasterSSHPassword)
		assert.NotEqual(t, "old-secret", arg.MasterSSHPassword)
	})

	t.Run("an explicitly named other master replaces what is remembered", func(t *testing.T) {
		withJoinViper(t, map[string]any{common.FlagMasterHost: "192.168.31.50"})
		arg := argWithRememberedMaster()
		assert.NoError(t, resolveMasterConnectionSources(arg))
		assert.Equal(t, "192.168.31.50", arg.MasterHost)
		assert.Empty(t, arg.MasterSSHPassword, "credentials of the previous master must not be reused")
	})

	t.Run("with nothing supplied, the remembered master resumes the join", func(t *testing.T) {
		withJoinViper(t, nil)
		arg := argWithRememberedMaster()
		assert.NoError(t, resolveMasterConnectionSources(arg))
		assert.Equal(t, "10.0.0.9", arg.MasterHost)
		assert.Equal(t, "old-admin", arg.MasterSSHUser)
		assert.Equal(t, "old-secret", arg.MasterSSHPassword)
	})

	t.Run("defaults fill in a bare connection", func(t *testing.T) {
		withJoinViper(t, map[string]any{common.FlagMasterHost: "192.168.31.50"})
		arg := &common.Argument{MasterHostConfig: &common.MasterHostConfig{}}
		assert.NoError(t, resolveMasterConnectionSources(arg))
		assert.Equal(t, "root", arg.MasterSSHUser)
		assert.Equal(t, 22, arg.MasterSSHPort)
	})

	t.Run("a malformed payload is reported", func(t *testing.T) {
		withJoinViper(t, map[string]any{common.FlagMasterAuthInfo: "not-base64"})
		assert.ErrorContains(t,
			resolveMasterConnectionSources(argWithRememberedMaster()),
			"decode MASTER_AUTH_INFO")
	})

	t.Run("an out-of-range port is reported", func(t *testing.T) {
		withJoinViper(t, map[string]any{
			common.FlagMasterHost:    "192.168.31.16",
			common.FlagMasterSSHPort: 70000,
		})
		assert.ErrorContains(t,
			resolveMasterConnectionSources(argWithRememberedMaster()),
			"between 1 and 65535")
	})
}

// argWithRememberedMaster mimics NewArgument() on a node that already has a
// master.conf: its contents have been loaded into the Argument.
func argWithRememberedMaster() *common.Argument {
	return &common.Argument{MasterHostConfig: &common.MasterHostConfig{
		MasterHost:        "10.0.0.9",
		MasterSSHUser:     "old-admin",
		MasterSSHPassword: "old-secret",
		MasterSSHPort:     22,
	}}
}

// withJoinViper isolates the global viper keys these subtests depend on.
func withJoinViper(t *testing.T, values map[string]any) {
	t.Helper()
	keys := []string{
		common.FlagMasterAuthInfo,
		common.FlagMasterHost,
		common.FlagMasterSSHUser,
		common.FlagMasterSSHPassword,
		common.FlagMasterSSHPrivateKeyPath,
		common.FlagMasterSSHPort,
	}
	saved := make(map[string]any, len(keys))
	for _, key := range keys {
		saved[key] = viper.Get(key)
		viper.Set(key, nil)
	}
	for key, value := range values {
		viper.Set(key, value)
	}
	t.Cleanup(func() {
		for _, key := range keys {
			viper.Set(key, saved[key])
		}
	})
}
