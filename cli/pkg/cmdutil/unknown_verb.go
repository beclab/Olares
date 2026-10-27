package cmdutil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type unknownVerbError struct {
	message string
}

func (e *unknownVerbError) Error() string {
	return e.message
}

func IsUnknownVerb(err error) bool {
	var target *unknownVerbError
	return errors.As(err, &target)
}

func RefuseUnknownVerbArgs(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	if args[0] == "help" {
		return nil
	}
	return unknownVerb(c, args[0])
}

func RefuseUnknownVerbGroupArgs(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return pflag.ErrHelp
	}
	return RefuseUnknownVerbArgs(c, args)
}

// RefuseUnknownVerb makes a command group runnable so Cobra cannot turn an
// unresolved word into flag.ErrHelp and then a successful exit.
func RefuseUnknownVerb(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return c.Help()
	}
	if args[0] == "help" {
		if len(args) > 1 {
			if child, _, err := c.Find(args[1:]); err == nil && child != c {
				return child.Help()
			}
		}
		return c.Help()
	}
	return unknownVerb(c, args[0])
}

const flagHintAnnotationPrefix = "olares-cli/flag-hint/"

// ExplainMissingFlag attaches the answer to a flag callers reach for on a
// command that deliberately does not have it, so the refusal says why
// instead of only that the flag is unknown.
func ExplainMissingFlag(c *cobra.Command, explanation string, names ...string) {
	if c.Annotations == nil {
		c.Annotations = make(map[string]string)
	}
	for _, name := range names {
		c.Annotations[flagHintAnnotationPrefix+name] = explanation
	}
}

// FlagErrorWithContext answers a flag the command does not have. In a command
// group the mistake is the word before the flag: the verb was not found, so
// the group parsed flags that belong to no verb. On a leaf it names the
// command whose --help lists the flags it does have.
func FlagErrorWithContext(c *cobra.Command, err error) error {
	var notExist *pflag.NotExistError
	if !errors.As(err, &notExist) {
		return err
	}
	if c.HasSubCommands() {
		if args := c.Flags().Args(); len(args) > 0 && args[0] != "help" {
			return unknownVerb(c, args[0])
		}
	}
	if hint := c.Annotations[flagHintAnnotationPrefix+notExist.GetSpecifiedName()]; hint != "" {
		return fmt.Errorf("%w: %s", err, hint)
	}
	return fmt.Errorf("%w\n`%s --help` lists the flags it has", err, c.CommandPath())
}

func unknownVerb(c *cobra.Command, verb string) error {
	if c.SuggestionsMinimumDistance <= 0 {
		c.SuggestionsMinimumDistance = 2
	}
	msg := fmt.Sprintf("unknown verb %q for %q", verb, c.CommandPath())
	if suggestions := c.SuggestionsFor(verb); len(suggestions) > 0 {
		msg += "\nDid you mean: " + strings.Join(suggestions, ", ")
	}
	return &unknownVerbError{
		message: fmt.Sprintf("%s\n`%s --help` lists the verbs it has", msg, c.CommandPath()),
	}
}
