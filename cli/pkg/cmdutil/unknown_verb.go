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
