package user

import (
	"context"
	"fmt"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"log"
)

type updateUserLimitsOptions struct {
	name string
	resourceLimit
	kubeConfig string
}

func NewCmdUpdateUserLimits() *cobra.Command {
	o := &updateUserLimitsOptions{}
	cmd := &cobra.Command{
		Use:     "update-limits {name}",
		Aliases: []string{"update-limit", "ulimit", "ulimits"},
		Short:   "update user resource limits",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			o.name = args[0]
			if err := o.Validate(); err != nil {
				log.Fatal(err)
			}
			if err := o.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func (o *updateUserLimitsOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.memoryLimit, "memory-limit", "m", "", "memory limit")
	cmd.Flags().StringVarP(&o.cpuLimit, "cpu-limit", "c", "", "cpu limit")
	cmd.Flags().StringVar(&o.kubeConfig, "kubeconfig", "", "path to kubeconfig file")
}

func (o *updateUserLimitsOptions) Validate() error {
	if o.name == "" {
		return fmt.Errorf("user name is required")
	}

	if o.memoryLimit == "" && o.cpuLimit == "" {
		return fmt.Errorf("one of memory limit or cpu limit is required")
	}

	if err := validateResourceLimit(o.resourceLimit); err != nil {
		return err
	}

	return nil
}

func (o *updateUserLimitsOptions) Run() error {
	ctx := context.Background()

	userClient, err := newUserClientFromKubeConfig(o.kubeConfig)
	if err != nil {
		return err
	}

	var user iamv1alpha2.User
	err = userClient.Get(ctx, types.NamespacedName{Name: o.name}, &user)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("user '%s' not found", o.name)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.Annotations == nil {
		user.Annotations = make(map[string]string)
	}

	if o.memoryLimit != "" {
		user.Annotations[annotationKeyMemoryLimit] = o.memoryLimit
	}
	if o.cpuLimit != "" {
		user.Annotations[annotationKeyCPULimit] = o.cpuLimit
	}

	err = userClient.Update(ctx, &user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	fmt.Printf("User '%s' resource limits updated successfully\n", o.name)
	return nil
}
