package v2alpha1

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	providerv2alpha1 "bytetrade.io/web3os/system-server/pkg/providerregistry/v2alpha1"
	"github.com/brancz/kube-rbac-proxy/pkg/authz"
	rbacv1 "k8s.io/api/rbac/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/endpoints/request"
	versionedinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	rbacregistryvalidation "k8s.io/kubernetes/pkg/registry/rbac/validation"
	"k8s.io/kubernetes/plugin/pkg/auth/authorizer/rbac"
)

type nonResourceWithServiceRuleResolver struct {
	clusterRoleGetter        rbacregistryvalidation.ClusterRoleGetter
	clusterRoleBindingLister rbacregistryvalidation.ClusterRoleBindingLister
}

var _ Authorizer = &nonResourceWithServiceRBACAuthorizor{}

type nonResourceWithServiceRBACAuthorizor struct {
	resolver *nonResourceWithServiceRuleResolver
}

func newNonResourceWithServiceRBACAuthorizor(informerFactory versionedinformers.SharedInformerFactory) *nonResourceWithServiceRBACAuthorizor {
	return &nonResourceWithServiceRBACAuthorizor{
		&nonResourceWithServiceRuleResolver{
			&rbac.ClusterRoleGetter{Lister: informerFactory.Rbac().V1().ClusterRoles().Lister()},
			&rbac.ClusterRoleBindingLister{Lister: informerFactory.Rbac().V1().ClusterRoleBindings().Lister()},
		},
	}
}

type authorizingVisitor struct {
	requestAttributes authorizer.Attributes

	allowed bool
	reason  string
	errors  []error
	service string
}

func (v *authorizingVisitor) visit(source fmt.Stringer, role *rbacv1.ClusterRole, rule *rbacv1.PolicyRule, err error) bool {
	if rule != nil {
		v.allowed = rbac.RuleAllows(v.requestAttributes, rule)
		if v.allowed {
			v.reason = fmt.Sprintf("RBAC: allowed by %s", source.String())
			if service, ok := role.Annotations[providerv2alpha1.ProviderServiceAnnotation]; ok {
				v.service = service
			}

			klog.V(5).Infof("RBAC: allowed by %s with service %q", source.String(), v.service)
			return false
		} else {
			v.reason = fmt.Sprintf("RBAC: denied by %s", source.String())
			klog.V(5).Infof("RBAC: denied by %s, [%s]", source.String(), v.requestAttributes.GetPath())
		}
	}

	// it's not a cluster role of a provider, we have no opinion
	if err != nil {
		v.errors = append(v.errors, err)
	}
	return true
}

func (r *nonResourceWithServiceRBACAuthorizor) Authorize(ctx context.Context, requestAttributes authorizer.Attributes) (string, authorizer.Decision, string, error) {
	ruleCheckingVisitor := &authorizingVisitor{requestAttributes: requestAttributes}

	r.resolver.VisitRulesFor(ctx, requestAttributes.GetUser(), requestAttributes.GetResource(), ruleCheckingVisitor.visit)
	if ruleCheckingVisitor.allowed {
		return ruleCheckingVisitor.service, authorizer.DecisionAllow, ruleCheckingVisitor.reason, nil
	}

	// Build a detailed log of the denial.
	// Make the whole block conditional so we don't do a lot of string-building we won't use.
	if klogV := klog.V(5); klogV.Enabled() {
		var operation string
		if requestAttributes.IsResourceRequest() {
			b := &bytes.Buffer{}
			b.WriteString(`"`)
			b.WriteString(requestAttributes.GetVerb())
			b.WriteString(`" resource "`)
			b.WriteString(requestAttributes.GetResource())
			if len(requestAttributes.GetAPIGroup()) > 0 {
				b.WriteString(`.`)
				b.WriteString(requestAttributes.GetAPIGroup())
			}
			if len(requestAttributes.GetSubresource()) > 0 {
				b.WriteString(`/`)
				b.WriteString(requestAttributes.GetSubresource())
			}
			b.WriteString(`"`)
			if len(requestAttributes.GetName()) > 0 {
				b.WriteString(` named "`)
				b.WriteString(requestAttributes.GetName())
				b.WriteString(`"`)
			}
			operation = b.String()
		} else {
			operation = fmt.Sprintf("%q nonResourceURL %q provider %q", requestAttributes.GetVerb(), requestAttributes.GetPath(), requestAttributes.GetResource())
		}

		var scope string
		if ns := requestAttributes.GetNamespace(); len(ns) > 0 {
			scope = fmt.Sprintf("in namespace %q", ns)
		} else {
			scope = "cluster-wide"
		}

		klogV.Infof("RBAC: no rules authorize user %q with groups %q to %s %s", requestAttributes.GetUser().GetName(), requestAttributes.GetUser().GetGroups(), operation, scope)
	}

	reason := ""
	if len(ruleCheckingVisitor.errors) > 0 {
		reason = fmt.Sprintf("RBAC: %v", utilerrors.NewAggregate(ruleCheckingVisitor.errors))
	}
	return "", authorizer.DecisionNoOpinion, reason, nil
}

func (rr *nonResourceWithServiceRuleResolver) VisitRulesFor(ctx context.Context, user user.Info, host string, visitor func(source fmt.Stringer, role *rbacv1.ClusterRole, rule *rbacv1.PolicyRule, err error) bool) {
	if clusterRoleBindings, err := rr.clusterRoleBindingLister.ListClusterRoleBindings(ctx); err != nil {
		if !visitor(nil, nil, nil, err) {
			return
		}
	} else {
		sourceDescriber := &clusterRoleBindingDescriber{}
		for _, clusterRoleBinding := range clusterRoleBindings {
			subjectIndex, applies := appliesTo(user, clusterRoleBinding.Subjects, "")
			if !applies {
				continue
			}
			role, rules, err := rr.GetRoleReferenceRules(ctx, clusterRoleBinding.RoleRef, host)
			if err != nil {
				if !visitor(nil, nil, nil, err) {
					return
				}
				continue
			}
			sourceDescriber.binding = clusterRoleBinding
			sourceDescriber.subject = &clusterRoleBinding.Subjects[subjectIndex]
			for i := range rules {
				if !visitor(sourceDescriber, role, &rules[i], nil) {
					return
				}
			}
		}
	}
}

// GetRoleReferenceRules attempts to resolve the RoleBinding or ClusterRoleBinding.
func (rr *nonResourceWithServiceRuleResolver) GetRoleReferenceRules(ctx context.Context, roleRef rbacv1.RoleRef, host string) (*rbacv1.ClusterRole, []rbacv1.PolicyRule, error) {
	switch roleRef.Kind {
	case "ClusterRole":
		clusterRole, err := rr.clusterRoleGetter.GetClusterRole(ctx, roleRef.Name)
		if err != nil {
			klog.Error("failed to get cluster role: ", err)
			return nil, nil, err
		}

		bindingProvider := providerv2alpha1.ProviderRefFromHost(host)
		if annotation, ok := clusterRole.Annotations[providerv2alpha1.ProviderRefAnnotation]; ok {
			if annotation == bindingProvider {
				klog.V(5).Info("it's a cluster role of a provider, annotation: ", annotation)
				return clusterRole, clusterRole.Rules, nil
			}
		}

		if annotation, ok := clusterRole.Annotations[providerv2alpha1.ProviderServiceAnnotation]; ok {
			if annotation == host {
				klog.Info("it's a cluster role of a provider service, annotation: ", annotation)
				return clusterRole, clusterRole.Rules, nil
			}
		}

		return nil, nil, fmt.Errorf("[nonResourceWithServiceRuleResolver] cluster role %q does not match binding provider %q", roleRef.Name, bindingProvider)

	default:
		return nil, nil, fmt.Errorf("[nonResourceWithServiceRuleResolver] unsupported role reference kind: %q", roleRef.Kind)
	}
}

func UnionAllAuthorizers(ctx context.Context, cfg *authz.Config, kubeClient kubernetes.Interface, informerFactory versionedinformers.SharedInformerFactory) (Authorizer, error) {
	// sarClient := kubeClient.AuthorizationV1()
	// sarAuthorizer, err := authz.NewSarAuthorizer(sarClient)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create sar authorizer: %w", err)
	// }

	staticAuthorizer, err := authz.NewStaticAuthorizer(cfg.Static)
	if err != nil {
		return nil, fmt.Errorf("failed to create static authorizer: %w", err)
	}

	nonResourceWithServiceRBACAuthorizor := newNonResourceWithServiceRBACAuthorizor(informerFactory)

	authorizer := unionAuthzHandler{
		wrapAuthorizer{staticAuthorizer},
		nonResourceWithServiceRBACAuthorizor,
		// wrapAuthorizer{sarAuthorizer},
	}

	return authorizer, nil
}

type key struct{}

var providerServiceKey = key{}

func WithProviderService(parent context.Context, service string) context.Context {
	return request.WithValue(parent, providerServiceKey, service)
}

func ProviderServiceFrom(ctx context.Context) (string, bool) {
	service, ok := ctx.Value(providerServiceKey).(string)
	if !ok {
		return "", false
	}
	return service, true
}

type Authorizer interface {
	Authorize(ctx context.Context, a authorizer.Attributes) (service string, authorized authorizer.Decision, reason string, err error)
}

type unionAuthzHandler []Authorizer

// Authorizes against a chain of authorizer.Authorizer objects and returns nil if successful and returns error if unsuccessful
func (authzHandler unionAuthzHandler) Authorize(ctx context.Context, a authorizer.Attributes) (string, authorizer.Decision, string, error) {
	var (
		errlist    []error
		reasonlist []string
	)

	for _, currAuthzHandler := range authzHandler {
		service, decision, reason, err := currAuthzHandler.Authorize(ctx, a)

		if err != nil {
			errlist = append(errlist, err)
		}
		if len(reason) != 0 {
			reasonlist = append(reasonlist, reason)
		}
		switch decision {
		case authorizer.DecisionAllow, authorizer.DecisionDeny:
			return service, decision, reason, err
		case authorizer.DecisionNoOpinion:
			// continue to the next authorizer
		}
	}

	return "", authorizer.DecisionNoOpinion, strings.Join(reasonlist, "\n"), utilerrors.NewAggregate(errlist)
}

type wrapAuthorizer struct {
	authorizer authorizer.Authorizer
}

func (t wrapAuthorizer) Authorize(ctx context.Context, a authorizer.Attributes) (service string, authorized authorizer.Decision, reason string, err error) {
	authorized, reason, err = t.authorizer.Authorize(ctx, a)
	return
}
