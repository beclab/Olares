package watchers

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/robfig/cron/v3"
	"olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/util/log"
)

func InitCrontabs() {
	c := cron.New()

	_, err := c.AddFunc("10 0 * * *", func() {
		clearHistoryRestores()
	})
	if err != nil {
		log.Errorf("AddFunc clearHistoryRestores err:%v", err)
	} else {
		log.Info("Crontab task: clearHistoryRestores added successfully.")
	}

	c.Start()
}

func clearHistoryRestores() {
	log.Info("Start clear restores...")
	f, err := client.NewFactory()
	if err != nil {
		log.Errorf("init client error: %v", err)
		return
	}

	c, err := f.Sysv1Client()
	if err != nil {
		log.Errorf("init sys client error: %v", err)
		return
	}

	restores, err := c.SysV1().Restores(constant.DefaultNamespaceOsFramework).List(context.Background(), metav1.ListOptions{})
	if restores == nil || restores.Items == nil || len(restores.Items) == 0 {
		log.Info("restores not exists")
		return
	}

	for _, item := range restores.Items {
		if !metav1.Now().Time.AddDate(0, 0, -15).After(item.CreationTimestamp.Time) {
			continue
		}

		if err := c.SysV1().Restores(constant.DefaultNamespaceOsFramework).Delete(context.Background(), item.Name, metav1.DeleteOptions{}); err != nil {
			log.Errorf("delete restore error: %v", err)
			continue
		} else {
			log.Infof("remove restore %s, %s", item.Name, item.Spec.Owner)
		}
	}
}
