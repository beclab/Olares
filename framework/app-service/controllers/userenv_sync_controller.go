package controllers

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/users"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils"

	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	userEnvConfigMapNamespace = "os-framework"
	userEnvConfigMapName      = "user-env"
	userEnvConfigMapKey       = "user-env.yaml"
)

type userEnvFile struct {
	APIVersion string                   `yaml:"apiVersion"`
	UserEnvs   []sysv1alpha1.EnvVarSpec `yaml:"userEnvs"`
}

// localeEnvAnnotations maps a user env to the user CR annotation that owns its
// value. These locale preferences are written during activation and via the
// locale settings page; the controller mirrors them into the env Value so apps
// referencing them through valueFrom always see the user's current choice.
var localeEnvAnnotations = map[string]string{
	"OLARES_USER_LANGUAGE": users.UserAnnotationLanguage,
	"OLARES_USER_THEME":    users.UserAnnotationTheme,
	"OLARES_USER_TIMEZONE": users.UserAnnotationTimezone,
}

// syncLocaleValue reconciles a locale-backed env's Value from the user CR
// annotation that owns it. The direction is driven by the env's Editable flag:
//
//   - editable == false (default): the annotation is the single source of
//     truth, so Value always tracks it. The env settings page rejects edits to
//     non-editable envs, so there is no conflicting second writer.
//   - editable == true: the value may be edited directly in settings, so the
//     annotation only seeds it when empty and never clobbers a user-set value.
//
// This lets us switch between the two models purely by flipping `editable` in
// user-env.yaml. Returns true if ue.Value was changed.
func syncLocaleValue(ue *sysv1alpha1.UserEnv, annotations map[string]string) bool {
	annoKey, ok := localeEnvAnnotations[ue.EnvName]
	if !ok {
		return false
	}
	desired := annotations[annoKey]
	if desired == "" {
		return false
	}
	if ue.Editable && ue.Value != "" {
		// user-editable and already set: respect the user's edit
		return false
	}
	if ue.Value == desired {
		return false
	}
	// never write a value outside the declared type/options
	if err := ue.ValidateValue(desired); err != nil {
		klog.Warningf("UserEnvSync: skip locale sync for %s/%s: invalid value %q: %v", ue.Namespace, ue.Name, desired, err)
		return false
	}
	ue.Value = desired
	return true
}

type UserEnvSyncController struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups=iam.kubesphere.io,resources=users,verbs=get;list;watch
//+kubebuilder:rbac:groups=sys.bytetrade.io,resources=userenvs,verbs=get;list;watch;create;patch;update

func (r *UserEnvSyncController) SetupWithManager(mgr ctrl.Manager) error {
	cmPred := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		return obj.GetNamespace() == userEnvConfigMapNamespace && obj.GetName() == userEnvConfigMapName
	})

	userPred := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		user, ok := obj.(*iamv1alpha2.User)
		if !ok {
			return false
		}
		return string(user.Status.State) == "Created"
	})

	return builder.ControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}, builder.WithPredicates(cmPred)).
		Watches(&iamv1alpha2.User{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			user, ok := obj.(*iamv1alpha2.User)
			if !ok {
				return nil
			}
			return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: user.Name}}}
		}), builder.WithPredicates(userPred)).
		Complete(r)
}

func (r *UserEnvSyncController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// the changes on the configmap triggers a sync operation for all users
	if req.Namespace == userEnvConfigMapNamespace && req.Name == userEnvConfigMapName {
		return r.reconcileAllUsers(ctx)
	}

	// the changes on a single user resource triggers a sync operation only for this particular user
	if req.Namespace == "" && req.Name != "" {
		return r.reconcileSingleUser(ctx, req.Name)
	}

	return ctrl.Result{}, nil
}

func (r *UserEnvSyncController) reconcileAllUsers(ctx context.Context) (ctrl.Result, error) {
	klog.Infof("UserEnvSync: detected %s/%s change, syncing all users", userEnvConfigMapNamespace, userEnvConfigMapName)

	base, err := r.loadBaseUserEnvFromConfigMap(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if base == nil {
		klog.Warningf("UserEnvSync: base user env config not found; skipping")
		return ctrl.Result{}, nil
	}

	var users iamv1alpha2.UserList
	if err := r.List(ctx, &users); err != nil {
		return ctrl.Result{}, fmt.Errorf("list users failed: %w", err)
	}

	failed := 0
	for i := range users.Items {
		user := &users.Items[i]
		if string(user.Status.State) != "Created" {
			continue
		}
		if _, err := r.syncUserEnvForUser(ctx, user.Name, user.Annotations, base.UserEnvs); err != nil {
			klog.Errorf("UserEnvSync: failed to sync for user %s: %v", user.Name, err)
			failed++
		}
	}

	if failed > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to sync userenv for %d users", failed)
	}
	return ctrl.Result{}, nil
}

func (r *UserEnvSyncController) reconcileSingleUser(ctx context.Context, username string) (ctrl.Result, error) {
	klog.Infof("UserEnvSync: user change detected for %s, syncing user envs", username)

	u := &iamv1alpha2.User{}
	if err := r.Get(ctx, types.NamespacedName{Name: username}, u); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if string(u.Status.State) != "Created" {
		klog.V(4).Infof("UserEnvSync: skipping user %s with state %s", username, u.Status.State)
		return ctrl.Result{}, nil
	}

	base, err := r.loadBaseUserEnvFromConfigMap(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if base == nil {
		klog.Warningf("UserEnvSync: base user env config not found; skipping for user %s", username)
		return ctrl.Result{}, nil
	}

	_, err = r.syncUserEnvForUser(ctx, username, u.Annotations, base.UserEnvs)
	return ctrl.Result{}, err
}

func (r *UserEnvSyncController) loadBaseUserEnvFromConfigMap(ctx context.Context) (*userEnvFile, error) {
	cm := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: userEnvConfigMapNamespace, Name: userEnvConfigMapName}, cm); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	content := cm.Data[userEnvConfigMapKey]
	if content == "" {
		return &userEnvFile{}, nil
	}
	var cfg userEnvFile
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("parse base user env config from cm failed: %w", err)
	}
	return &cfg, nil
}

func (r *UserEnvSyncController) syncUserEnvForUser(ctx context.Context, username string, userAnnotations map[string]string, base []sysv1alpha1.EnvVarSpec) (int, error) {
	userNs := apputils.UserspaceName(username)
	var existing sysv1alpha1.UserEnvList
	if err := r.List(ctx, &existing, client.InNamespace(userNs)); err != nil {
		return 0, fmt.Errorf("list userenvs in %s failed: %w", userNs, err)
	}

	existByName := make(map[string]*sysv1alpha1.UserEnv, len(existing.Items))
	for i := range existing.Items {
		existByName[existing.Items[i].EnvName] = &existing.Items[i]
	}

	created := 0
	for _, spec := range base {
		if ue, ok := existByName[spec.EnvName]; ok {
			original := ue.DeepCopy()
			updated := false

			if ue.Default == "" && spec.Default != "" {
				ue.Default = spec.Default
				updated = true
			}
			if ue.Type == "" && spec.Type != "" {
				ue.Type = spec.Type
				updated = true
			}
			if ue.Title == "" && spec.Title != "" {
				ue.Title = spec.Title
				updated = true
			}
			if ue.Description == "" && spec.Description != "" {
				ue.Description = spec.Description
				updated = true
			}
			if ue.RemoteOptions == "" && spec.RemoteOptions != "" {
				ue.RemoteOptions = spec.RemoteOptions
				updated = true
			}
			if ue.Regex == "" && spec.Regex != "" {
				ue.Regex = spec.Regex
				updated = true
			}

			if len(spec.Options) > 0 {
				existOpt := make(map[string]struct{}, len(ue.Options))
				for _, it := range ue.Options {
					existOpt[it.Value] = struct{}{}
				}
				for _, it := range spec.Options {
					if _, exists := existOpt[it.Value]; exists {
						continue
					}
					ue.Options = append(ue.Options, it)
					existOpt[it.Value] = struct{}{}
					updated = true
				}
			}

			// Keep the declared editability in sync for locale-managed envs so
			// flipping `editable` in user-env.yaml takes effect on already
			// provisioned users. Do this before syncLocaleValue, which reads
			// ue.Editable to decide whether to overwrite or merely seed.
			if _, isLocale := localeEnvAnnotations[spec.EnvName]; isLocale && ue.Editable != spec.Editable {
				ue.Editable = spec.Editable
				updated = true
			}
			if syncLocaleValue(ue, userAnnotations) {
				updated = true
			}

			if updated {
				if err := r.Patch(ctx, ue, client.MergeFrom(original)); err != nil {
					return created, fmt.Errorf("patch userenv %s/%s failed: %w", ue.Namespace, ue.Name, err)
				}
				klog.Infof("UserEnvSync: patched userenv %s/%s for user %s", ue.Namespace, ue.Name, username)
			}
			continue
		}
		name, err := apputils.EnvNameToResourceName(spec.EnvName)
		if err != nil {
			klog.Warningf("UserEnvSync: skip invalid env name %s for user %s: %v", spec.EnvName, username, err)
			continue
		}
		ue := &sysv1alpha1.UserEnv{}
		ue.Name = name
		ue.Namespace = userNs
		ue.EnvVarSpec = spec
		// seed locale-managed envs from the user CR annotation on first creation
		syncLocaleValue(ue, userAnnotations)
		if err := r.Create(ctx, ue); err != nil {
			return created, fmt.Errorf("create userenv %s/%s failed: %w", userNs, name, err)
		}
		created++
		klog.Infof("UserEnvSync: created userenv %s/%s for user %s", userNs, name, username)
	}
	return created, nil
}
