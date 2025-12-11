package handlers

import (
	"fmt"
	"testing"
	"time"

	sysv1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
)

func TestWeekly(t *testing.T) {
	var bp = sysv1.BackupPolicy{
		SnapshotFrequency: "@daily",
		TimesOfDay:        "14:06",
	}

	var res = GetNextBackupTime(bp)
	fmt.Println(*res)
}

func TestS(t *testing.T) {
	var a int64 = 1744777260
	var b = time.Unix(a, 0)
	fmt.Println(b.String())
}
