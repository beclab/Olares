package pipelines

import (
	"fmt"

	"github.com/beclab/Olares/cli/pkg/terminus"
)

func PrintTerminusInfo() {
	var cli = &terminus.GetOlaresVersion{}
	terminusVersion, err := cli.Execute()
	if err != nil {
		fmt.Printf("Olares: not installed\n")
		return
	}

	fmt.Printf("Olares: %s\n", terminusVersion)
}
