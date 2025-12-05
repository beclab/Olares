package kubesphere

import (
	"context"
	"errors"
	"fmt"

	v1alpha1client "bytetrade.io/web3os/app-service/pkg/client/clientset/v1alpha1"

	"github.com/dgrijalva/jwt-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	annotationGroup              = "bytetrade.io"
	userAnnotationZoneKey        = fmt.Sprintf("%s/zone", annotationGroup)
	userAnnotationOwnerRole      = fmt.Sprintf("%s/owner-role", annotationGroup)
	userAnnotationCPULimitKey    = "bytetrade.io/user-cpu-limit"
	userAnnotationMemoryLimitKey = "bytetrade.io/user-memory-limit"
	userIndex                    = "bytetrade.io/user-index"
)

type Options struct {
	JwtSecret string `yaml:"jwtSecret"`
}

type Config struct {
	AuthenticationOptions *Options `yaml:"authentication,omitempty"`
}

type Type string

type Claims struct {
	jwt.StandardClaims
	// Private Claim Names
	// TokenType defined the type of the token
	TokenType Type `json:"token_type,omitempty"`
	// Username user identity, deprecated field
	Username string `json:"username,omitempty"`
	// Extra contains the additional information
	Extra map[string][]string `json:"extra,omitempty"`

	// Used for issuing authorization code
	// Scopes can be used to request that specific sets of information be made available as Claim Values.
	Scopes []string `json:"scopes,omitempty"`

	// The following is well-known ID Token fields

	// End-User's full name in displayable form including all name parts,
	// possibly including titles and suffixes, ordered according to the End-User's locale and preferences.
	Name string `json:"name,omitempty"`
	// String value used to associate a Client session with an ID Token, and to mitigate replay attacks.
	// The value is passed through unmodified from the Authentication Request to the ID Token.
	Nonce string `json:"nonce,omitempty"`
	// End-User's preferred e-mail address.
	Email string `json:"email,omitempty"`
	// End-User's locale, represented as a BCP47 [RFC5646] language tag.
	Locale string `json:"locale,omitempty"`
	// Shorthand name by which the End-User wishes to be referred to at the RP,
	PreferredUsername string `json:"preferred_username,omitempty"`
}

func getLLdapJwtKey(ctx context.Context, kubeConfig *rest.Config) ([]byte, error) {
	kubeClientInService, err := v1alpha1client.NewKubeClient("", kubeConfig)
	if err != nil {
		return nil, err
	}

	secret, err := kubeClientInService.Kubernetes().
		CoreV1().Secrets("os-platform").
		Get(ctx, "lldap-credentials", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	jwtSecretKey, ok := secret.Data["lldap-jwt-secret"]
	if !ok {
		return nil, fmt.Errorf("failed to get lldap jwt secret")
	}

	return jwtSecretKey, nil
}

// ValidateToken validates a token by performing an authentication check.
func ValidateToken(ctx context.Context, kubeConfig *rest.Config, tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		jwtSecretKey, err := getLLdapJwtKey(ctx, kubeConfig)
		if err != nil {
			return nil, err
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.Username, nil
	}
	return "", fmt.Errorf("invalid token, or claims not match")
}

// GetUserZone returns user zone, an error if there is any.
func GetUserZone(ctx context.Context, username string) (string, error) {
	return GetUserAnnotation(ctx, username, userAnnotationZoneKey)
}

// GetUserRole returns user role, an error if there is any.
func GetUserRole(ctx context.Context, username string) (string, error) {
	return GetUserAnnotation(ctx, username, userAnnotationOwnerRole)
}

// GetUserAnnotation returns user annotation, an error if there is any.
func GetUserAnnotation(ctx context.Context, username, annotation string) (string, error) {
	gvr := schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
	config, err := ctrl.GetConfig()
	if err != nil {
		return "", err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", err
	}
	data, err := client.Resource(gvr).Get(ctx, username, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get user=%s err=%v", username, err)
		return "", err
	}

	a, ok := data.GetAnnotations()[annotation]
	if !ok {
		return "", fmt.Errorf("user annotation %s not found", annotation)
	}

	return a, nil
}

// GetUserCPULimit returns user cpu limit value, an error if there is any.
func GetUserCPULimit(ctx context.Context, username string) (string, error) {
	return GetUserAnnotation(ctx, username, userAnnotationCPULimitKey)
}

// GetUserMemoryLimit returns user memory limit value, an error if there is any.
func GetUserMemoryLimit(ctx context.Context, username string) (string, error) {
	return GetUserAnnotation(ctx, username, userAnnotationMemoryLimitKey)
}

// GetAdminUsername returns admin username, an error if there is any.
func GetAdminUsername(ctx context.Context, kubeConfig *rest.Config) (string, error) {
	gvr := schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
	client, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return "", err
	}
	data, err := client.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get user list err=%v", err)
		return "", err
	}

	var admin string
	for _, u := range data.Items {
		if u.Object == nil {
			continue
		}
		annotations := u.GetAnnotations()
		role := annotations[userAnnotationOwnerRole]
		if role == "owner" || role == "admin" {
			admin = u.GetName()
			break
		}
	}

	return admin, nil
}

func GetUserIndexByName(ctx context.Context, name string) (string, error) {
	return GetUserAnnotation(ctx, name, userIndex)
}

type UserInfo struct {
	Name string
	Role string
}

// GetAdminUserList returns admin list, an error if there is any.
func GetAdminUserList(ctx context.Context, kubeConfig *rest.Config) ([]UserInfo, error) {
	adminUserList := make([]UserInfo, 0)

	gvr := schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
	client, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return adminUserList, err
	}
	data, err := client.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to get user list err=%v", err)
		return adminUserList, err
	}

	for _, u := range data.Items {
		if u.Object == nil {
			continue
		}
		annotations := u.GetAnnotations()
		role := annotations[userAnnotationOwnerRole]
		if role == "owner" || role == "admin" {
			adminUserList = append(adminUserList, UserInfo{Name: u.GetName(), Role: role})
		}
	}

	return adminUserList, nil
}

func IsAdmin(ctx context.Context, kubeConfig *rest.Config, owner string) (bool, error) {
	adminList, err := GetAdminUserList(ctx, kubeConfig)
	if err != nil {
		return false, err
	}
	for _, user := range adminList {
		if user.Name == owner {
			return true, nil
		}
	}
	return false, nil
}

func GetOwner(ctx context.Context, kubeConfig *rest.Config) (string, error) {
	adminList, err := GetAdminUserList(ctx, kubeConfig)
	if err != nil {
		return "", err
	}
	for _, user := range adminList {
		if user.Role == "owner" {
			return user.Name, nil
		}
	}
	return "", errors.New("user with role owner not found")
}
