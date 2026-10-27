package certmanager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	"bytetrade.io/web3os/bfl/pkg/constants"
	iamv1 "github.com/beclab/api/iam/v1alpha2"
	"k8s.io/apimachinery/pkg/util/validation"
)

// NewCertManagerForUser snapshots this user's cloud name and JWS together.
// The local username selects credentials; it is not a cloud account name.
func NewCertManagerForUser(ctx context.Context, username string) (Interface, error) {
	op, err := operator.NewUserOperatorWithContext(ctx)
	if err != nil {
		return nil, err
	}
	user, err := op.GetUser(username)
	if err != nil {
		return nil, err
	}
	return newCertManagerForUser(ctx, user)
}

func newCertManagerForUser(ctx context.Context, user *iamv1.User) (Interface, error) {
	if user == nil {
		return nil, fmt.Errorf("cloud credential user is missing")
	}
	jws := strings.TrimSpace(user.Annotations[constants.UserCertManagerJWSToken])
	if jws == "" {
		return nil, fmt.Errorf("user has no cloud JWS credentials")
	}
	name, err := cloudNameFromJWS(jws)
	if err != nil {
		return nil, err
	}
	return NewCertManagerWithCredentials(ctx, name, jws), nil
}

// cloudNameFromJWS reads the asserted cloud identity for API routing only.
// The cloud service still verifies the signature and account authorization.
func cloudNameFromJWS(jws string) (constants.TerminusName, error) {
	parts := strings.Split(jws, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid cloud JWS envelope")
	}
	decoded := make([][]byte, len(parts))
	for i, part := range parts {
		var err error
		decoded[i], err = base64.RawURLEncoding.DecodeString(part)
		if err != nil || len(decoded[i]) == 0 {
			return "", fmt.Errorf("invalid cloud JWS encoding")
		}
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if json.Unmarshal(decoded[0], &header) != nil || header.Alg != "EdDSA" || header.Kid == "" {
		return "", fmt.Errorf("invalid cloud JWS header")
	}
	var payload struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(decoded[1], &payload) != nil {
		return "", fmt.Errorf("invalid cloud JWS payload")
	}
	identity := strings.Split(payload.Name, "@")
	if len(identity) != 2 || len(validation.IsDNS1123Label(identity[0])) != 0 || len(validation.IsDNS1123Subdomain(identity[1])) != 0 {
		return "", fmt.Errorf("cloud JWS requires a fully qualified account name")
	}
	return constants.TerminusName(payload.Name), nil
}
