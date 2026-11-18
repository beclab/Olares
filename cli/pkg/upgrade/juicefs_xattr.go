package upgrade

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/core/prepare"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/storage"
)

func juiceFSXattrTasks() []task.Interface {
	return []task.Interface{&task.LocalTask{
		Name:    "EnableJuiceFsService",
		Prepare: &prepare.FileExist{FilePath: storage.JuiceFsServiceFile},
		Action:  &storage.EnableJuiceFsService{ConfigOnly: true},
		Retry:   3,
		Delay:   5 * time.Second,
	}}
}
