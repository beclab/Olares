// Package systemcomponents describes the Kubernetes workloads that make up an
// Olares installation and provides the readiness check used to decide whether
// the system is up.
//
// The check is workload based rather than pod based: it resolves every declared
// component to its Deployment, StatefulSet or DaemonSet and compares the ready
// replica counts against the desired ones. A pod based scan can only validate
// pods that already exist, so it silently accepts a component whose workload was
// never created, was scaled to zero, or has not started rolling out yet.
package systemcomponents

import (
	"fmt"
	"strings"
)

// Kind is the workload type a component is deployed as.
type Kind string

const (
	Deployment  Kind = "Deployment"
	StatefulSet Kind = "StatefulSet"
	DaemonSet   Kind = "DaemonSet"
)

// UserPlaceholder appears in the namespace of per user components and is
// replaced with each provisioned Olares user when the registry is resolved
// against a cluster.
const UserPlaceholder = "<user>"

// Presence says whether a component is expected on every installation.
type Presence int

const (
	// Required components must exist and be ready on every installation.
	// A missing workload fails the check.
	Required Presence = iota

	// Optional components depend on hardware or on install options that cannot
	// be reliably re-derived from a running cluster: the GPU vendor, whether
	// JuiceFS is used, the CNI choice, and the host OS. They are checked only
	// when the workload exists, so a cluster that legitimately does not have
	// them passes, while a broken one still fails.
	Optional
)

// Component is a single Olares system workload.
type Component struct {
	// Namespace is where the workload lives. Per user components carry
	// UserPlaceholder here and are expanded per user by Resolve.
	Namespace string

	Kind Kind

	// Name identifies the workload. It is empty when Selector is set.
	Name string

	// Selector matches the workload by labels instead of by name, for workloads
	// whose names are generated per cluster and are therefore not portable.
	Selector string

	Presence Presence
}

// ID returns a stable human readable identifier such as
// "Deployment os-framework/backup".
func (c Component) ID() string {
	target := c.Name
	if target == "" {
		target = fmt.Sprintf("[%s]", c.Selector)
	}
	return fmt.Sprintf("%s %s/%s", c.Kind, c.Namespace, target)
}

// forUser returns a copy of the component with the user placeholder in its
// namespace replaced by the given user name.
func (c Component) forUser(user string) Component {
	c.Namespace = strings.ReplaceAll(c.Namespace, UserPlaceholder, user)
	return c
}

// isPerUser reports whether the component is instantiated once per Olares user.
func (c Component) isPerUser() bool {
	return strings.Contains(c.Namespace, UserPlaceholder)
}
