package chart

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	oac "github.com/beclab/Olares/framework/oac"
	"github.com/beclab/api/manifest"
	"sigs.k8s.io/yaml"
)

func TestFromComposeProducesV3LintableChart(t *testing.T) {
	root := t.TempDir()
	composePath := filepath.Join(root, "compose.yaml")
	compose := []byte(`services:
  web:
    image: beclab/hello:1.0.0
    labels:
      olares.service.type: Entrance
    ports:
      - "8080:80"
`)
	if err := os.WriteFile(composePath, compose, 0o600); err != nil {
		t.Fatal(err)
	}

	chartDir := filepath.Join(root, "testapp")
	result, err := FromCompose(Options{
		ComposeFiles: []string{composePath},
		OutputDir:    chartDir,
		Name:         "testapp",
		Title:        "Test App",
	})
	if err != nil {
		t.Fatalf("FromCompose() error: %v", err)
	}
	if result.EntranceHost != "web" || result.EntrancePort != 8080 {
		t.Fatalf("entrance = %s:%d, want web:8080", result.EntranceHost, result.EntrancePort)
	}

	raw, err := os.ReadFile(filepath.Join(chartDir, appCfgFileName))
	if err != nil {
		t.Fatal(err)
	}
	var cfg manifest.AppConfiguration
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.APIVersion != appAPIVersion {
		t.Fatalf("apiVersion = %q, want %q", cfg.APIVersion, appAPIVersion)
	}
	if cfg.ConfigVersion != configVersion {
		t.Fatalf("olaresManifest.version = %q, want %q", cfg.ConfigVersion, configVersion)
	}
	if cfg.WorkloadReplicas == nil || (*cfg.WorkloadReplicas)["testapp"] != 1 {
		t.Fatalf("workloadReplicas = %#v, want testapp: 1", cfg.WorkloadReplicas)
	}

	var systemDependency *manifest.Dependency
	for i := range cfg.Options.Dependencies {
		dependency := &cfg.Options.Dependencies[i]
		if dependency.Name == olaresSystemDepName && dependency.Type == "system" {
			systemDependency = dependency
			break
		}
	}
	if systemDependency == nil {
		t.Fatal("missing olares system dependency")
	}
	if systemDependency.Version != olaresSystemDepVersion {
		t.Fatalf("olares dependency version = %q, want %q", systemDependency.Version, olaresSystemDepVersion)
	}

	if err := oac.Lint(chartDir, oac.WithAutoOwnerScenarios()); err != nil {
		t.Fatalf("freshly scaffolded chart must pass lint: %v", err)
	}
}

// TestFromComposeMultiServiceChart covers what a single-service compose cannot:
// names kompose leaves invalid or unreachable from Helm, and a bundled datastore
// that must not be mistaken for the app itself.
func TestFromComposeMultiServiceChart(t *testing.T) {
	root := t.TempDir()
	composePath := filepath.Join(root, "compose.yaml")
	compose := []byte(`services:
  db:
    image: postgres:16
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
  web_app:
    image: nginx:1.27
    ports:
      - "8080:80"
  redis_cache:
    image: redis:7
    ports:
      - "6379:6379"
volumes:
  pgdata:
`)
	if err := os.WriteFile(composePath, compose, 0o600); err != nil {
		t.Fatal(err)
	}

	chartDir := filepath.Join(root, "testapp")
	result, err := FromCompose(Options{
		ComposeFiles: []string{composePath},
		OutputDir:    chartDir,
		Name:         "testapp",
		Title:        "Test App",
	})
	if err != nil {
		t.Fatalf("FromCompose() error: %v", err)
	}

	if result.EntranceHost != "web-app" || result.EntrancePort != 8080 {
		t.Fatalf("entrance = %s:%d, want web-app:8080 (the datastores must be skipped)", result.EntranceHost, result.EntrancePort)
	}

	primary := readTemplate(t, chartDir, "deployment-testapp.yaml")
	if !strings.Contains(primary, "image: nginx:1.27") {
		t.Fatalf("workload named after the app must be the web service, got:\n%s", primary)
	}

	// A dashed workload name is only reachable through index.
	cached := readTemplate(t, chartDir, "deployment-redis-cache.yaml")
	if !strings.Contains(cached, `replicas: {{ (index .Values.workloads "redis-cache").replicaCount }}`) {
		t.Fatalf("dashed workload must use the index form, got:\n%s", cached)
	}
	db := readTemplate(t, chartDir, "deployment-db.yaml")
	if !strings.Contains(db, "replicas: {{ .Values.workloads.db.replicaCount }}") {
		t.Fatalf("plain workload name should keep the dotted form, got:\n%s", db)
	}

	entries, err := os.ReadDir(filepath.Join(chartDir, "templates"))
	if err != nil {
		t.Fatal(err)
	}
	invalidName := regexp.MustCompile(`(?m)^\s*name: \S*_`)
	for _, entry := range entries {
		if strings.Contains(entry.Name(), "_") {
			t.Fatalf("template file %q keeps an invalid Kubernetes name", entry.Name())
		}
		if body := readTemplate(t, chartDir, entry.Name()); invalidName.MatchString(body) {
			t.Fatalf("template %q declares a name with an underscore:\n%s", entry.Name(), body)
		}
	}

	values := readFile(t, filepath.Join(chartDir, "values.yaml"))
	for _, workload := range []string{"testapp", "db", "redis-cache"} {
		if !strings.Contains(values, workload+":") {
			t.Fatalf("values.yaml is missing workload %q:\n%s", workload, values)
		}
	}

	if err := oac.Lint(chartDir, oac.WithAutoOwnerScenarios()); err != nil {
		t.Fatalf("multi-service chart must pass lint: %v", err)
	}
}

// TestFromComposeRestartNoBecomesDeployment pins the workload kind for a service
// kompose would render as a bare Pod, which Olares cannot install or suspend.
func TestFromComposeRestartNoBecomesDeployment(t *testing.T) {
	root := t.TempDir()
	composePath := filepath.Join(root, "compose.yaml")
	compose := []byte(`services:
  worker:
    image: alpine:3.20
    restart: "no"
    command: ["sleep", "3600"]
    ports:
      - "9000:9000"
`)
	if err := os.WriteFile(composePath, compose, 0o600); err != nil {
		t.Fatal(err)
	}

	chartDir := filepath.Join(root, "workerapp")
	result, err := FromCompose(Options{
		ComposeFiles: []string{composePath},
		OutputDir:    chartDir,
		Name:         "workerapp",
	})
	if err != nil {
		t.Fatalf("FromCompose() error: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(chartDir, "templates"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "pod-") {
			t.Fatalf("restart: no must not render a bare Pod, got %q", entry.Name())
		}
	}

	workload := readTemplate(t, chartDir, "deployment-workerapp.yaml")
	if !strings.Contains(workload, "restartPolicy: Always") {
		t.Fatalf("a Deployment pod template only accepts restartPolicy Always, got:\n%s", workload)
	}

	raw := readFile(t, filepath.Join(chartDir, appCfgFileName))
	var cfg manifest.AppConfiguration
	if err := yaml.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.WorkloadReplicas == nil || (*cfg.WorkloadReplicas)["workerapp"] != 1 {
		t.Fatalf("workloadReplicas = %#v, want workerapp: 1", cfg.WorkloadReplicas)
	}
	if len(result.Notices) == 0 {
		t.Fatal("converting restart: no into a Deployment must be reported")
	}

	if err := oac.Lint(chartDir, oac.WithAutoOwnerScenarios()); err != nil {
		t.Fatalf("chart must pass lint: %v", err)
	}
}

// TestFromComposeEntranceLabelBeatsAppNamedWorkload keeps an explicit entrance
// label authoritative when another compose service happens to be named after the
// app: that workload blocks the rename, it must not steal the entrance.
func TestFromComposeEntranceLabelBeatsAppNamedWorkload(t *testing.T) {
	root := t.TempDir()
	composePath := filepath.Join(root, "compose.yaml")
	// app owns the app name and the lower port, so it would win both the
	// already-named-after-the-app and the lowest-port paths.
	compose := []byte(`services:
  app:
    image: alpine:3.20
    command: ["sleep", "3600"]
    ports:
      - "8080:8080"
  web:
    image: nginx:1.27
    labels:
      olares.service.type: Entrance
    ports:
      - "9000:80"
`)
	if err := os.WriteFile(composePath, compose, 0o600); err != nil {
		t.Fatal(err)
	}

	chartDir := filepath.Join(root, "app")
	result, err := FromCompose(Options{
		ComposeFiles: []string{composePath},
		OutputDir:    chartDir,
		Name:         "app",
	})
	if err != nil {
		t.Fatalf("FromCompose() error: %v", err)
	}

	if result.EntranceHost != "web" || result.EntrancePort != 9000 {
		t.Fatalf("entrance = %s:%d, want web:9000 (%s)", result.EntranceHost, result.EntrancePort, result.EntranceReason)
	}
	if result.EntranceGuessed {
		t.Fatalf("a labeled entrance is not a guess, reason: %s", result.EntranceReason)
	}

	// Renaming web to app would have overwritten deployment-app.yaml.
	if body := readTemplate(t, chartDir, "deployment-app.yaml"); !strings.Contains(body, "image: alpine:3.20") {
		t.Fatalf("the workload named after the app must still be the app service, got:\n%s", body)
	}
	if body := readTemplate(t, chartDir, "deployment-web.yaml"); !strings.Contains(body, "image: nginx:1.27") {
		t.Fatalf("the labeled service keeps its own workload, got:\n%s", body)
	}

	var cfg manifest.AppConfiguration
	if err := yaml.Unmarshal([]byte(readFile(t, filepath.Join(chartDir, appCfgFileName))), &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.WorkloadReplicas == nil ||
		(*cfg.WorkloadReplicas)["app"] != 1 || (*cfg.WorkloadReplicas)["web"] != 1 {
		t.Fatalf("workloadReplicas = %#v, want app and web", cfg.WorkloadReplicas)
	}

	if !strings.Contains(strings.Join(result.Notices, "\n"), "already carries the app name") {
		t.Fatalf("the skipped rename must be explained, notices: %v", result.Notices)
	}

	if err := oac.Lint(chartDir, oac.WithAutoOwnerScenarios()); err != nil {
		t.Fatalf("chart must pass lint: %v", err)
	}
}

// TestFromComposeWithoutScalableWorkloadWritesNothing pins the fail-fast: a
// compose file that renders no Deployment/StatefulSet must not leave a partial
// chart behind.
func TestFromComposeWithoutScalableWorkloadWritesNothing(t *testing.T) {
	root := t.TempDir()
	composePath := filepath.Join(root, "compose.yaml")
	compose := []byte(`services:
  agent:
    image: alpine:3.20
    command: ["sleep", "3600"]
    labels:
      kompose.controller.type: daemonset
`)
	if err := os.WriteFile(composePath, compose, 0o600); err != nil {
		t.Fatal(err)
	}

	chartDir := filepath.Join(root, "agentapp")
	if _, err := FromCompose(Options{
		ComposeFiles: []string{composePath},
		OutputDir:    chartDir,
		Name:         "agentapp",
	}); err == nil {
		t.Fatal("a chart without any Deployment/StatefulSet must be rejected")
	} else if !strings.Contains(err.Error(), "no Deployment or StatefulSet") {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(chartDir, "templates"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("a rejected conversion must not write templates, got %d files", len(entries))
	}
}

// TestFromComposeKeepsRenamedReferences covers the names kompose only
// half-normalizes: renaming the object without following its references would
// leave a chart that lints clean and then never starts.
func TestFromComposeKeepsRenamedReferences(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "Prod.env"), []byte("TOKEN=abc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	composePath := filepath.Join(root, "compose.yaml")
	// web.ui exercises the Service and container names kompose leaves dotted while
	// labeling the pods (and the Ingress backend) with web-ui; PGData, my.data and
	// Prod.env exercise the volume and env-file names it never fully normalizes.
	compose := []byte(`services:
  web.ui:
    image: nginx:1.27
    labels:
      kompose.service.expose: "ui.example.com"
    env_file:
      - ./Prod.env
    ports:
      - "8080:80"
    volumes:
      - PGData:/var/lib/data
      - my.data:/var/lib/more
volumes:
  PGData:
  my.data:
`)
	if err := os.WriteFile(composePath, compose, 0o600); err != nil {
		t.Fatal(err)
	}

	chartDir := filepath.Join(root, "webui")
	if _, err := FromCompose(Options{
		ComposeFiles: []string{composePath},
		OutputDir:    chartDir,
		Name:         "webui",
	}); err != nil {
		t.Fatalf("FromCompose() error: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(chartDir, "templates"))
	if err != nil {
		t.Fatal(err)
	}
	names := make(map[string]bool)
	var rendered strings.Builder
	for _, entry := range entries {
		names[entry.Name()] = true
		rendered.WriteString(readTemplate(t, chartDir, entry.Name()))
	}
	for _, want := range []string{
		"service-web-ui.yaml", "ingress-web-ui.yaml", "persistentvolumeclaim-pgdata.yaml",
		"persistentvolumeclaim-my.data.yaml", "configmap-prod-env.yaml",
	} {
		if !names[want] {
			t.Fatalf("missing template %q, got %v", want, names)
		}
	}

	// A name or claim still spelled the compose way points at no object. Labels
	// are exempt: kompose keeps the raw compose name there and that is legal.
	nameValues := regexp.MustCompile(`(?m)^\s*-?\s*(?:name|claimName): (\S+)$`)
	for _, match := range nameValues.FindAllStringSubmatch(rendered.String(), -1) {
		switch match[1] {
		case "PGData", "Prod-env", "web.ui":
			t.Fatalf("templates still reference the pre-normalization name %q:\n%s", match[1], rendered.String())
		}
	}
	workload := readTemplate(t, chartDir, "deployment-webui.yaml")
	for _, want := range []string{"claimName: pgdata", "name: pgdata", "name: prod-env",
		// A PVC name may carry a dot, a volume name may not, so the two differ.
		"claimName: my.data", "name: my-data"} {
		if !strings.Contains(workload, want) {
			t.Fatalf("workload must reference the renamed object (%q):\n%s", want, workload)
		}
	}
	if strings.Contains(workload, "name: my.data") {
		t.Fatalf("a volume and its mounts must be a DNS label, without dots:\n%s", workload)
	}
	if body := readTemplate(t, chartDir, "ingress-web-ui.yaml"); !strings.Contains(body, "name: web-ui") {
		t.Fatalf("ingress must point at the rendered service name:\n%s", body)
	}

	if err := oac.Lint(chartDir, oac.WithAutoOwnerScenarios()); err != nil {
		t.Fatalf("chart must pass lint: %v", err)
	}
}

// TestFromComposeRejectsNameCollision pins that two objects normalizing onto one
// name fail instead of overwriting each other's template.
func TestFromComposeRejectsNameCollision(t *testing.T) {
	root := t.TempDir()
	composePath := filepath.Join(root, "compose.yaml")
	compose := []byte(`services:
  app:
    image: nginx:1.27
    ports:
      - "8080:80"
    volumes:
      - PGData:/data/one
      - pgdata:/data/two
volumes:
  PGData:
  pgdata:
`)
	if err := os.WriteFile(composePath, compose, 0o600); err != nil {
		t.Fatal(err)
	}

	chartDir := filepath.Join(root, "collide")
	if _, err := FromCompose(Options{
		ComposeFiles: []string{composePath},
		OutputDir:    chartDir,
		Name:         "collide",
	}); err == nil {
		t.Fatal("two PVCs normalizing onto one name must be rejected")
	} else if !strings.Contains(err.Error(), "both become") {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(chartDir, "templates"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("a rejected conversion must not write templates, got %d files", len(entries))
	}
}

// TestFromComposeReportsDroppedCronJobSchedule covers the compose label the
// pinned Deployment controller makes unreachable: the scheduled service becomes
// an always-on workload, which has to be said out loud, and it goes through the
// same normalization and resource stamping as every other workload.
func TestFromComposeReportsDroppedCronJobSchedule(t *testing.T) {
	root := t.TempDir()
	composePath := filepath.Join(root, "compose.yaml")
	// restart: no is what kompose needs to honour the schedule at all, so this is
	// the compose file that would have produced a CronJob without the pinning.
	compose := []byte(`services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
  backup:
    image: alpine:3.20
    restart: "no"
    command: ["sh", "-c", "echo backup"]
    labels:
      kompose.cronjob.schedule: "0 3 * * *"
    volumes:
      - PGData:/var/lib/data
volumes:
  PGData:
`)
	if err := os.WriteFile(composePath, compose, 0o600); err != nil {
		t.Fatal(err)
	}

	chartDir := filepath.Join(root, "cronapp")
	result, err := FromCompose(Options{
		ComposeFiles: []string{composePath},
		OutputDir:    chartDir,
		Name:         "cronapp",
	})
	if err != nil {
		t.Fatalf("FromCompose() error: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(chartDir, "templates"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "cronjob-") || strings.HasPrefix(entry.Name(), "pod-") {
			t.Fatalf("a scheduled service must still render as a Deployment, got %q", entry.Name())
		}
	}

	if !strings.Contains(strings.Join(result.Notices, "\n"), "kompose.cronjob.schedule") {
		t.Fatalf("the dropped schedule must be reported, notices: %v", result.Notices)
	}

	workload := readTemplate(t, chartDir, "deployment-backup.yaml")
	for _, want := range []string{"claimName: pgdata", "name: pgdata", "restartPolicy: Always",
		"cpu: 100m", "memory: 128Mi"} {
		if !strings.Contains(workload, want) {
			t.Fatalf("the scheduled service's workload must contain %q:\n%s", want, workload)
		}
	}
	if strings.Contains(workload, "name: PGData") {
		t.Fatalf("volume names must be normalized here too:\n%s", workload)
	}

	var cfg manifest.AppConfiguration
	if err := yaml.Unmarshal([]byte(readFile(t, filepath.Join(chartDir, appCfgFileName))), &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.WorkloadReplicas == nil ||
		(*cfg.WorkloadReplicas)["cronapp"] != 1 || (*cfg.WorkloadReplicas)["backup"] != 1 {
		t.Fatalf("workloadReplicas = %#v, want cronapp and backup", cfg.WorkloadReplicas)
	}

	if err := oac.Lint(chartDir, oac.WithAutoOwnerScenarios()); err != nil {
		t.Fatalf("chart must pass lint: %v", err)
	}
}

// TestFromComposeReportsInvalidServiceName covers the compose service name no
// normalization can repair: a leading digit is legal in compose and illegal in a
// Service name, and renaming the Service alone would orphan the pod selector.
func TestFromComposeReportsInvalidServiceName(t *testing.T) {
	root := t.TempDir()
	composePath := filepath.Join(root, "compose.yaml")
	compose := []byte(`services:
  1web:
    image: nginx:1.27
    ports:
      - "8080:80"
`)
	if err := os.WriteFile(composePath, compose, 0o600); err != nil {
		t.Fatal(err)
	}

	chartDir := filepath.Join(root, "digitapp")
	result, err := FromCompose(Options{
		ComposeFiles: []string{composePath},
		OutputDir:    chartDir,
		Name:         "digitapp",
	})
	if err != nil {
		t.Fatalf("FromCompose() error: %v", err)
	}

	notices := strings.Join(result.Notices, "\n")
	if !strings.Contains(notices, `Service "1web" is not a valid Kubernetes name`) {
		t.Fatalf("an unusable Service name must be reported, notices: %v", result.Notices)
	}
	// Renaming it would point the pod selector at nothing, so it stays as it is.
	if body := readTemplate(t, chartDir, "service-1web.yaml"); !strings.Contains(body, "name: 1web") {
		t.Fatalf("the Service must keep the name kompose selected on:\n%s", body)
	}
}

// TestFromComposeReportsLostSemanticsForLabeledController pins that the notices
// follow the workload a service really renders as: kompose decides the CronJob
// and bare-Pod question from the pinned controller alone, so a service moved to a
// StatefulSet by label loses its schedule and its restart policy just the same.
func TestFromComposeReportsLostSemanticsForLabeledController(t *testing.T) {
	root := t.TempDir()
	composePath := filepath.Join(root, "compose.yaml")
	compose := []byte(`services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
  store:
    image: alpine:3.20
    restart: "no"
    command: ["sh", "-c", "echo store"]
    labels:
      kompose.controller.type: statefulset
      kompose.cronjob.schedule: "0 3 * * *"
`)
	if err := os.WriteFile(composePath, compose, 0o600); err != nil {
		t.Fatal(err)
	}

	chartDir := filepath.Join(root, "mixapp")
	result, err := FromCompose(Options{
		ComposeFiles: []string{composePath},
		OutputDir:    chartDir,
		Name:         "mixapp",
	})
	if err != nil {
		t.Fatalf("FromCompose() error: %v", err)
	}

	workload := readTemplate(t, chartDir, "statefulset-store.yaml")
	if !strings.Contains(workload, "restartPolicy: Always") {
		t.Fatalf("a StatefulSet pod template only accepts restartPolicy Always, got:\n%s", workload)
	}

	notices := strings.Join(result.Notices, "\n")
	for _, want := range []string{
		`service "store" sets kompose.cronjob.schedule "0 3 * * *"`,
		"always-on StatefulSet",
		"rendered as a StatefulSet with restartPolicy Always",
	} {
		if !strings.Contains(notices, want) {
			t.Fatalf("notices must contain %q, got: %v", want, result.Notices)
		}
	}
	if strings.Contains(notices, `service "store" sets restart: no, but was rendered as a Deployment`) {
		t.Fatalf("a labeled controller must not be described as a Deployment, notices: %v", result.Notices)
	}

	if err := oac.Lint(chartDir, oac.WithAutoOwnerScenarios()); err != nil {
		t.Fatalf("chart must pass lint: %v", err)
	}
}

func readTemplate(t *testing.T, chartDir, name string) string {
	t.Helper()
	return readFile(t, filepath.Join(chartDir, "templates", name))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
