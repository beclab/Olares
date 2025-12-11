package runtime

import (
	"fmt"

	v1alpha1client "bytetrade.io/web3os/bfl/pkg/client/clientset/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/constants"

	"github.com/emicklei/go-restful/v3"
	"github.com/golang-jwt/jwt/v5"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	APIRootPath = "/bfl"
)

type ModuleVersion struct {
	Name    string
	Version string
}

func NewWebService(mv ModuleVersion) *restful.WebService {
	webservice := restful.WebService{}

	webservice.Path(fmt.Sprintf("%s/%s/%s", APIRootPath, mv.Name, mv.Version)).
		Produces(restful.MIME_JSON)

	return &webservice
}

func NewKubeClientInCluster() (v1alpha1client.ClientInterface, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	c, err := v1alpha1client.NewKubeClient(config)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewKubeClientWithToken(token string) (v1alpha1client.ClientInterface, error) {
	return v1alpha1client.NewKubeClientWithToken(token)
}

func ParseToken(tokenStr string) (*Claims, error) {
	if tokenStr == "" {
		return nil, fmt.Errorf("parse token err, empty token string")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return constants.KubeSphereJwtKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Alg(), jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if ok && claims.Username != "" {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token, or claims not match")
}
