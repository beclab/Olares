package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var BackupConfigMap = template.Must(template.New("cm-backup-config.yaml").Parse(
	dedent.Dedent(`apiVersion: v1
data:
  terminus.cloudVersion: "{{ .CloudInstance }}"
  backup.clusterBucket: "{{ .StorageBucket }}"
  backup.keyPrefix: "{{ .StoragePrefix }}"
  backup.secret: "{{ .StorageSyncSecret }}"
kind: ConfigMap
metadata:
  name: backup-config
  namespace: os-system`),
))
