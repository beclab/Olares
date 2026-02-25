package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/common"
)

// Init initializes viper as the single source of truth:
// 1) Load /etc/olares/release (dotenv) into the process env if present
// 2) Enable viper to read environment variables
// 3) Bind environment variables for all known keys we care about
func Init() {
	godotenv.Load(common.OlaresReleaseFile)
	viper.SetEnvPrefix("OLARES")
	viper.SetEnvKeyReplacer(envKeyReplacer)
	viper.AutomaticEnv()
}

var aliasToFlag = map[string]string{}
var envToFlag = map[string]string{}
var envKeyReplacer = strings.NewReplacer("-", "_")

type CommandFlagSetter interface {
	Add(flag string, short string, defValue any, description string) CommandFlagItem
}

type CommandFlagItem interface {
	WithAlias(aliases ...string) CommandFlagItem
	WithEnv(envs ...string) CommandFlagItem
}

func NewFlagSetterFor(cmd *cobra.Command) CommandFlagSetter {
	if cmd == nil {
		panic(fmt.Errorf("command is nil"))
	}
	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		if f, ok := aliasToFlag[name]; ok {
			return pflag.NormalizedName(f)
		}
		return pflag.NormalizedName(name)
	})
	return &commandFlagSetterImpl{
		command: cmd,
	}
}

type commandFlagSetterImpl struct {
	command *cobra.Command
}

type commandFlagItemImpl struct {
	command *cobra.Command
	flag    string
}

func (c *commandFlagSetterImpl) Add(flag string, short string, defValue any, description string) CommandFlagItem {
	switch reflect.TypeOf(defValue).Kind() {
	case reflect.Bool:
		c.command.Flags().BoolP(flag, short, defValue.(bool), description)
	case reflect.String:
		c.command.Flags().StringP(flag, short, defValue.(string), description)
	case reflect.Int:
		c.command.Flags().IntP(flag, short, defValue.(int), description)
	}
	viper.BindPFlag(flag, c.command.Flags().Lookup(flag))

	// transitional support for legacy envs without prefix
	// it should be removed after all envs are migrated
	viper.BindEnv(flag, strings.ToUpper(envKeyReplacer.Replace(flag)))
	return &commandFlagItemImpl{
		flag:    flag,
		command: c.command,
	}
}

func (c *commandFlagItemImpl) WithAlias(aliases ...string) CommandFlagItem {
	for _, a := range aliases {
		if f, ok := aliasToFlag[a]; ok {
			if f != c.flag {
				panic(fmt.Errorf("flag alias %s already exists for flag %s, please use a different alias", a, f))
			}
			continue
		}
		viper.BindEnv(c.flag, a)
		aliasToFlag[a] = c.flag
	}
	return c
}

func (c *commandFlagItemImpl) WithEnv(envs ...string) CommandFlagItem {
	for _, e := range envs {
		if f, ok := envToFlag[e]; ok {
			if f != c.flag {
				panic(fmt.Errorf("env %s already exists for flag %s, please use a different env", e, f))
			}
			continue
		}
		viper.BindEnv(c.flag, e)
		envToFlag[e] = c.flag
	}
	return c
}
