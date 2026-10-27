package pipelines

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/utils"
	"golang.org/x/term"
)

const (
	// maxCredentialAttempts bounds the interactive credential prompts. Without a
	// limit, a master that can never accept these credentials (sshd with
	// password authentication disabled being the common case) would loop
	// forever with no hint about the real cause.
	maxCredentialAttempts = 3

	masterSSHTimeout = 10 * time.Second

	// masterAuthInfoVersion prefixes the MASTER_AUTH_INFO payload so its layout
	// can change later without an older command being silently mis-parsed.
	masterAuthInfoVersion = "v1"
)

// encodeMasterAuthInfo packs the master connection details into the single
// opaque token carried by the generated worker bootstrap command. Every field is
// Base64-encoded on its own, so a password may contain any character, including
// the ':' separator and newlines.
//
// This is encoding, not encryption: the token is exactly as secret as the
// command that carries it.
func encodeMasterAuthInfo(cfg *common.MasterHostConfig) string {
	encodeField := func(value string) string {
		return base64.StdEncoding.EncodeToString([]byte(value))
	}
	inner := strings.Join([]string{
		masterAuthInfoVersion,
		encodeField(cfg.MasterHost),
		encodeField(cfg.MasterSSHUser),
		encodeField(cfg.MasterSSHPassword),
		strconv.Itoa(cfg.MasterSSHPort),
	}, ":")
	return base64.StdEncoding.EncodeToString([]byte(inner))
}

func decodeMasterAuthInfo(payload string) (*common.MasterHostConfig, error) {
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", common.ENV_MASTER_AUTH_INFO, err)
	}
	parts := strings.Split(string(decoded), ":")
	if len(parts) != 5 || parts[0] != masterAuthInfoVersion {
		return nil, fmt.Errorf("%s has an unsupported format; regenerate it with 'olares-cli node join-command' on the master",
			common.ENV_MASTER_AUTH_INFO)
	}
	decodeField := func(name, value string) (string, error) {
		field, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return "", fmt.Errorf("decode %s %s: %w", common.ENV_MASTER_AUTH_INFO, name, err)
		}
		return string(field), nil
	}

	cfg := &common.MasterHostConfig{}
	if cfg.MasterHost, err = decodeField("host", parts[1]); err != nil {
		return nil, err
	}
	if cfg.MasterSSHUser, err = decodeField("username", parts[2]); err != nil {
		return nil, err
	}
	if cfg.MasterSSHPassword, err = decodeField("password", parts[3]); err != nil {
		return nil, err
	}
	cfg.MasterSSHPort, err = strconv.Atoi(parts[4])
	if err != nil || cfg.MasterSSHPort < 1 || cfg.MasterSSHPort > 65535 {
		return nil, fmt.Errorf("%s contains an invalid SSH port", common.ENV_MASTER_AUTH_INFO)
	}
	// The password may legitimately be empty: a command generated with --yes
	// deliberately leaves it out, and the worker then takes one from the master
	// SSH environment variables, from a key, or by asking.
	if cfg.MasterHost == "" || cfg.MasterSSHUser == "" {
		return nil, fmt.Errorf("%s is missing the master address or SSH user", common.ENV_MASTER_AUTH_INFO)
	}
	return cfg, nil
}

// hasMasterCredential reports whether the config carries any usable SSH
// authentication material. Without one, connecting is pointless: the SSH client
// rejects the attempt before reaching the master.
func hasMasterCredential(cfg *common.MasterHostConfig) bool {
	return cfg.MasterSSHPassword != "" || cfg.MasterSSHPrivateKeyPath != ""
}

// verifyMasterSSH checks that the credentials can both log in to the master and
// run a command through sudo there, which is what every later step requires.
func verifyMasterSSH(cfg *common.MasterHostConfig) error {
	host := connector.NewHost()
	host.SetAddress(cfg.MasterHost)
	host.SetInternalAddress(cfg.MasterHost)
	host.SetPort(cfg.MasterSSHPort)
	host.SetUser(cfg.MasterSSHUser)
	host.SetPassword(cfg.MasterSSHPassword)
	host.SetPrivateKeyPath(cfg.MasterSSHPrivateKeyPath)

	conn, err := connector.NewConnection(connector.Cfg{
		Username: cfg.MasterSSHUser,
		Password: cfg.MasterSSHPassword,
		KeyFile:  cfg.MasterSSHPrivateKeyPath,
		Address:  cfg.MasterHost,
		Port:     cfg.MasterSSHPort,
		Timeout:  masterSSHTimeout,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	_, code, err := conn.Exec(host.SudoPrefixIfNecessary("true"), host)
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("cannot run sudo on the master as %q (exit status %d)", cfg.MasterSSHUser, code)
	}
	return nil
}

// defaultPrivateKeyPath returns the conventional SSH key location when it
// exists, so users who already authenticate to the master by key are not asked
// for a password they may not even have.
func defaultPrivateKeyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(home, ".ssh", "id_rsa")
	if !utils.IsExist(path) {
		return ""
	}
	return path
}

// promptMasterCredentials asks for an SSH user and password for the master and
// keeps the first pair that is verified to work, giving up after
// maxCredentialAttempts so an sshd that refuses password authentication produces
// an explanation rather than an endless loop.
//
// The username is only asked for once it is actually in question: a known user
// that simply had the wrong password is offered back as the default.
func promptMasterCredentials(cfg *common.MasterHostConfig) error {
	for attempt := 1; attempt <= maxCredentialAttempts; attempt++ {
		if cfg.MasterSSHUser == "" || attempt > 1 {
			user, err := promptNonEmptyLine(
				fmt.Sprintf("SSH username on the master [%s]: ", orDefault(cfg.MasterSSHUser, "root")),
				orDefault(cfg.MasterSSHUser, "root"))
			if err != nil {
				return err
			}
			cfg.MasterSSHUser = user
		}
		password, err := promptHiddenPassword(
			fmt.Sprintf("SSH password for %s@%s: ", cfg.MasterSSHUser, cfg.MasterHost))
		if err != nil {
			return err
		}
		if password == "" {
			fmt.Fprintln(os.Stderr, "The SSH password cannot be empty.")
			continue
		}

		candidate := *cfg
		candidate.MasterSSHPassword = password
		// A password was entered deliberately, so a key that happens to exist
		// must not decide the outcome of this attempt.
		candidate.MasterSSHPrivateKeyPath = ""
		if err := verifyMasterSSH(&candidate); err != nil {
			fmt.Fprintf(os.Stderr, "Could not verify SSH login and sudo access: %v\n", err)
			continue
		}

		cfg.MasterSSHPassword = password
		cfg.MasterSSHPrivateKeyPath = ""
		return nil
	}
	return fmt.Errorf(
		"could not verify SSH access to %s:%d after %d attempts\n"+
			"check that the master's SSH service accepts this user and password;\n"+
			"password authentication must be enabled in its sshd configuration",
		cfg.MasterHost, cfg.MasterSSHPort, maxCredentialAttempts)
}

func orDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func promptConfirmation(label string, defaultYes bool) (bool, error) {
	for {
		answer, err := promptLine(label)
		if err != nil {
			return false, err
		}
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "":
			return defaultYes, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(os.Stderr, "Please answer yes or no.")
		}
	}
}

func promptNonEmptyLine(label, defaultValue string) (string, error) {
	for {
		value, err := promptLine(label)
		if err != nil {
			return "", err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			value = defaultValue
		}
		if value != "" {
			return value, nil
		}
		fmt.Fprintln(os.Stderr, "A value is required.")
	}
}

func promptLine(label string) (string, error) {
	tty, reader, err := terminalSession()
	if err != nil {
		return "", err
	}
	if _, err := fmt.Fprint(tty, label); err != nil {
		return "", err
	}
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read terminal input: %w", err)
	}
	return strings.TrimRight(value, "\r\n"), nil
}

func promptHiddenPassword(label string) (string, error) {
	tty, _, err := terminalSession()
	if err != nil {
		return "", err
	}
	if _, err := fmt.Fprint(tty, label); err != nil {
		return "", err
	}
	password, err := term.ReadPassword(int(tty.Fd()))
	fmt.Fprintln(tty)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(password), nil
}

// The terminal is opened once and its reader kept for the whole process. A
// per-prompt reader would drop anything already buffered past the newline, so
// input typed or pasted ahead of a later prompt would silently vanish.
var (
	ttyOnce   sync.Once
	ttyFile   *os.File
	ttyReader *bufio.Reader
	ttyErr    error
)

// terminalSession returns the controlling terminal to prompt on. The bootstrap
// script is piped into bash, so stdin carries the script rather than the user;
// /dev/tty is the only reliable way to reach them.
func terminalSession() (*os.File, *bufio.Reader, error) {
	ttyOnce.Do(func() {
		ttyFile, ttyErr = os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if ttyErr != nil {
			ttyErr = fmt.Errorf(
				"this step needs an interactive terminal, but /dev/tty is not available: %w\n"+
					"run the command from a terminal, or supply every value up front through flags or environment variables",
				ttyErr)
			return
		}
		ttyReader = bufio.NewReader(ttyFile)
	})
	return ttyFile, ttyReader, ttyErr
}
