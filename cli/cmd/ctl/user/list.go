package user

import (
	"context"
	"encoding/json"
	"fmt"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"log"
	"slices"
	"sort"
	"strings"
)

var sortFuncs = map[string]func(users []iamv1alpha2.User, i, j int) bool{
	"name": func(users []iamv1alpha2.User, i, j int) bool {
		return strings.Compare(users[i].Name, users[j].Name) == -1
	},
	"role": func(users []iamv1alpha2.User, i, j int) bool {
		return strings.Compare(users[i].Annotations[annotationKeyRole], users[j].Annotations[annotationKeyRole]) == -1
	},
	"create-time": func(users []iamv1alpha2.User, i, j int) bool {
		return users[i].CreationTimestamp.Before(&users[j].CreationTimestamp)
	},
	"memory": func(users []iamv1alpha2.User, i, j int) bool {
		iMemoryStr, ok := users[i].Annotations[annotationKeyMemoryLimit]
		if !ok || iMemoryStr == "" {
			return false
		}
		jMemoryStr, ok := users[j].Annotations[annotationKeyMemoryLimit]
		if !ok || jMemoryStr == "" {
			return true
		}
		iMemory, err := resource.ParseQuantity(iMemoryStr)
		if err != nil {
			fmt.Printf("Warning: invalid memory limit '%s' is set on user '%s'\n", iMemoryStr, users[i].Name)
			return false
		}
		jMemory, err := resource.ParseQuantity(jMemoryStr)
		if err != nil {
			fmt.Printf("Warning: invalid memory limit '%s' is set on user '%s'\n", jMemoryStr, users[j].Name)
			return true
		}
		return iMemory.Cmp(jMemory) == -1
	},
	"cpu": func(users []iamv1alpha2.User, i, j int) bool {
		iCPUStr, ok := users[i].Annotations[annotationKeyCPULimit]
		if !ok || iCPUStr == "" {
			return false
		}
		jCPUStr, ok := users[j].Annotations[annotationKeyCPULimit]
		if !ok || jCPUStr == "" {
			return true
		}
		iCPU, err := resource.ParseQuantity(iCPUStr)
		if err != nil {
			fmt.Printf("Warning: invalid cpu limit '%s' is set on user '%s'", iCPUStr, users[i].Name)
			return false
		}
		jCPU, err := resource.ParseQuantity(jCPUStr)
		if err != nil {
			fmt.Printf("Warning: invalid cpu limit '%s' is set on user '%s'", jCPUStr, users[j].Name)
			return true
		}
		return iCPU.Cmp(jCPU) == -1
	},
}

var sortAliases = map[string]sets.Set[string]{
	"name":        sets.New[string]("n", "N", "Name"),
	"role":        sets.New[string]("r", "R", "Role"),
	"create-time": sets.New[string]("creation", "created", "created-at", "createdat", "createtime"),
	"cpu":         sets.New[string]("c", "C", "CPU"),
	"memory":      sets.New[string]("m", "M", "Memory"),
}

func getSortFunc(sortBy string) func(users []iamv1alpha2.User, i, j int) bool {
	if f, ok := sortFuncs[sortBy]; ok {
		return f
	}
	for origin, sortAlias := range sortAliases {
		if sortAlias.Has(sortBy) {
			return sortFuncs[origin]
		}
	}
	return nil
}

type listUsersOptions struct {
	kubeConfig string
	output     string
	noHeaders  bool
	sortBys    []string
	reverse    bool
}

func NewCmdListUsers() *cobra.Command {
	o := &listUsersOptions{}
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "list all users",
		Run: func(cmd *cobra.Command, args []string) {
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

func (o *listUsersOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.kubeConfig, "kubeconfig", "", "path to kubeconfig file")
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", "output format (table, json)")
	cmd.Flags().BoolVar(&o.noHeaders, "no-headers", false, "disable headers")
	cmd.Flags().StringSliceVar(&o.sortBys, "sort", []string{}, "sort output order by (name, role, create-time, memory, cpu)")
	cmd.Flags().BoolVarP(&o.reverse, "reverse", "r", false, "reverse order")
}

func (o *listUsersOptions) Validate() error {
	for _, sortBy := range o.sortBys {
		f := getSortFunc(sortBy)
		if f == nil {
			return fmt.Errorf("unknown sort option: %s", sortBy)
		}
	}
	return nil
}

func (o *listUsersOptions) Run() error {
	ctx := context.Background()

	userClient, err := newUserClientFromKubeConfig(o.kubeConfig)
	if err != nil {
		return err
	}

	var userList iamv1alpha2.UserList
	err = userClient.List(ctx, &userList)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	for _, sortBy := range o.sortBys {
		sort.SliceStable(userList.Items, func(i, j int) bool {
			f := getSortFunc(sortBy)
			if f == nil {
				log.Fatalf("unkown sort option: %s", sortBy)
			}
			return f(userList.Items, i, j)
		})
	}

	if o.reverse {
		slices.Reverse(userList.Items)
	}

	users := make([]userInfo, 0, len(userList.Items))
	for _, user := range userList.Items {
		users = append(users, convertUserObjectToUserInfo(user))
	}

	if o.output == "json" {
		jsonOutput, _ := json.MarshalIndent(users, "", "  ")
		fmt.Println(string(jsonOutput))
	} else {
		if !o.noHeaders {
			printUserTableHeaders()
		}
		for _, user := range users {
			printUserTableRow(user)
		}
	}

	return nil
}
