package infisical

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bytetrade.io/web3os/tapr/pkg/constants"
	infisical_crypto "bytetrade.io/web3os/tapr/pkg/vault/infisical/crypto"
	"golang.org/x/crypto/nacl/box"
	"k8s.io/klog/v2"
)

var (
	Owner              = ""
	Org                = ""
	InfisicalNamespace = constants.ProtectedNamespace
	InfisicalDBUser    = "infisical"
	InfisicalDBName    = "infisical"
	InfisicalDBAddr    = ""
	InfisicalAddr      = ""

	Password = ""
)

func init() {
	// Owner = os.Getenv("OWNER")
	// InfisicalNamespace = fmt.Sprintf("user-space-%s", Owner)
	InfisicalDBUser = GetenvOrDefault("PG_USER", InfisicalDBUser)
	InfisicalDBName = GetenvOrDefault("PG_DB", InfisicalDBName)
	InfisicalDBAddr = os.Getenv("PG_ADDR")
	InfisicalAddr = GetenvOrDefault("INFISICAL_URL", "http://infisical-service."+InfisicalNamespace)

	// FIXME: use the kubesphere's iam instead of infisical
	Password = os.Getenv("PASSWORD")
}

func GetenvOrDefault(env string, d string) string {
	v := os.Getenv(env)
	if v == "" {
		return d
	}

	return v
}

func InitSuperAdmin(ctx context.Context, pg *PostgresClient) error {
	var err error
	Org, err = pg.UpdateSuperAdmin(ctx)
	if err != nil {
		klog.Error("update super admin error, ", err)
		return err
	}

	return nil
}

func InsertKsUserToPostgres(ctx context.Context, pg *PostgresClient, username, email, password string) error {
	klog.Info("insert user: ", username, ", email: ", email, ", password: ", password)

	publicKeyBytes, privateKeyBytes, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	publicKey := base64.StdEncoding.EncodeToString((*publicKeyBytes)[:])
	privateKey := base64.StdEncoding.EncodeToString((*privateKeyBytes)[:])
	encryptPrivateKey, encryptIV, encryptTag, err := infisical_crypto.Encrypt(privateKey, secretOfPassword(password))
	if err != nil {
		klog.Error("encrypt private key error, ", err)
		return err
	}

	resBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, "node", "tools/jsrp-client/client.js")
	cmd.Env = []string{
		"OWNER=" + email,
		"PASSWORD=" + password,
		"NODE_PATH=/usr/local/lib/node_modules",
	}
	cmd.Stderr = errBuf
	cmd.Stdout = resBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("jsrp-client: %v: %s", err, cmd.Stderr)
	}

	resStr := resBuf.String()
	res := strings.Split(resStr, ":")
	if len(res) < 2 {
		return fmt.Errorf("jsrp-client: return invalid, %s", resStr)
	}
	salt := res[0]
	verifier := res[1]

	dbUser := &UserPG{
		Email:      email,
		LastName:   username,
		Username:   username,
		IsAccepted: true,
	}

	dbUserEnc := &UserEncryptionKeysPG{
		PublicKey:           publicKey,
		Salt:                salt,
		Verifier:            verifier,
		EncryptedPrivateKey: encryptPrivateKey,
		IV:                  encryptIV,
		Tag:                 encryptTag,
		EncryptionVersion:   1,
	}

	userid, err := pg.SaveUser(ctx, Org, dbUser, dbUserEnc)
	if err != nil {
		return err
	}

	klog.Info("init user id, ", userid)

	return nil

}

func InsertKsUserToMongo(ctx context.Context, mongo *MongoClient, username, email, password string) error {
	klog.Info("insert user: ", username, ", email: ", email, ", password: ", password)
	publicKeyBytes, privateKeyBytes, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	publicKey := base64.StdEncoding.EncodeToString((*publicKeyBytes)[:])
	privateKey := base64.StdEncoding.EncodeToString((*privateKeyBytes)[:])
	encryptPrivateKey, encryptIV, encryptTag, err := infisical_crypto.Encrypt(privateKey, secretOfPassword(password))
	if err != nil {
		klog.Error("encrypt private key error, ", err)
		return err
	}

	resBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, "node", "tools/jsrp-client/client.js")
	cmd.Env = []string{
		"OWNER=" + email,
		"PASSWORD=" + password,
		"NODE_PATH=/usr/local/lib/node_modules",
	}
	cmd.Stderr = errBuf
	cmd.Stdout = resBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("jsrp-client: %v: %s", err, cmd.Stderr)
	}

	resStr := resBuf.String()
	res := strings.Split(resStr, ":")
	if len(res) < 2 {
		return fmt.Errorf("jsrp-client: return invalid, %s", resStr)
	}
	salt := res[0]
	verifier := res[1]

	dbUser := &UserMDB{
		Email:               email,
		LastName:            username,
		PublicKey:           publicKey,
		Salt:                salt,
		Verifier:            verifier,
		EncryptedPrivateKey: encryptPrivateKey,
		IV:                  encryptIV,
		Tag:                 encryptTag,
		EncryptionVersion:   1,
	}

	userid, err := mongo.SaveUser(ctx, dbUser)
	if err != nil {
		return err
	}

	klog.Info("init user id, ", userid)

	return nil
}

func DeleteUserFromPostgres(ctx context.Context, pg *PostgresClient, userId string) error {
	return pg.DeleteUser(ctx, userId)
}
