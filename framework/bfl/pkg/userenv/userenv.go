// Package userenv provides thin read/write helpers over the per-user UserEnv
// (sys.bytetrade.io/v1alpha1) custom resources. bfl's config-system (locale
// settings) endpoints use it so the locale values share a single source of
// truth with the app-service env settings page.
package userenv

import (
	"context"
	"fmt"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/constants"

	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Locale-related user environment variables backed by per-user UserEnv CRs.
// These mirror the entries declared in build/user-env.yaml and the locale map
// in app-service's UserEnvSyncController.
const (
	EnvLanguage = "OLARES_USER_LANGUAGE"
	EnvTimezone = "OLARES_USER_TIMEZONE"
	EnvTheme    = "OLARES_USER_THEME"
)

// resourceName mirrors app-service's EnvNameToResourceName: lowercase the env
// name and replace underscores with hyphens (OLARES_USER_LANGUAGE ->
// olares-user-language).
func resourceName(envName string) string {
	return strings.ReplaceAll(strings.ToLower(envName), "_", "-")
}

func namespace(username string) string {
	return fmt.Sprintf(constants.UserspaceNameFormat, username)
}

// GetValue returns the stored Value of the given user env for the user. ok is
// false when the UserEnv has not been provisioned yet (e.g. during activation,
// before the sync controller creates it from the user annotation); callers
// should fall back to the annotation in that case.
func GetValue(ctx context.Context, c client.Client, username, envName string) (value string, ok bool, err error) {
	var ue sysv1alpha1.UserEnv
	key := types.NamespacedName{Namespace: namespace(username), Name: resourceName(envName)}
	if getErr := c.Get(ctx, key, &ue); getErr != nil {
		if apierrors.IsNotFound(getErr) {
			return "", false, nil
		}
		return "", false, getErr
	}
	return ue.GetEffectiveValue(), true, nil
}

// SetValue writes value into the user env's Value field. ok is false (a no-op)
// when the UserEnv does not exist yet; the caller-written user annotation seeds
// it once app-service's sync controller provisions it. Editability is enforced
// by the env settings API, not here: config-system is a trusted system writer.
func SetValue(ctx context.Context, c client.Client, username, envName, value string) (ok bool, err error) {
	var ue sysv1alpha1.UserEnv
	key := types.NamespacedName{Namespace: namespace(username), Name: resourceName(envName)}
	if getErr := c.Get(ctx, key, &ue); getErr != nil {
		if apierrors.IsNotFound(getErr) {
			return false, nil
		}
		return false, getErr
	}
	if ue.Value == value {
		return true, nil
	}
	original := ue.DeepCopy()
	ue.Value = value
	if patchErr := c.Patch(ctx, &ue, client.MergeFrom(original)); patchErr != nil {
		return false, patchErr
	}
	return true, nil
}
