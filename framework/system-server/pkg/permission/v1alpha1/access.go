package permission

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	sysv1alpha1 "bytetrade.io/web3os/system-server/pkg/apis/sys/v1alpha1"
	"bytetrade.io/web3os/system-server/pkg/utils"

	"github.com/jellydator/ttlcache/v3"
	"golang.org/x/crypto/bcrypt"
	"k8s.io/klog/v2"
)

const (
	TokenCacheTTL      = 5 * time.Minute
	TokenCacheCapacity = 1000
)

type AccessManager struct {
	cache *ttlcache.Cache[string, *sysv1alpha1.PermissionRequire]
}

func NewAccessManager() *AccessManager {
	return &AccessManager{
		cache: ttlcache.New(
			ttlcache.WithTTL[string, *sysv1alpha1.PermissionRequire](TokenCacheTTL),
			ttlcache.WithCapacity[string, *sysv1alpha1.PermissionRequire](TokenCacheCapacity),
		),
	}
}

func (a *AccessManager) getAccessToken(accReq *AccessTokenRequest, ap *sysv1alpha1.ApplicationPermission) (string, error) {

	now := time.Now().UnixMilli() / 1000 // to seconds
	if math.Abs(float64(now-accReq.Timestamp)) > 10 {
		return "", errors.New("request time expired")
	}

	compareHash := accReq.AppKey + strconv.Itoa(int(accReq.Timestamp)) + ap.Spec.Secret
	if err := bcrypt.CompareHashAndPassword([]byte(accReq.Token), []byte(compareHash)); err != nil {
		klog.Error("invalid request: ", utils.PrettyJSON(accReq))
		return "", fmt.Errorf("invalid auth token: %s", err.Error())
	}

	return base64.StdEncoding.EncodeToString([]byte(accReq.Token)), nil
}

func (a *AccessManager) cacheAccessToken(token string, permReq *sysv1alpha1.PermissionRequire) {
	a.cache.Set(token, permReq, TokenCacheTTL)
}

func (a *AccessManager) getPermWithToken(token string) (*sysv1alpha1.PermissionRequire, error) {
	perm := a.cache.Get(token)
	if perm == nil {
		return nil, errors.New("token not found in cache or expired")
	}

	return perm.Value(), nil
}
