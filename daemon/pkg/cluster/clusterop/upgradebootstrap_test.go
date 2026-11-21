package clusterop

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func fakeCLI(version, calls string) string {
	return `#!/bin/sh
if [ "$1" = "-v" ]; then echo "olares-cli version ` + version + `"; exit 0; fi
echo "` + version + ` $*" >> ` + calls + `
`
}

type scriptRun struct {
	dir, calls, cli, curlLog string
}

// runLegacyScript runs the Job's script with the host paths pointed into a
// scratch directory, an installed olares-cli at installed, and a CDN that
// serves an olares-cli at target.
func runLegacyScript(t *testing.T, installed, target string, env ...string) (scriptRun, string, error) {
	t.Helper()
	dir := t.TempDir()
	r := scriptRun{dir: dir, calls: filepath.Join(dir, "calls"), cli: filepath.Join(dir, "olares-cli"),
		curlLog: filepath.Join(dir, "curl")}
	envFile := filepath.Join(dir, "olaresd.service.env")
	if err := os.WriteFile(envFile, []byte("BASE_DIR='"+dir+"'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(r.cli, []byte(fakeCLI(installed, r.calls)), 0755); err != nil {
		t.Fatal(err)
	}

	archive := filepath.Join(dir, "served.tar.gz")
	writeCLIArchive(t, archive, fakeCLI(target, r.calls))
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0755); err != nil {
		t.Fatal(err)
	}
	curl := `#!/bin/sh
out=""; url=""
while [ $# -gt 0 ]; do
  case "$1" in -o) out=$2; shift 2 ;; -*) shift ;; *) url=$1; shift ;; esac
done
echo "$url" >> ` + r.curlLog + `
cp ` + archive + ` "$out"
`
	if err := os.WriteFile(filepath.Join(bin, "curl"), []byte(curl), 0755); err != nil {
		t.Fatal(err)
	}

	script := strings.NewReplacer(
		"/etc/systemd/system/olaresd.service.env", envFile,
		"/usr/local/bin/olares-cli", r.cli,
	).Replace(legacyBootstrapScript)
	cmd := exec.Command("sh", "-ec", script)
	cmd.Env = append(os.Environ(), "PATH="+bin+":"+os.Getenv("PATH"), "UPGRADE_VERSION=1.12.10",
		"UPGRADE_CDN=https://cdn.example/", "CLI_URL=", "WIZARD_URL=")
	cmd.Env = append(cmd.Env, env...)
	out, err := cmd.CombinedOutput()
	return r, string(out), err
}

func writeCLIArchive(t *testing.T, path, contents string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "olares-cli", Mode: 0755, Size: int64(len(contents))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(contents)); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(b)
}

func TestLegacyBootstrapInstallsTheTargetCLIAndRunsItsBootstrapNode(t *testing.T) {
	r, out, err := runLegacyScript(t, "1.12.8", "1.12.10", "WIZARD_URL=https://mirror/wizard.tar.gz")
	if err != nil {
		t.Fatalf("script: %v\n%s", err, out)
	}
	arch := map[string]string{"x86_64": "amd64", "amd64": "amd64", "aarch64": "arm64", "arm64": "arm64"}
	uname, _ := exec.Command("uname", "-m").Output()
	wantURL := "https://cdn.example/olares-cli-v1.12.10_linux_" + arch[strings.TrimSpace(string(uname))] + ".tar.gz"
	if got := strings.TrimSpace(readFile(t, r.curlLog)); got != wantURL {
		t.Errorf("downloaded %q, want %q", got, wantURL)
	}
	want := "1.12.10 upgrade bootstrap-node --version 1.12.10 --base-dir " + r.dir +
		" --cdn-service https://cdn.example/ --url-override https://mirror/wizard.tar.gz"
	if got := strings.TrimSpace(readFile(t, r.calls)); got != want {
		t.Errorf("ran %q, want %q", got, want)
	}
}

func TestLegacyBootstrapUsesTheOperatorsCLIURL(t *testing.T) {
	r, out, err := runLegacyScript(t, "1.12.8", "1.12.10", "CLI_URL=https://mirror/cli.tar.gz")
	if err != nil {
		t.Fatalf("script: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(readFile(t, r.curlLog)); got != "https://mirror/cli.tar.gz" {
		t.Errorf("downloaded %q", got)
	}
}

// A Job that ran before olaresd restarted already installed the CLI.
func TestLegacyBootstrapDoesNotDownloadACLIAlreadyAtTheTarget(t *testing.T) {
	r, out, err := runLegacyScript(t, "1.12.10", "1.12.10")
	if err != nil {
		t.Fatalf("script: %v\n%s", err, out)
	}
	if got := readFile(t, r.curlLog); got != "" {
		t.Errorf("downloaded %q for a CLI already at the target", got)
	}
	if !strings.HasPrefix(readFile(t, r.calls), "1.12.10 upgrade bootstrap-node") {
		t.Errorf("bootstrap-node was not run: %q", readFile(t, r.calls))
	}
}

func TestLegacyBootstrapRefusesACLIOfAnotherVersion(t *testing.T) {
	r, _, err := runLegacyScript(t, "1.12.8", "1.12.9")
	if err == nil {
		t.Fatal("installed a CLI that is not the target")
	}
	if readFile(t, r.calls) != "" {
		t.Errorf("ran %q with the wrong CLI", readFile(t, r.calls))
	}
	if got, _ := exec.Command(r.cli, "-v").Output(); !strings.Contains(string(got), "1.12.8") {
		t.Errorf("the installed CLI was replaced: %s", got)
	}
}

// bootstrapCluster is a fake cluster whose Job controller is the function
// given: it decides, for each Job created, the pod and status it ends with.
func bootstrapCluster(t *testing.T, outcome func(job *batchv1.Job) (corev1.PodStatus, batchv1.JobStatus)) (*fake.Clientset, *[]string) {
	t.Helper()
	prev := legacyBootstrapPoll
	legacyBootstrapPoll = time.Millisecond
	t.Cleanup(func() { legacyBootstrapPoll = prev })

	client := fake.NewSimpleClientset()
	var mu sync.Mutex
	var created []string
	client.PrependReactor("create", "jobs", func(action ktesting.Action) (bool, runtime.Object, error) {
		job := action.(ktesting.CreateAction).GetObject().(*batchv1.Job)
		podStatus, jobStatus := outcome(job)
		job.Status = jobStatus
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: job.Name + "-pod", Namespace: job.Namespace,
				Labels: map[string]string{"job-name": job.Name}},
			Status: podStatus,
		}
		if err := client.Tracker().Add(pod); err != nil {
			t.Error(err)
		}
		mu.Lock()
		created = append(created, job.Spec.Template.Spec.Containers[0].Image)
		mu.Unlock()
		return false, nil, nil
	})
	return client, &created
}

func pulling(reason string) corev1.PodStatus {
	return corev1.PodStatus{Phase: corev1.PodPending, ContainerStatuses: []corev1.ContainerStatus{{
		Name: "bootstrap", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: reason}},
	}}}
}

var worker1 = inventory.Node{NodeName: "worker-1", Role: inventory.RoleWorker}

func TestLegacyBootstrapJobIsPinnedToTheNodeAndCanReachTheHost(t *testing.T) {
	job := legacyBootstrapJob("j", "worker-1", legacyBootstrapImages[0], "1.12.10",
		UpgradeLocations{CDN: "https://cdn.example", CliURL: "https://mirror/cli", WizardURL: "https://mirror/wizard"})
	spec := job.Spec.Template.Spec
	if spec.NodeName != "worker-1" || !spec.HostPID {
		t.Fatalf("job is not pinned to the node in its PID namespace: %+v", spec)
	}
	c := spec.Containers[0]
	if c.SecurityContext == nil || c.SecurityContext.Privileged == nil || !*c.SecurityContext.Privileged {
		t.Fatal("bootstrap cannot enter the host filesystem")
	}
	for name, want := range map[string]string{"UPGRADE_VERSION": "1.12.10", "UPGRADE_CDN": "https://cdn.example",
		"CLI_URL": "https://mirror/cli", "WIZARD_URL": "https://mirror/wizard"} {
		if got := jobEnv(job, name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestLegacyBootstrapTriesTheNextImageWhenANodeCannotPullOne(t *testing.T) {
	client, created := bootstrapCluster(t, func(job *batchv1.Job) (corev1.PodStatus, batchv1.JobStatus) {
		switch job.Spec.Template.Spec.Containers[0].Image {
		case legacyBootstrapImages[0]:
			return pulling("ErrImagePull"), batchv1.JobStatus{}
		case legacyBootstrapImages[1]:
			return pulling("ImagePullBackOff"), batchv1.JobStatus{}
		}
		return corev1.PodStatus{Phase: corev1.PodSucceeded}, batchv1.JobStatus{Succeeded: 1}
	})
	if err := bootstrapLegacyUpgradeWorker(context.Background(), client, worker1, "op-1", "1.12.10", UpgradeLocations{}); err != nil {
		t.Fatal(err)
	}
	if len(*created) != 3 || (*created)[2] != legacyBootstrapImages[2] {
		t.Fatalf("created jobs with %v", *created)
	}
	jobs, _ := client.BatchV1().Jobs(upgradeSecretNamespace).List(context.Background(), metav1.ListOptions{})
	if len(jobs.Items) != 1 {
		t.Errorf("the Jobs stuck pulling were left behind: %d Jobs", len(jobs.Items))
	}
}

func TestLegacyBootstrapFailsWhenNoImageCanBePulled(t *testing.T) {
	client, _ := bootstrapCluster(t, func(*batchv1.Job) (corev1.PodStatus, batchv1.JobStatus) {
		return pulling("ErrImagePull"), batchv1.JobStatus{}
	})
	err := bootstrapLegacyUpgradeWorker(context.Background(), client, worker1, "op-1", "1.12.10", UpgradeLocations{})
	if err == nil || !strings.Contains(err.Error(), "none of the bootstrap images") {
		t.Fatalf("err = %v", err)
	}
}

// A failed Job says why in its log, and that is what the stage record shows.
func TestLegacyBootstrapFailureCarriesThePodLog(t *testing.T) {
	client, created := bootstrapCluster(t, func(*batchv1.Job) (corev1.PodStatus, batchv1.JobStatus) {
		return corev1.PodStatus{Phase: corev1.PodFailed}, batchv1.JobStatus{Failed: 1}
	})
	err := bootstrapLegacyUpgradeWorker(context.Background(), client, worker1, "op-1", "1.12.10", UpgradeLocations{})
	failed, ok := err.(*LegacyBootstrapError)
	if !ok || failed.Log == "" {
		t.Fatalf("err = %v, want the pod log", err)
	}
	if len(*created) != 1 {
		t.Errorf("a failed Job was retried in another image: %v", *created)
	}
	if got := legacyBootstrapReason(worker1, err); got != strings.TrimSpace(failed.Log) && !strings.Contains(got, strings.TrimSpace(failed.Log)) {
		t.Errorf("stage record reason %q does not carry the log %q", got, failed.Log)
	}
}

// After olaresd restarts the stage asks again, and finds the Job it started.
func TestLegacyBootstrapReusesTheJobOfTheSameOperationAndNode(t *testing.T) {
	client, created := bootstrapCluster(t, func(*batchv1.Job) (corev1.PodStatus, batchv1.JobStatus) {
		return corev1.PodStatus{Phase: corev1.PodSucceeded}, batchv1.JobStatus{Succeeded: 1}
	})
	for range 2 {
		if err := bootstrapLegacyUpgradeWorker(context.Background(), client, worker1, "op-1", "1.12.10", UpgradeLocations{}); err != nil {
			t.Fatal(err)
		}
	}
	if len(*created) != 1 {
		t.Fatalf("created %d Jobs for one operation and node", len(*created))
	}
	other := inventory.Node{NodeName: "worker-2", Role: inventory.RoleWorker}
	if err := bootstrapLegacyUpgradeWorker(context.Background(), client, other, "op-1", "1.12.10", UpgradeLocations{}); err != nil {
		t.Fatal(err)
	}
	if len(*created) != 2 {
		t.Fatalf("worker-2 reused worker-1's Job")
	}
}

// The legacy Job is for a worker whose olaresd cannot run the bootstrap stage:
// 1.12.8 answers readiness without the capability, and anything older serves
// no readiness route. A worker that refuses the token is not reachable at all,
// and is not bootstrapped behind its back.
func TestLegacyWorkerBootstrapOnlyForAnOlaresdThatCannotBootstrap(t *testing.T) {
	for _, tc := range []struct {
		name          string
		before        func() (UpgradeReadiness, error)
		wantBootstrap bool
	}{
		{"no readiness route", func() (UpgradeReadiness, error) {
			return UpgradeReadiness{}, &UpgradeHTTPStatusError{StatusCode: http.StatusNotFound, Status: "404 Not Found"}
		}, true},
		{"no bootstrap capability", func() (UpgradeReadiness, error) {
			return UpgradeReadiness{Supported: true, CLIVersion: "1.12.7"}, nil
		}, true},
		{"auth refusal", func() (UpgradeReadiness, error) {
			return UpgradeReadiness{}, &UpgradeHTTPStatusError{StatusCode: http.StatusForbidden, Status: "403 Forbidden"}
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newUpgradeHarness(t, samplePlan(), twoNodeCluster())
			h.deps.Store = newTestStore(t)
			var mu sync.Mutex
			bootstrapped := false
			h.deps.Upgrade.Readiness = func(_ context.Context, n inventory.Node, _, _ string) (UpgradeReadiness, error) {
				mu.Lock()
				defer mu.Unlock()
				if n.NodeName != "worker-1" || bootstrapped {
					return readyOn("1.12.8"), nil
				}
				return tc.before()
			}
			h.deps.Upgrade.BootstrapLegacy = func(_ context.Context, node inventory.Node, _, version string, _ UpgradeLocations) error {
				if node.NodeName != "worker-1" || version != "1.12.8" {
					t.Errorf("wrong bootstrap target: %s %s", node.NodeName, version)
				}
				mu.Lock()
				bootstrapped = true
				mu.Unlock()
				return nil
			}
			m, err := NewManager(h.deps)
			if err != nil {
				t.Fatal(err)
			}
			op, err := m.Create(context.Background(), CreateRequest{
				Type: TypeUpgrade, RequestID: "olares-upgrade-1.12.8-test",
				Scope: ScopeCluster, ClusterID: "cluster-1", Owner: "alice@olares.com",
				UpgradeVersion: "1.12.8",
			})
			if err != nil {
				t.Fatal(err)
			}
			got := waitForTerminal(t, m, op.ID)
			if bootstrapped != tc.wantBootstrap {
				t.Fatalf("bootstrap = %t, want %t", bootstrapped, tc.wantBootstrap)
			}
			if !tc.wantBootstrap {
				return
			}
			if got.Status != StatusSucceeded {
				t.Fatalf("upgrade = %s/%s", got.Status, got.Code)
			}
			for _, key := range h.order() {
				if key == "worker-1/"+StageBootstrap {
					t.Error("the bootstrap stage was also dispatched to the olaresd that cannot run it")
				}
			}
		})
	}
}
