package runtime

import (
	"os"

	"olares.com/backup-server/pkg/util/log"
)

func Must(err error) {
	if err != nil {
		log.Errorf("%+v\n", err)
		os.Exit(1)
	}
}
