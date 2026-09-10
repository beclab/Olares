package cmdutil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
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
	if c.SuggestionsMinimumDistance <= 0 {
		c.SuggestionsMinimumDistance = 2
	}
	msg := fmt.Sprintf("unknown verb %q for %q", args[0], c.CommandPath())
	if suggestions := c.SuggestionsFor(args[0]); len(suggestions) > 0 {
		msg += "\nDid you mean: " + strings.Join(suggestions, ", ")
	}
	return &unknownVerbError{
		message: fmt.Sprintf("%s\n`%s --help` lists the verbs it has", msg, c.CommandPath()),
	}
}
