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

type deleteUserOptions struct {
	name       string
	kubeConfig string
}

func NewCmdDeleteUser() *cobra.Command {
	o := &deleteUserOptions{}
	cmd := &cobra.Command{
		Use:     "delete {name}",
		Short:   "delete an existing user",
		Aliases: []string{"d", "del", "rm", "remove"},
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

func (o *deleteUserOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.kubeConfig, "kubeconfig", "", "path to kubeconfig file")
}

func (o *deleteUserOptions) Validate() error {
	if o.name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func (o *deleteUserOptions) Run() error {
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

	if user.Status.State == "Creating" {
		return fmt.Errorf("user '%s' is under creation", o.name)
	}

	if role, ok := user.Annotations[annotationKeyRole]; ok && role == roleOwner {
		return fmt.Errorf("cannot delete user '%s' with role '%s' ", o.name, role)
	}

	err = userClient.Delete(ctx, &user)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	fmt.Printf("User '%s' deleted successfully\n", o.name)
	return nil
}
