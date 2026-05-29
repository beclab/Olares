package workload

import (
	"encoding/json"
	"testing"
)

func TestCollectBatchImageRefsReadsJobTemplate(t *testing.T) {
	const raw = `{
		"metadata": {"name": "backup", "namespace": "user-space"},
		"spec": {
			"template": {
				"spec": {
					"initContainers": [{"name": "prep", "image": "busybox:1.36"}],
					"containers": [{"name": "main", "image": "docker.io/example/backup:v2"}]
				}
			}
		}
	}`
	var job batchWorkload
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		t.Fatalf("unmarshal job: %v", err)
	}

	refs := collectBatchImageRefs("jobs", []batchWorkload{job})

	if len(refs) != 2 {
		t.Fatalf("refs len = %d, want 2: %#v", len(refs), refs)
	}
	if got, want := refs[0].Image, "busybox:1.36"; got != want {
		t.Fatalf("init image = %q, want %q", got, want)
	}
	if got, want := refs[0].Kind, "Job"; got != want {
		t.Fatalf("kind = %q, want %q", got, want)
	}
	if got, want := refs[1].Image, "docker.io/example/backup:v2"; got != want {
		t.Fatalf("container image = %q, want %q", got, want)
	}
}

func TestCollectBatchImageRefsReadsCronJobJobTemplate(t *testing.T) {
	const raw = `{
		"metadata": {"name": "rotate", "namespace": "user-space"},
		"spec": {
			"jobTemplate": {
				"spec": {
					"template": {
						"spec": {
							"containers": [{"name": "rotate", "image": "docker.io/example/rotate:v1"}]
						}
					}
				}
			}
		}
	}`
	var cron batchWorkload
	if err := json.Unmarshal([]byte(raw), &cron); err != nil {
		t.Fatalf("unmarshal cronjob: %v", err)
	}

	refs := collectBatchImageRefs("cronjobs", []batchWorkload{cron})

	if len(refs) != 1 {
		t.Fatalf("refs len = %d, want 1: %#v", len(refs), refs)
	}
	if got, want := refs[0].Image, "docker.io/example/rotate:v1"; got != want {
		t.Fatalf("image = %q, want %q", got, want)
	}
	if got, want := refs[0].Kind, "CronJob"; got != want {
		t.Fatalf("kind = %q, want %q", got, want)
	}
}
