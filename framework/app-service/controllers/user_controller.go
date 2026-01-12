package controllers

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	natsevent "github.com/beclab/Olares/framework/app-service/pkg/event"
	"github.com/beclab/Olares/framework/app-service/pkg/users"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace/v1"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/beclab/Olares/framework/app-service/pkg/utils/sliceutil"

	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/beclab/lldap-client/pkg/cache/memory"
	lclient "github.com/beclab/lldap-client/pkg/client"
	lconfig "github.com/beclab/lldap-client/pkg/config"
	lapierrors "github.com/beclab/lldap-client/pkg/errors"
	"github.com/beclab/lldap-client/pkg/generated"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	applyCorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyMetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	needSyncToLLdapAna = "iam.kubesphere.io/sync-to-lldap"
	syncedToLLdapAna   = "iam.kubesphere.io/synced-to-lldap"
	userIndexAna       = "bytetrade.io/user-index"
	interval           = time.Second
	timeout            = 15 * time.Second
)

// UserController reconciles a User object
type UserController struct {
	client.Client
	KubeConfig  *rest.Config
	LLdapClient *lclient.Client
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserController) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("user-controller", mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler:              r,
	})
	if err != nil {
		klog.Errorf("user-controller setup failed %v", err)
		return fmt.Errorf("user-controller setup failed %w", err)
	}

	err = c.Watch(source.Kind(
		mgr.GetCache(),
		&iamv1alpha2.User{},
		handler.TypedEnqueueRequestsFromMapFunc(
			func(ctx context.Context, user *iamv1alpha2.User) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Name: user.GetName(),
				}}}
			}),
		predicate.TypedFuncs[*iamv1alpha2.User]{
			CreateFunc: func(e event.TypedCreateEvent[*iamv1alpha2.User]) bool {
				obj := e.Object
				if obj.Status.State == "Failed" {
					return false
				}
				klog.Infof("create enque name: %s, state: %s", obj.Name, obj.Status.State)
				return true
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*iamv1alpha2.User]) bool {
				oldObj := e.ObjectOld
				newObj := e.ObjectNew
				oldObj.Spec.InitialPassword = newObj.Spec.InitialPassword

				isDeletionUpdate := newObj.DeletionTimestamp != nil
				specChanged := !reflect.DeepEqual(oldObj.Spec, newObj.Spec)

				shouldReconcile := isDeletionUpdate || specChanged
				return shouldReconcile
				//return true
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*iamv1alpha2.User]) bool {
				return true
			},
		},
	))

	if err != nil {
		klog.Errorf("user-controller add watch failed %v", err)
		return fmt.Errorf("add watch failed %w", err)
	}

	return nil
}

// Reconcile is part of the main kubernetes reconciliation loop
func (r *UserController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("start reconcile user %s", req.Name)

	// Fetch the User instance
	user := &iamv1alpha2.User{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// User was deleted, handle cleanup if needed
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if user.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sliceutil.HasString(user.Finalizers, userFinalizer) {
			user.ObjectMeta.Finalizers = append(user.ObjectMeta.Finalizers, userFinalizer)
			if updateErr := r.Update(ctx, user); updateErr != nil {
				klog.Errorf("failed to update user %v", err)
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(user.ObjectMeta.Finalizers, userFinalizer) {
			if r.LLdapClient != nil {
				if err = r.waitForDeleteFromLLDAP(user.Name); err != nil {
					klog.Infof("wait for delete user from lldap failed %v", err)
					return ctrl.Result{}, err
				}
			}
			if err = r.deleteRoleBindings(ctx, user); err != nil {
				klog.V(0).Infof("delete rolebinding failed %v", err)
				return ctrl.Result{}, err
			}
			err = r.handleUserDeletion(ctx, user)
			if err != nil {
				klog.Errorf("failed to delete user resource %v", err)
				return ctrl.Result{}, err
			}

			user.Finalizers = sliceutil.RemoveString(user.ObjectMeta.Finalizers, func(item string) bool {
				return item == userFinalizer
			})
			if updateErr := r.Update(ctx, user, &client.UpdateOptions{}); updateErr != nil {
				klog.Infof("update user failed %v", updateErr)
				return ctrl.Result{}, updateErr
			}
			r.publish("Delete", user.Name, user.Annotations[users.AnnotationUserDeleter])
		}
		return ctrl.Result{}, nil
	}
	if r.LLdapClient == nil {
		lldapClient, err := r.getLLdapClient()
		if err != nil {
			return ctrl.Result{}, err
		}
		r.LLdapClient = lldapClient
	}

	if r.LLdapClient != nil {
		if err = r.waitForSyncToLLDAP(user); err != nil {
			klog.V(0).Infof("wait for sync to lldap failed %v", err)
			return ctrl.Result{RequeueAfter: time.Second}, nil
		}
		klog.V(0).Infof("user %s sync to lldap successes", user.Name)
	}

	if user.Status.State == "" || user.Status.State == "Creating" {
		ret, err := r.handleUserCreation(ctx, user)
		time.Sleep(time.Second)
		return ret, err
	}
	klog.Infof("finish reconcile user %s", req.Name)

	return ctrl.Result{}, nil
}

func (r *UserController) deleteRoleBindings(ctx context.Context, user *iamv1alpha2.User) error {
	if len(user.Name) > validation.LabelValueMaxLength {
		// ignore invalid label value error
		return nil
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	err := r.Client.DeleteAllOf(ctx, clusterRoleBinding, client.MatchingLabels{iamv1alpha2.UserReferenceLabel: user.Name})
	if err != nil {
		klog.Errorf("failed to delete all of clusterrolebinding %v", err)
		return err
	}

	roleBindingList := &rbacv1.RoleBindingList{}
	err = r.Client.List(ctx, roleBindingList, client.MatchingLabels{iamv1alpha2.UserReferenceLabel: user.Name})
	if err != nil {
		klog.Errorf("failed to get rolebindinglist %v", err)
		return err
	}

	for _, roleBinding := range roleBindingList.Items {
		err = r.Client.Delete(ctx, &roleBinding)
		if err != nil {
			klog.Errorf("failed to delete rolebinding %v", err)
			return err
		}
	}
	return nil
}

func (r *UserController) handleUserCreation(ctx context.Context, user *iamv1alpha2.User) (ctrl.Result, error) {
	klog.Infof("starting user creation for %s", user.Name)

	// Update status to Creating
	if user.Status.State != "Creating" {
		err := r.updateUserStatus(ctx, user, "Creating", "Starting user creation process")
		if err != nil {
			klog.Errorf("failed to update user status to Created %v", err)
			return ctrl.Result{}, err
		}
	}

	// Check cluster pod capacity
	klog.Infof("start check cluster pod capacity.....")
	isSatisfied, err := r.checkClusterPodCapacity(ctx)
	if err != nil {
		message := fmt.Sprintf("failed to check cluster capacity %v", err)
		klog.Error(message)
		updateErr := r.updateUserStatus(ctx, user, "Failed", message)
		if updateErr != nil {
			klog.Errorf("failed to update user status to Created %v", updateErr)
		}
		return ctrl.Result{}, updateErr
	}
	if !isSatisfied {
		updateErr := r.updateUserStatus(ctx, user, "Failed", "Insufficient pods can allocate in the cluster")
		if updateErr != nil {
			klog.Errorf("failed to update user status to Failed %v", updateErr)
		}
		return ctrl.Result{}, updateErr
	}

	// Validate resource limits
	klog.Infof("start to validate resource limits.....")

	err = r.validateResourceLimits(user)
	// invalid resource limit, no need to requeue
	if err != nil {
		klog.Errorf("failed to validate resource limits %v", err)
		updateErr := r.updateUserStatus(ctx, user, "Failed", err.Error())
		if updateErr != nil {
			klog.Errorf("failed to update user status: %v", updateErr)
		}
		return ctrl.Result{}, updateErr
	}

	klog.Infof("start to checkResource.....")

	err = r.checkResource(user)
	if err != nil {
		klog.Errorf("failed to checkResource %v", err)
		updateErr := r.updateUserStatus(ctx, user, "Failed", err.Error())
		if updateErr != nil {
			klog.Errorf("failed to update user status to Failed %v", updateErr)
		}
		return ctrl.Result{}, updateErr
	}

	// Create user resources
	err = r.createUserResources(ctx, user)
	if err != nil {
		klog.Errorf("failed to create user resource %v", err)
		updateErr := r.updateUserStatus(ctx, user, "Failed", fmt.Sprintf("Failed to create user resources: %v", err))
		if updateErr != nil {
			klog.Errorf("failed to update user status: %v", updateErr)
		}
		return ctrl.Result{}, updateErr
	}
	klog.Infof("create user resource success: %s", user.Name)
	updateErr := r.updateUserStatus(ctx, user, "Created", "Created user success")
	if updateErr != nil {
		klog.Errorf("failed to update user status to Created %v", updateErr)
	} else {
		klog.Infof("publish user creation event.....")
		r.publish("Create", user.Name, user.Annotations[users.AnnotationUserCreator])
	}
	return ctrl.Result{}, updateErr
}

func (r *UserController) publish(topic, user, operator string) {
	natsevent.PublishUserEventToQueue(topic, user, operator)
}

func (r *UserController) checkResource(user *iamv1alpha2.User) error {
	metrics, _, err := apputils.GetClusterResource("")
	if err != nil {
		return err
	}
	memoryLimit := user.Annotations[users.UserAnnotationLimitsMemoryKey]

	memory, _ := resource.ParseQuantity(memoryLimit)
	if memory.CmpInt64(int64(metrics.Memory.Total-metrics.Memory.Usage)) >= 0 {
		return fmt.Errorf("unable to create user: Insufficient memory available in the cluster to meet the quota, required is: %.0f bytes, but available is: %.0f bytes", memory.AsApproximateFloat64(), metrics.Memory.Total-metrics.Memory.Usage)
	}
	return nil
}

func (r *UserController) handleUserDeletion(ctx context.Context, user *iamv1alpha2.User) error {
	klog.Infof("starting user deletion for %s", user.Name)

	// Update status to Deleting if not already
	if user.Status.State != "Deleting" {
		updateErr := r.updateUserStatus(ctx, user, "Deleting", "Starting user deletion process")
		if updateErr != nil {
			klog.Errorf("failed to update user %v", updateErr)
			return updateErr
		}
	}

	// Clean up user resources
	err := r.cleanupUserResources(ctx, user)
	if err != nil {
		klog.Errorf("failed to cleanup user resources: %v", err)
		return err
	}
	// wait for user-space, user-system namespace to be deleted
	userspaceNs := fmt.Sprintf("user-space-%s", user.Name)
	userSystemNs := fmt.Sprintf("user-system-%s", user.Name)
	userspaceExist, userSystemExist := true, true
	err = utilwait.PollImmediate(2*time.Second, 5*time.Minute, func() (done bool, err error) {
		var ns corev1.Namespace
		err = r.Get(ctx, types.NamespacedName{Name: userspaceNs}, &ns)
		if apierrors.IsNotFound(err) {
			userspaceExist = false
		}
		err = r.Get(ctx, types.NamespacedName{Name: userSystemNs}, &ns)
		if apierrors.IsNotFound(err) {
			userSystemExist = false
		}
		if !userspaceExist && !userSystemExist {
			return true, nil
		}
		return false, nil

	})
	if err != nil {
		klog.Errorf("wait for user namespace to deleted failed %v", err)
		return err
	}
	return nil
}

func (r *UserController) validateResourceLimits(user *iamv1alpha2.User) error {
	return users.ValidateResourceLimits(user)
}

func (r *UserController) waitForDeleteFromLLDAP(username string) error {
	err := utilwait.PollImmediate(interval, timeout, func() (done bool, err error) {
		err = r.LLdapClient.Users().Delete(context.TODO(), username)
		if err != nil && lapierrors.IsNotFound(err) {
			klog.Error(err)
			return false, err
		}
		return true, nil
	})
	return err
}

func (r *UserController) createUserResources(ctx context.Context, user *iamv1alpha2.User) error {
	// Create user using userspace manager
	klog.Infof("creating user resources for %s", user.Name)

	// create globalrolebinding
	globalRoleBinding := iamv1alpha2.GlobalRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: iamv1alpha2.SchemeGroupVersion.String(),
			Kind:       iamv1alpha2.ResourceKindGlobalRoleBinding,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: user.Name,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.String(),
			Kind:     iamv1alpha2.ResourceKindGlobalRole,
			Name:     getGlobalRole(user.Annotations[users.UserAnnotationOwnerRole]),
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: iamv1alpha2.SchemeGroupVersion.String(),
				Kind:     iamv1alpha2.ResourceKindUser,
				Name:     user.Name,
			},
		},
	}
	err := r.Create(ctx, &globalRoleBinding)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		klog.Errorf("failed to create gloabalrolebinding %v", err)
		return err
	}

	err = r.createNamespace(ctx, user)
	if err != nil {
		klog.Errorf("failed to create namespace %v", err)
		return err
	}

	ksClient, err := kubernetes.NewForConfig(r.KubeConfig)
	if err != nil {
		klog.Errorf("make ksClient failed %v", err)
		return err
	}

	// copy ssl configmap to new userspace
	var applyCm *applyCorev1.ConfigMapApplyConfiguration
	creatorUser, err := utils.FindOwnerUser(r.Client, user)
	if err != nil {
		klog.Errorf("failed to find user with owner role %v", err)
		return err
	}

	ownerUserspace := fmt.Sprintf("user-space-%s", creatorUser.Name)
	nsName := fmt.Sprintf("user-space-%s", user.Name)
	sslConfig, err := ksClient.CoreV1().ConfigMaps(ownerUserspace).Get(ctx, "zone-ssl-config", metav1.GetOptions{})
	if err == nil && sslConfig != nil {
		sslConfig.Data["ephemeral"] = "true"

		applyCm = NewApplyConfigmap(nsName, sslConfig.Data)
		_, err = ksClient.CoreV1().ConfigMaps(nsName).Apply(ctx, applyCm, metav1.ApplyOptions{
			FieldManager: "application/apply-patch"})
		if err != nil {
			klog.Errorf("failed to apply configmap %v", err)
			return err
		}
	}

	err = r.createUserApps(ctx, user)
	if err != nil {
		klog.Errorf("failed to create user apps %v", err)
		return err
	}

	return nil
}

func (r *UserController) createNamespace(ctx context.Context, user *iamv1alpha2.User) error {

	// create namespace user-space-<user>
	userspaceNs := fmt.Sprintf("user-space-%s", user.Name)
	userSystemNs := fmt.Sprintf("user-system-%s", user.Name)
	creatorUser, err := utils.FindOwnerUser(r.Client, user)
	if err != nil {
		klog.Error(err)
		return err
	}

	// create user-space namespace
	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: userspaceNs,
			Annotations: map[string]string{
				creator: creatorUser.Name,
			},
			Finalizers: []string{
				namespaceFinalizer,
			},
		},
	}
	err = r.Create(ctx, &ns)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		klog.Errorf("failed to create user-space namespace %v", err)
		return err
	}

	// create user-system namespace
	userSystemNamespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: userSystemNs,
			Annotations: map[string]string{
				"kubesphere.io/creator": "",
			},
			Finalizers: []string{
				namespaceFinalizer,
			},
		},
	}
	err = r.Create(ctx, &userSystemNamespace)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		klog.Errorf("failed to create user-system namespace %v", err)
		return err
	}
	return nil
}

func (r *UserController) createUserApps(ctx context.Context, user *iamv1alpha2.User) error {
	creator := userspace.NewCreator(r.Client, r.KubeConfig, user.Name)
	_, _, err := creator.CreateUserApps(ctx)

	if err != nil {
		klog.Errorf("failed to create user apps %v", err)
		return err
	}
	return nil
}

func (r *UserController) cleanupUserResources(ctx context.Context, user *iamv1alpha2.User) error {
	deleter := userspace.NewDeleter(r.Client, r.KubeConfig, user.Name)
	err := deleter.DeleteUserResource(ctx)
	if err != nil {
		klog.Errorf("failed to delete user %v", err)
		return err
	}

	return nil
}

func (r *UserController) checkClusterPodCapacity(ctx context.Context) (bool, error) {
	return users.CheckClusterPodCapacity(ctx, r.Client)
}

func (r *UserController) updateUserStatus(ctx context.Context, user *iamv1alpha2.User, state, reason string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the latest version of the user
		latestUser := &iamv1alpha2.User{}
		err := r.Get(ctx, types.NamespacedName{Name: user.Name}, latestUser)
		if err != nil {
			return err
		}

		latestUser.Status.State = iamv1alpha2.UserState(state)
		latestUser.Status.Reason = reason

		return r.Update(ctx, latestUser)
	})

}

func (r *UserController) getCredentialVal(ctx context.Context, key string) (string, error) {
	var secret corev1.Secret
	k := types.NamespacedName{Name: "lldap-credentials", Namespace: "os-platform"}
	err := r.Client.Get(ctx, k, &secret)
	if err != nil {
		return "", err
	}
	if value, ok := secret.Data[key]; ok {
		return string(value), nil
	}
	return "", fmt.Errorf("can not find credentialval for key %s", key)

}

func (r *UserController) getLLdapClient() (*lclient.Client, error) {
	bindUsername, err := r.getCredentialVal(context.TODO(), "lldap-ldap-user-dn")
	if err != nil {
		klog.Infof("get lldap secret failed %v", err)
		return nil, err
	}
	bindPassword, err := r.getCredentialVal(context.TODO(), "lldap-ldap-user-pass")
	if err != nil {
		klog.Infof("get lldap secret failed %v", err)
		return nil, err
	}

	lldapClient, err := lclient.New(&lconfig.Config{
		Host:       "http://lldap-service.os-platform:17170",
		Username:   bindUsername,
		Password:   bindPassword,
		TokenCache: memory.New(),
	})
	if err != nil {
		klog.Infof("get lldap client failed %v", err)
		return nil, err
	}
	return lldapClient, nil
}

func (r *UserController) waitForSyncToLLDAP(user *iamv1alpha2.User) error {
	ana := user.Annotations
	if ana == nil {
		return nil
	}
	isNeedSyncToLLDap, _ := strconv.ParseBool(ana[needSyncToLLdapAna])
	//synced, _ := strconv.ParseBool(ana[syncedToLLdapAna])
	if !isNeedSyncToLLDap {
		return nil
	}
	var userIndex int

	err := utilwait.PollImmediate(interval, timeout, func() (done bool, err error) {
		klog.Infof("poll info from lldap...")
		_, err = r.LLdapClient.Users().Get(context.TODO(), user.Name)

		if err != nil {
			// user not synced to lldap
			if lapierrors.IsNotFound(err) {
				u := generated.CreateUserInput{
					Id:          user.Name,
					Email:       user.Spec.Email,
					DisplayName: user.Name,
				}
				userRes, err := r.LLdapClient.Users().Create(context.TODO(), &u, user.Spec.InitialPassword)
				if err != nil && !lapierrors.IsAlreadyExists(err) {
					return false, err
				}
				// user created success in lldap

				userIndex = userRes.CreateUser.UserIndex

				for _, groupName := range user.Spec.Groups {
					var gid int
					g, err := r.LLdapClient.Groups().GetByName(context.TODO(), groupName)
					if err != nil && !lapierrors.IsNotFound(err) {
						return false, err
					}

					if err == nil {
						// group already exist in lldap
						gid = g.Id
					}

					// group does not exist in lldap, so create it
					if lapierrors.IsNotFound(err) {
						newGroup, err := r.LLdapClient.Groups().Create(context.TODO(), groupName, "")
						if err != nil && !lapierrors.IsAlreadyExists(err) {
							return false, err
						}

						if err == nil {
							gid = newGroup.Id
						}
					}
					if gid == 0 {
						return false, errors.New("invalid group id")
					}
					err = r.LLdapClient.Groups().AddUser(context.TODO(), user.Name, gid)
					if err != nil && !lapierrors.IsAlreadyExists(err) {
						return false, err
					}
				}

			} else {
				return false, err
			}
		} else {
			// user already exists in lldap, should add/remove group
			u, err := r.LLdapClient.Users().Get(context.TODO(), user.Name)
			if err != nil {
				return false, err
			}
			userIndex = u.UserIndex

			getGroups := func(u *generated.GetUserDetailsUser) (groups []string) {
				for _, group := range u.Groups {
					groups = append(groups, group.DisplayName)
				}
				return groups
			}
			oldGroups := sets.NewString(getGroups(u)...)
			curGroups := sets.NewString(user.Spec.Groups...)
			groupToDelete := oldGroups.Difference(curGroups)
			groupToAdd := curGroups.Difference(oldGroups)

			for groupName := range groupToDelete {
				group, err := r.LLdapClient.Groups().GetByName(context.TODO(), groupName)
				if err != nil {
					return false, err
				}
				err = r.LLdapClient.Groups().RemoveUser(context.TODO(), user.Name, group.Id)
				if err != nil {
					return false, err
				}
			}
			for groupName := range groupToAdd {
				groupId := 0
				group, err := r.LLdapClient.Groups().GetByName(context.TODO(), groupName)
				if err != nil {
					if !lapierrors.IsNotFound(err) {
						return false, err
					}
					groupNew, err := r.LLdapClient.Groups().Create(context.TODO(), groupName, "")
					if err != nil && !lapierrors.IsAlreadyExists(err) {
						return false, err
					}
					groupId = groupNew.Id
				} else {
					groupId = group.Id
				}
				err = r.LLdapClient.Groups().AddUser(context.TODO(), user.Name, groupId)
				if err != nil && !lapierrors.IsAlreadyExists(err) {
					return false, err
				}
			}
		}
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			var u iamv1alpha2.User
			err = r.Get(context.TODO(), types.NamespacedName{Name: user.Name}, &u)
			if err != nil {
				return err
			}
			u.Annotations[syncedToLLdapAna] = "true"
			u.Annotations[userIndexAna] = strconv.FormatInt(int64(userIndex-2), 10)
			u.Spec.InitialPassword = ""
			err = r.Update(context.TODO(), &u, &client.UpdateOptions{})
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return false, err
		}

		return true, nil
	})
	klog.V(0).Infof("poll result %v", err)
	return err
}

// UserCreateOption represents the options for creating a user
type UserCreateOption struct {
	Name         string
	OwnerRole    string
	DisplayName  string
	Email        string
	Password     string
	Description  string
	TerminusName string
	MemoryLimit  string
	CpuLimit     string
}

func NewApplyConfigmap(namespace string, data map[string]string) *applyCorev1.ConfigMapApplyConfiguration {
	return &applyCorev1.ConfigMapApplyConfiguration{
		TypeMetaApplyConfiguration: applyMetav1.TypeMetaApplyConfiguration{
			Kind:       pointer.String("ConfigMap"),
			APIVersion: pointer.String(corev1.SchemeGroupVersion.String()),
		},
		ObjectMetaApplyConfiguration: &applyMetav1.ObjectMetaApplyConfiguration{
			Name:      pointer.String("zone-ssl-config"),
			Namespace: pointer.String(namespace),
		},
		Data: data,
	}
}

func getGlobalRole(role string) string {
	m := map[string]string{
		"owner":  "platform-admin",
		"admin":  "platform-admin",
		"normal": "workspaces-manager",
	}
	return m[role]
}
