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

package connector

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Cfg struct {
	Username    string
	Password    string
	Address     string
	Port        int
	PrivateKey  string
	KeyFile     string
	AgentSocket string
	Timeout     time.Duration
	Bastion     string
	BastionPort int
	BastionUser string
}

const socketEnvPrefix = "env:"

type connection struct {
	mu         sync.Mutex
	sftpclient *sftp.Client
	sshclient  *ssh.Client
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewConnection(cfg Cfg) (Connection, error) {
	cfg, err := validateOptions(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to validate ssh connection parameters")
	}

	authMethods := make([]ssh.AuthMethod, 0)

	if len(cfg.Password) > 0 {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}

	if len(cfg.PrivateKey) > 0 {
		signer, parseErr := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
		if parseErr != nil {
			return nil, errors.Wrap(parseErr, "The given SSH key could not be parsed")
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(cfg.AgentSocket) > 0 {
		addr := cfg.AgentSocket

		if strings.HasPrefix(cfg.AgentSocket, socketEnvPrefix) {
			envName := strings.TrimPrefix(cfg.AgentSocket, socketEnvPrefix)

			if envAddr := os.Getenv(envName); len(envAddr) > 0 {
				addr = envAddr
			}
		}

		socket, dialErr := net.Dial("unix", addr)
		if dialErr != nil {
			return nil, errors.Wrapf(dialErr, "could not open socket %q", addr)
		}

		agentClient := agent.NewClient(socket)

		signers, signersErr := agentClient.Signers()
		if signersErr != nil {
			_ = socket.Close()
			return nil, errors.Wrap(signersErr, "error when creating signer for SSH agent")
		}

		authMethods = append(authMethods, ssh.PublicKeys(signers...))
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.Username,
		Timeout:         cfg.Timeout,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	targetHost := cfg.Address
	targetPort := strconv.Itoa(cfg.Port)

	if cfg.Bastion != "" {
		targetHost = cfg.Bastion
		targetPort = strconv.Itoa(cfg.BastionPort)
		sshConfig.User = cfg.BastionUser
	}

	endpoint := net.JoinHostPort(targetHost, targetPort)

	client, err := ssh.Dial("tcp", endpoint, sshConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not establish connection to %s", endpoint)
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	sshConn := &connection{
		ctx:    ctx,
		cancel: cancelFn,
	}

	if cfg.Bastion == "" {
		sshConn.sshclient = client
		sftpClient, err := sftp.NewClient(sshConn.sshclient)
		if err != nil {
			return nil, errors.Wrapf(err, "new sftp client failed: %v", err)
		}
		sshConn.sftpclient = sftpClient
		return sshConn, nil
	}

	endpointBehindBastion := net.JoinHostPort(cfg.Address, strconv.Itoa(cfg.Port))

	conn, err := client.Dial("tcp", endpointBehindBastion)
	if err != nil {
		return nil, errors.Wrapf(err, "could not establish connection to %s", endpointBehindBastion)
	}

	sshConfig.User = cfg.Username
	ncc, chans, reqs, err := ssh.NewClientConn(conn, endpointBehindBastion, sshConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not establish connection to %s", endpointBehindBastion)
	}

	sshConn.sshclient = ssh.NewClient(ncc, chans, reqs)
	sftpClient, err := sftp.NewClient(sshConn.sshclient)
	if err != nil {
		return nil, errors.Wrapf(err, "new sftp client failed: %v", err)
	}
	sshConn.sftpclient = sftpClient
	return sshConn, nil
}

func validateOptions(cfg Cfg) (Cfg, error) {
	if len(cfg.Username) == 0 {
		return cfg, errors.New("No username specified for SSH connection")
	}

	if len(cfg.Address) == 0 {
		return cfg, errors.New("No address specified for SSH connection")
	}

	if len(cfg.Password) == 0 && len(cfg.PrivateKey) == 0 && len(cfg.KeyFile) == 0 && len(cfg.AgentSocket) == 0 {
		return cfg, errors.New("Must specify at least one of password, private key, keyfile or agent socket")
	}

	if len(cfg.PrivateKey) == 0 && len(cfg.KeyFile) > 0 {
		content, err := ioutil.ReadFile(cfg.KeyFile)
		if err != nil {
			return cfg, errors.Wrapf(err, "Failed to read keyfile %q", cfg.KeyFile)
		}

		cfg.PrivateKey = string(content)
		cfg.KeyFile = ""
	}

	if cfg.Port <= 0 {
		cfg.Port = 22
	}

	if cfg.BastionPort <= 0 {
		cfg.BastionPort = 22
	}

	if cfg.BastionUser == "" {
		cfg.BastionUser = cfg.Username
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 15 * time.Second
	}

	return cfg, nil
}

func (c *connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil && c.sftpclient == nil {
		return
	}
	c.cancel()

	if c.sshclient != nil {
		c.sshclient.Close()
		c.sshclient = nil
	}
	if c.sftpclient != nil {
		c.sftpclient.Close()
		c.sftpclient = nil
	}
}

func (c *connection) session() (*ssh.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil {
		return nil, errors.New("connection closed")
	}

	sess, err := c.sshclient.NewSession()
	if err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = sess.RequestPty("xterm", 100, 50, modes)
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (c *connection) PExec(cmd string, stdin io.Reader, stdout io.Writer, stderr io.Writer, host Host) (int, error) {
	sess, err := c.session()
	if err != nil {
		return 1, errors.Wrap(err, "failed to get SSH session")
	}
	defer sess.Close()

	sess.Stdin = stdin
	sess.Stdout = stdout
	sess.Stderr = stderr

	exitCode := 0

	in, _ := sess.StdinPipe()
	out, _ := sess.StdoutPipe()

	err = sess.Start(strings.TrimSpace(cmd))
	if err != nil {
		exitCode = -1
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		}
	}

	var (
		output []byte
		line   = ""
		r      = bufio.NewReader(out)
	)

	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}

		output = append(output, b)

		if b == byte('\n') {
			line = ""
			continue
		}

		line += string(b)

		if (strings.HasPrefix(line, "[sudo] password for ") || strings.HasPrefix(line, "Password")) && strings.HasSuffix(line, ": ") {
			_, err = in.Write([]byte(host.GetPassword() + "\n"))
			if err != nil {
				break
			}
		}
	}
	err = sess.Wait()
	if err != nil {
		exitCode = -1
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		}
	}

	// preserve original error
	return exitCode, err
}

func (c *connection) Exec(cmd string, host Host) (stdout string, code int, err error) {
	sess, err := c.session()
	if err != nil {
		return "", 1, errors.Wrap(err, "failed to get SSH session")
	}
	defer sess.Close()

	exitCode := 0

	in, _ := sess.StdinPipe()
	out, _ := sess.StdoutPipe()

	err = sess.Start(strings.TrimSpace(cmd))
	if err != nil {
		exitCode = -1
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		}
		return err.Error(), exitCode, err
	}

	var (
		output     []byte
		line       = ""
		r          = bufio.NewReader(out)
		trimPrefix = ""
	)

	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}

		output = append(output, b)

		if b == byte('\n') {
			line = ""
			continue
		}

		line += string(b)

		// we explicitly specify a custom sudo prompt to avoid unexpected prompts at best effort
		// which should fit in almost all the cases in our program,
		// but other prompts are also possible
		// e.g., because of a hardcoded "sudo" command string in the program
		// or third-party scripts executed by us that contain the "sudo" command
		// the Chinese colon suffix are checked to fit a machine with Chinese locale
		// in very rare cases, if the native sudo prompt does not start with "[sudo]" and end with a colon,
		// this match can still miss
		if strings.HasPrefix(line, "[sudo]") && (strings.HasSuffix(line, ": ") || strings.HasSuffix(line, "： ")) {
			trimPrefix = line
			_, err = in.Write([]byte(host.GetPassword() + "\n"))
			if err != nil {
				break
			}
		}
	}
	err = sess.Wait()
	if err != nil {
		exitCode = -1
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		}
	}
	outStr := strings.TrimPrefix(string(output), trimPrefix)

	// preserve original error
	return strings.TrimSpace(outStr), exitCode, errors.Wrapf(err, "Failed to exec command: %s \n%s", cmd, strings.TrimSpace(outStr))
}

func (c *connection) Fetch(local, remote string, host Host) error {
	//srcFile, err := c.sftpclient.Open(remote)
	//if err != nil {
	//	return fmt.Errorf("open remote file failed %v, remote path: %s", err, remote)
	//}
	//defer srcFile.Close()

	// Base64 encoding is performed on the contents of the file to prevent garbled code in the target file.
	output, _, err := c.Exec(host.SudoPrefixIfNecessary(fmt.Sprintf("cat %s | base64 -w 0", remote)), host)
	if err != nil {
		return fmt.Errorf("open remote file failed %v, remote path: %s", err, remote)
	}

	err = util.MkFileFullPathDir(local)
	if err != nil {
		return err
	}
	// open local Destination file
	dstFile, err := os.Create(local)
	if err != nil {
		return fmt.Errorf("create local file failed %v", err)
	}
	defer dstFile.Close()
	// copy to local file
	//_, err = srcFile.WriteTo(dstFile)
	if base64Str, err := base64.StdEncoding.DecodeString(output); err != nil {
		return err
	} else {
		if _, err = dstFile.WriteString(string(base64Str)); err != nil {
			return err
		}
	}

	return nil
}

type scpErr struct {
	err error
}

func (c *connection) Scp(src, dst string, host Host) error {
	baseRemotePath := filepath.Dir(dst)

	if err := c.MkDirAll(baseRemotePath, "777", host); err != nil {
		return err
	}
	f, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("get file stat failed: %s", err)
	}

	number := 1
	if f.IsDir() {
		number = util.CountDirFiles(src)
	}
	// empty dir
	if number == 0 {
		return nil
	}

	scpErr := new(scpErr)
	if f.IsDir() {
		c.copyDirToRemote(src, dst, scpErr, host)
		if scpErr.err != nil {
			return scpErr.err
		}
	} else {
		if err := c.copyFileToRemote(src, dst, host); err != nil {
			return err
		}
	}
	return nil
}

func (c *connection) copyDirToRemote(src, dst string, scrErr *scpErr, host Host) {
	localFiles, err := ioutil.ReadDir(src)
	if err != nil {
		logger.Errorf("read local path dir %s failed %v", src, err)
		scrErr.err = err
		return
	}
	if err = c.MkDirAll(dst, "", host); err != nil {
		logger.Errorf("failed to create remote path %s:%v", dst, err)
		scrErr.err = err
		return
	}
	for _, file := range localFiles {
		local := path.Join(src, file.Name())
		remote := path.Join(dst, file.Name())
		if file.IsDir() {
			if err = c.MkDirAll(remote, "", host); err != nil {
				logger.Errorf("failed to create remote path %s:%v", remote, err)
				scrErr.err = err
				return
			}
			c.copyDirToRemote(local, remote, scrErr, host)
		} else {
			err := c.copyFileToRemote(local, remote, host)
			if err != nil {
				logger.Errorf("copy local file %s to remote file %s failed %v ", local, remote, err)
				scrErr.err = err
				return
			}
		}
	}
}

func (c *connection) copyFileToRemote(src, dst string, host Host) error {
	// check remote file md5 first
	var (
		srcMd5, dstMd5 string
	)
	srcMd5 = util.LocalMd5Sum(src)
	if c.RemoteFileExist(dst, host) {
		dstMd5 = c.RemoteMd5Sum(dst, host)
		if srcMd5 == dstMd5 {
			logger.Debug("remote file %s md5 value is the same as local file, skip scp", dst)
			return nil
		}
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	// the dst file mod will be 0666
	dstFile, err := c.sftpclient.Create(dst)
	if err != nil {
		return err
	}
	fileStat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("get file stat failed %v", err)
	}
	if err := dstFile.Chmod(fileStat.Mode()); err != nil {
		return fmt.Errorf("chmod remote file failed %v", err)
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	dstMd5 = c.RemoteMd5Sum(dst, host)
	if srcMd5 != dstMd5 {
		return fmt.Errorf("validate md5sum failed %s != %s", srcMd5, dstMd5)
	}
	return nil
}

func (c *connection) RemoteMd5Sum(dst string, host Host) string {
	cmd := fmt.Sprintf("md5sum %s | cut -d\" \" -f1", dst)
	remoteMd5, _, err := c.Exec(cmd, host)
	if err != nil {
		logger.Errorf("exec countRemoteMd5Command %s failed: %v", cmd, err)
	}
	return remoteMd5
}

func (c *connection) RemoteFileExist(dst string, host Host) bool {
	remoteFileName := path.Base(dst)
	remoteFileDirName := path.Dir(dst)

	remoteFileCommand := fmt.Sprintf(host.SudoPrefixIfNecessary("ls -l %s/%s 2>/dev/null |wc -l"), remoteFileDirName, remoteFileName)

	out, _, err := c.Exec(remoteFileCommand, host)
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("exec remoteFileCommand %s err: %v", remoteFileCommand, err)
		}
	}()
	if err != nil {
		panic(1)
	}
	count, err := strconv.Atoi(strings.TrimSpace(out))
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("check remote file exist err: %v", err)
		}
	}()
	if err != nil {
		panic(1)
	}
	return count != 0
}

func (c *connection) RemoteDirExist(dst string, host Host) (bool, error) {
	if _, err := c.sftpclient.ReadDir(dst); err != nil {
		return false, err
	}
	return true, nil
}

func (c *connection) MkDirAll(path string, mode string, host Host) error {
	if mode == "" {
		mode = "775"
	}
	mkDstDir := fmt.Sprintf("mkdir -p -m %s %s || true", mode, path)
	if _, _, err := c.Exec(host.SudoPrefixIfNecessary(mkDstDir), host); err != nil {
		return err
	}

	return nil
}

func (c *connection) Chmod(path string, mode os.FileMode) error {
	remotePath := filepath.Dir(path)
	if err := c.sftpclient.Chmod(remotePath, mode); err != nil {
		return err
	}
	return nil
}

func SudoPrefix(cmd string) string {
	cmd = strings.ReplaceAll(cmd, `"`, `\"`)
	return fmt.Sprintf("sudo -p \"[sudo] Passowrd: \" -E /bin/bash -c \"%s\"", cmd)
}
