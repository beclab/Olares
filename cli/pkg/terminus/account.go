package terminus

import (
	"context"
	"fmt"
	"net/mail"
	"os"
	"path"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/logger"
	corev1 "k8s.io/api/core/v1"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	accounttemplates "bytetrade.io/web3os/installer/pkg/terminus/templates"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"

	ctrl "sigs.k8s.io/controller-runtime"
)

type GetUserInfo struct {
	common.KubeAction
}

func (s *GetUserInfo) Execute(runtime connector.Runtime) error {
	var err error
	if len(s.KubeConf.Arg.User.DomainName) == 0 {
		s.KubeConf.Arg.User.DomainName, err = s.getDomainName()
		if err != nil {
			return err
		}
		logger.Infof("using Domain Name: %s", s.KubeConf.Arg.User.DomainName)
	}
	if len(s.KubeConf.Arg.User.UserName) == 0 {
		s.KubeConf.Arg.User.UserName, err = s.getUserName()
		if err != nil {
			return err
		}
		logger.Infof("using Olares Local Name: %s", s.KubeConf.Arg.User.UserName)
	}
	s.KubeConf.Arg.User.Email, err = s.getUserEmail()
	if err != nil {
		return err
	}
	logger.Infof("using Olares ID: %s", s.KubeConf.Arg.User.Email)
	s.KubeConf.Arg.User.Password, s.KubeConf.Arg.User.EncryptedPassword, err = s.getUserPassword()
	if err != nil {
		return err
	}
	logger.Infof("using password: %s", s.KubeConf.Arg.User.Password)

	return nil
}

func (s *GetUserInfo) getDomainName() (string, error) {
	domainName := strings.TrimSpace(os.Getenv("TERMINUS_OS_DOMAINNAME"))
	if len(domainName) > 0 {
		if !utils.IsValidDomain(domainName) {
			return "", errors.New(fmt.Sprintf("invalid domain name \"%s\" set in env, please reset", domainName))
		}
		return domainName, nil
	}

	reader, err := utils.GetBufIOReaderOfTerminalInput()
	if err != nil {
		return "", errors.Wrap(err, "failed to get terminal input reader")
	}
LOOP:
	fmt.Printf("\nEnter the domain name ( %s by default ): ", cc.DefaultDomainName)
	domainName, err = reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return domainName, errors.Wrap(errors.WithStack(err), "read domain name failed")
	}
	domainName = strings.TrimSpace(domainName)
	if domainName == "" {
		return cc.DefaultDomainName, nil
	}

	if !utils.IsValidDomain(domainName) {
		fmt.Printf("\ninvalid domain name, please try again")
		goto LOOP
	}
	return domainName, nil
}

func (s *GetUserInfo) getUserName() (string, error) {
	userName := os.Getenv("TERMINUS_OS_USERNAME")
	if strings.Contains(userName, "@") {
		userName = strings.Split(userName, "@")[0]
	}
	userName = strings.TrimSpace(userName)
	if len(userName) > 0 {
		if err := utils.ValidateUserName(userName); err != nil {
			return "", fmt.Errorf("invalid username \"%s\" set in env: %s, please reset", userName, err.Error())
		}
		return userName, nil
	}
	reader, err := utils.GetBufIOReaderOfTerminalInput()
	if err != nil {
		return "", errors.Wrap(err, "failed to get terminal input reader")
	}
LOOP:
	fmt.Printf("\nEnter the Olares ID (which you registered in the LarePass app): ")
	userName, err = reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return "", errors.Wrap(errors.WithStack(err), "read username failed")
	}
	if strings.Contains(userName, "@") {
		userName = strings.Split(userName, "@")[0]
	}
	userName = strings.TrimSpace(userName)
	if err := utils.ValidateUserName(userName); err != nil {
		fmt.Printf("\ninvalid username: %s, please try again", err.Error())
		goto LOOP
	}

	return userName, nil
}

func (s *GetUserInfo) getUserEmail() (string, error) {
	userEmail := strings.TrimSpace(os.Getenv("TERMINUS_OS_EMAIL"))
	if len(userEmail) == 0 {
		return s.KubeConf.Arg.User.UserName + "@" + s.KubeConf.Arg.User.DomainName, nil
	}
	if _, err := mail.ParseAddress(userEmail); err != nil {
		return "", fmt.Errorf("invalid email address \"%s\" set in env: %s, please reset", userEmail, err.Error())
	}
	return userEmail, nil
}

func (s *GetUserInfo) getUserPassword() (string, string, error) {
	userPassword := strings.TrimSpace(os.Getenv("TERMINUS_OS_PASSWORD"))
	if len(userPassword) != 32 && len(userPassword) != 0 {
		return "", "", fmt.Errorf("invalid password \"%s\" set in env: length should be equal 32, please reset", userPassword)

	}
	if len(userPassword) == 0 {
		return utils.GenerateEncryptedPassword(8)
	}
	return userPassword, userPassword, nil
}

type SetAccountValues struct {
	common.KubeAction
}

func (p *SetAccountValues) Execute(runtime connector.Runtime) error {
	var accountFile = path.Join(runtime.GetInstallerDir(), "wizard", "config", "account", accounttemplates.AccountValues.Name())

	var data = util.Data{
		"UserName":   p.KubeConf.Arg.User.UserName,
		"Password":   p.KubeConf.Arg.User.EncryptedPassword,
		"Email":      p.KubeConf.Arg.User.Email,
		"DomainName": p.KubeConf.Arg.User.DomainName,
	}

	accountStr, err := util.Render(accounttemplates.AccountValues, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render account template failed")
	}

	if err := util.WriteFile(accountFile, []byte(accountStr), cc.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write account %s failed", accountFile))
	}

	return nil
}

type InstallAccount struct {
	common.KubeAction
}

func (t *InstallAccount) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	ns := corev1.NamespaceDefault
	actionConfig, settings, err := utils.InitConfig(config, ns)
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var accountPath = path.Join(runtime.GetInstallerDir(), "wizard", "config", "account")

	if !util.IsExist(accountPath) {
		return fmt.Errorf("account not exists")
	}

	vals := make(map[string]interface{})
	if si := runtime.GetSystemInfo(); si.GetNATGateway() != "" {
		vals["nat_gateway_ip"] = si.GetNATGateway()
	}

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameAccount, accountPath, "", ns, vals, false); err != nil {
		return err
	}

	return nil
}

type InstallAccountModule struct {
	common.KubeModule
}

func (m *InstallAccountModule) Init() {
	logger.InfoInstallationProgress("Installing account ...")
	m.Name = "InstallAccount"

	getUserInfo := &task.LocalTask{
		Name:   "GetUserInfo",
		Action: new(GetUserInfo),
	}

	setAccountValues := &task.LocalTask{
		Name:   "SetAccountValues",
		Action: &SetAccountValues{},
		Retry:  1,
	}

	installAccount := &task.LocalTask{
		Name:   "InstallAccount",
		Action: &InstallAccount{},
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		getUserInfo,
		setAccountValues,
		installAccount,
	}
}
