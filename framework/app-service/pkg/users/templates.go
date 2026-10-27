package users

import (
	"fmt"
)

var (
	AnnotationGroup       = "bytetrade.io"
	AnnotationUserCreator = "bytetrade.io/creator"
	AnnotationUserDeleter = "bytetrade.io/deleter"

	UserAnnotationTerminusNameKey = fmt.Sprintf("%s/terminus-name", AnnotationGroup)
	UserAvatar                    = fmt.Sprintf("%s/avatar", AnnotationGroup)

	UserAnnotationZoneKey = fmt.Sprintf("%s/zone", AnnotationGroup)

	UserAnnotationUninitialized = fmt.Sprintf("%s/uninitialized", AnnotationGroup)

	UserAnnotationOwnerRole = fmt.Sprintf("%s/owner-role", AnnotationGroup)

	UserAnnotationIsEphemeral = fmt.Sprintf("%s/is-ephemeral", AnnotationGroup)

	// Locale preferences set during activation / via the locale settings page.
	// These act as the source of truth for the matching OLARES_USER_* envs,
	// which UserEnvSyncController keeps in sync one-way.
	UserAnnotationLanguage = fmt.Sprintf("%s/language", AnnotationGroup)
	UserAnnotationLocation = fmt.Sprintf("%s/location", AnnotationGroup)
	UserAnnotationTimezone = fmt.Sprintf("%s/timezone", AnnotationGroup)
	UserAnnotationTheme    = fmt.Sprintf("%s/theme", AnnotationGroup)

	UserLauncherAuthPolicy        = fmt.Sprintf("%s/launcher-auth-policy", AnnotationGroup)
	UserLauncherAccessLevel       = fmt.Sprintf("%s/launcher-access-level", AnnotationGroup)
	UserAnnotationLimitsCpuKey    = "bytetrade.io/user-cpu-limit"
	UserAnnotationLimitsMemoryKey = "bytetrade.io/user-memory-limit"
)
