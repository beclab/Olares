package user

import (
	"context"
	"encoding/json"
	"fmt"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"log"
)

type getUserOptions struct {
	name       string
	kubeConfig string
	output     string
	noHeaders  bool
}

func NewCmdGetUser() *cobra.Command {
	o := &getUserOptions{}
	cmd := &cobra.Command{
		Use:   "get {name}",
		Short: "get user details",
		Args:  cobra.ExactArgs(1),
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

func (o *getUserOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.kubeConfig, "kubeconfig", "", "path to kubeconfig file")
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", "output format (table, json)")
	cmd.Flags().BoolVar(&o.noHeaders, "no-headers", false, "disable headers")
}

func (o *getUserOptions) Validate() error {
	if o.name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func (o *getUserOptions) Run() error {
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

	info := convertUserObjectToUserInfo(user)

	if o.output == "json" {
		jsonOutput, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(jsonOutput))
	} else {
		if !o.noHeaders {
			printUserTableHeaders()
		}
		printUserTableRow(info)
	}

	return nil
}
