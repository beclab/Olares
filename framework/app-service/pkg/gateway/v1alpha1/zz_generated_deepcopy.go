// Hand-written deepcopy implementations (mirrors output of
// sigs.k8s.io/controller-tools deepcopy-gen). Keep in sync when adding
// fields to SharedRouteRegistry / SharedRouteRegistrySpec / *Status.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto copies the receiver into out.
func (in *UpstreamRef) DeepCopyInto(out *UpstreamRef) {
	*out = *in
}

// DeepCopy returns a deep copy of UpstreamRef.
func (in *UpstreamRef) DeepCopy() *UpstreamRef {
	if in == nil {
		return nil
	}
	out := new(UpstreamRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into out.
func (in *SharedRouteRegistrySpec) DeepCopyInto(out *SharedRouteRegistrySpec) {
	*out = *in
	if in.HostPatterns != nil {
		out.HostPatterns = make([]string, len(in.HostPatterns))
		copy(out.HostPatterns, in.HostPatterns)
	}
	in.Upstream.DeepCopyInto(&out.Upstream)
}

// DeepCopy returns a deep copy of SharedRouteRegistrySpec.
func (in *SharedRouteRegistrySpec) DeepCopy() *SharedRouteRegistrySpec {
	if in == nil {
		return nil
	}
	out := new(SharedRouteRegistrySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into out.
func (in *SharedRouteRegistryStatus) DeepCopyInto(out *SharedRouteRegistryStatus) {
	*out = *in
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

// DeepCopy returns a deep copy of SharedRouteRegistryStatus.
func (in *SharedRouteRegistryStatus) DeepCopy() *SharedRouteRegistryStatus {
	if in == nil {
		return nil
	}
	out := new(SharedRouteRegistryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into out.
func (in *SharedRouteRegistry) DeepCopyInto(out *SharedRouteRegistry) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy returns a deep copy of SharedRouteRegistry.
func (in *SharedRouteRegistry) DeepCopy() *SharedRouteRegistry {
	if in == nil {
		return nil
	}
	out := new(SharedRouteRegistry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject satisfies runtime.Object.
func (in *SharedRouteRegistry) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copies the receiver into out.
func (in *SharedRouteRegistryList) DeepCopyInto(out *SharedRouteRegistryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]SharedRouteRegistry, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

// DeepCopy returns a deep copy of SharedRouteRegistryList.
func (in *SharedRouteRegistryList) DeepCopy() *SharedRouteRegistryList {
	if in == nil {
		return nil
	}
	out := new(SharedRouteRegistryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject satisfies runtime.Object.
func (in *SharedRouteRegistryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
