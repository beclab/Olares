package nats

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/constants"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/crypto/bcrypt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

const ConfPath = "/dbdata/nats_data/config/nats.conf"
const Allow = "allow"

var (
	defaultPubPerm = []string{"$JS.API.INFO", "$JS.API.STREAM.NAMES", "$JS.API.CONSUMER.CREATE.>",
		"_INBOX.>", "$JS.ACK.>", "$SYS.ACCOUNT.*.CONNECT", "$SYS.ACCOUNT.*.DISCONNECT", "$JS.FC.>",
		"$SYS._INBOX_.>", "$SYS.SERVER.*.CLIENT.AUTH.ERR", "$SYS.REQ.SERVER.PING.>"}
	defaultSubPerm = []string{"$JS.API.STREAM.NAMES", "$JS.API.CONSUMER.CREATE.>", "_INBOX.>",
		"$SYS.ACCOUNT.*.CONNECT", "$SYS.ACCOUNT.*.DISCONNECT", "$JS.FC.>", "$SYS._INBOX_.>",
		"$SYS.SERVER.*.CLIENT.AUTH.ERR", "$SYS.REQ.SERVER.PING.>"}
)

func createOrUpdateUser(request *aprv1.MiddlewareRequest, namespace, password string, loadConfig func() (*Config, error)) (*Config, error) {
	encryptedPassword, err := encryptPassword(password)

	if err != nil {
		return nil, err
	}
	allowPubSubject, allowSubSubject, err := getAllowPubSubSubjectFromMR(request, namespace)
	if err != nil {
		klog.Infof("getAllowPubSubSubjectFromMR, err=%v", err)
		return nil, err
	}
	req := request.Spec.Nats
	user := User{
		Username: req.User,
		Password: encryptedPassword,
		Permissions: Permissions{
			Publish: Publish{
				Allow: allowPubSubject,
			},
			Subscribe: Subscribe{
				Allow: allowSubSubject,
			},
		},
	}
	config, err := loadConfig()
	if err != nil {
		klog.Infof("loadconfig err=%v", err)
		return nil, err
	}
	klog.Infof("nats Config: %#v", config)
	isUserExists := false
	for i, c := range config.Accounts.Terminus.Users {
		if c.Username == req.User {
			config.Accounts.Terminus.Users[i] = user
			isUserExists = true
		}
	}
	if !isUserExists {
		config.Accounts.Terminus.Users = append(config.Accounts.Terminus.Users, user)
	}
	return config, nil
}
func CreateOrUpdateUser(request *aprv1.MiddlewareRequest, namespace, password string) (*Config, error) {
	config, err := createOrUpdateUser(request, namespace, password, loadConf)
	if err != nil {
		klog.Infof("createOrUpdateUser err=%v", err)
		return nil, err
	}
	err = RenderConfigFile(config)
	if err != nil {
		klog.Infof("renderConfigFile err=%v", err)
		return nil, err
	}
	return config, nil
}

func getAllowPubSubSubjectFromMR(request *aprv1.MiddlewareRequest, namespace string) ([]string, []string, error) {
	req := request.Spec.Nats.DeepCopy()
	for i, s := range req.Subjects {
		req.Subjects[i].Name = MakeRealSubjectName(s.Name, request.Spec.AppNamespace)

	}
	for i, ref := range req.Refs {
		for j, s := range ref.Subjects {
			req.Refs[i].Subjects[j].Name = MakeRealNameForRefSubjectName(ref.AppNamespace, ref.AppName, s.Name, GetOwnerNameFromNs(request.Namespace))
		}
	}

	allowPubSubject := make([]string, 0)
	allowSubSubject := make([]string, 0)
	for _, subject := range req.Subjects {
		if subject.Permission.Pub == Allow {
			allowPubSubject = append(allowPubSubject, subject.Name)
		}
		if subject.Permission.Sub == Allow {
			allowSubSubject = append(allowSubSubject, subject.Name)
		}
	}
	for _, subject := range req.Refs {
		for _, s := range subject.Subjects {
			if s.Pub == Allow {
				allowPubSubject = append(allowPubSubject, s.Name)
			}
			if s.Sub == Allow {
				allowSubSubject = append(allowSubSubject, s.Name)
			}
		}
	}

	klog.Infof("req.Nats: %#v", req)

	if len(allowPubSubject) > 0 {
		allowPubSubject = append(allowPubSubject, defaultPubPerm...)
	}
	if len(allowSubSubject) > 0 {
		allowSubSubject = append(allowSubSubject, defaultSubPerm...)
	}

	return allowPubSubject, allowSubSubject, nil
}

func CreateOrUpdateStream(appNamespace, app string) error {
	//name := fmt.Sprintf("%s-%s", appNamespace, app)
	adminPassword, err := getAdminPassword()
	if err != nil {
		return err
	}
	nc, err := nats.Connect("nats://nats."+constants.PlatformNamespace, nats.UserInfo("admin", adminPassword))
	if err != nil {
		return err
	}
	defer nc.Drain()
	js, err := jetstream.New(nc)
	if err != nil {
		return err
	}
	cfg := jetstream.StreamConfig{
		Name:     "os-stream",
		Subjects: []string{"os.>"},
		Storage:  jetstream.FileStorage,
		MaxAge:   24 * 60 * 60 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = js.CreateStream(ctx, cfg)
	if err != nil && !errors.Is(err, jetstream.ErrStreamNameAlreadyInUse) {
		klog.Errorf("create os-stream failed %v", err)
		return err
	}
	return nil
}

func DeleteStream(appNamespace, app string) error {
	name := fmt.Sprintf("%s-%s", appNamespace, app)
	adminPassword, err := getAdminPassword()
	if err != nil {
		return err
	}
	nc, err := nats.Connect("nats://nats."+constants.PlatformNamespace, nats.UserInfo("admin", adminPassword))
	if err != nil {
		return err
	}
	defer nc.Drain()
	js, err := jetstream.New(nc)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = js.DeleteStream(ctx, name)
	if err != nil && errors.Is(err, nats.ErrStreamNotFound) {
		return err
	}
	return nil
}

func DeleteUser(username string) error {
	config, err := loadConf()
	if err != nil {
		return err
	}
	for i, u := range config.Accounts.Terminus.Users {
		if u.Username == username {
			config.Accounts.Terminus.Users = append(config.Accounts.Terminus.Users[:i],
				config.Accounts.Terminus.Users[i+1:]...)
		}
	}
	err = RenderConfigFile(config)
	if err != nil {
		return err
	}
	return nil
}

func encryptPassword(password string) (string, error) {
	encryptedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(encryptedPass), nil
}

func loadConf() (*Config, error) {
	password, err := getAdminPassword()
	if err != nil {
		return nil, err
	}
	err = os.Setenv("ADMIN_PASSWORD", password)
	if err != nil {
		klog.Infof("set env error=%v", err)
		return nil, err
	}
	return ParseFile(ConfPath)
}

var ch = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@$#%^&*()")

func sizedBytes(sz int) []byte {
	b := make([]byte, sz)
	for i := range b {
		b[i] = ch[rand.Intn(len(ch))]
	}
	return b
}

func sizedString(sz int) string {
	return string(sizedBytes(sz))
}

var re = regexp.MustCompile(`^(?:[^.]*\.){3}(.*)$`)

func GetOriginSubjectName(subjectName string) string {
	match := re.FindStringSubmatch(subjectName)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func newClientSet() (*kubernetes.Clientset, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		klog.Infof("get config err=%v", err)
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Infof("create clientset err=%v", clientSet)
		return nil, err
	}
	return clientSet, nil
}

func getAdminPassword() (string, error) {
	clientSet, err := newClientSet()
	if err != nil {
		return "", err
	}
	secret, err := clientSet.CoreV1().Secrets(constants.PlatformNamespace).Get(context.TODO(), "nats-secrets", metav1.GetOptions{})
	if err != nil {
		klog.Infof("get nats-secrets err=%v", err)
		return "", err
	}
	password, ok := secret.Data["nats_password"]
	if !ok {
		klog.Infof("empty nats-Password")
		return "", err
	}

	return string(password), nil
}

func MakeRealSubjectName(subject string, appNamespace string) string {
	return fmt.Sprintf("%s.%s", appNamespace, subject)
}

func MakeRealNameForRefSubjectName(refNamespace, app, subject, ownerName string) string {
	refAppNs := ""
	if strings.HasPrefix(refNamespace, "user-space") {
		refAppNs = fmt.Sprintf("user-space-%s", ownerName)
	} else if strings.HasPrefix(refNamespace, "user-system") {
		refAppNs = fmt.Sprintf("user-system-%s", ownerName)
	} else {
		refAppNs = refNamespace
	}
	return fmt.Sprintf("%s.%s", refAppNs, subject)
}

func GetOwnerNameFromNs(ns string) string {
	nsSplict := strings.Split(ns, "-")
	return nsSplict[len(nsSplict)-1]
}

func FindNatsAdminUser(ctx context.Context, k8sClient *kubernetes.Clientset) (user, password string, err error) {
	secret, err := k8sClient.CoreV1().Secrets(constants.PlatformNamespace).Get(ctx, "nats-secrets", metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	return "admin", string(secret.Data["nats_password"]), nil
}
