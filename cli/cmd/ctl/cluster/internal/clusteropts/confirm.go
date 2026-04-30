package clusteropts

import "github.com/beclab/Olares/cli/pkg/cliutil"

// ConfirmDestructive is a re-export of cliutil.ConfirmDestructive
// kept under clusteropts so the existing cluster verbs that call
// `clusteropts.ConfirmDestructive(...)` don't have to re-import the
// shared package directly. New cluster code is welcome to import
// cliutil straight away — both paths reach the same function.
var ConfirmDestructive = cliutil.ConfirmDestructive
