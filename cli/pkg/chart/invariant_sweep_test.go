package chart

import (
	"os"
	"path/filepath"
	"testing"

	oac "github.com/beclab/Olares/framework/oac"
)

// TestFromComposeHoldsInvariantsAcrossComposeFeatures runs the compose features
// that make kompose emit an object with a reference — env_file and configs
// ConfigMaps, secrets, PVCs, an Ingress, a DaemonSet — through the invariants, so
// the answer to "are all the references still consistent" comes from the code
// rather than from whichever one a reviewer thought to look at.
func TestFromComposeHoldsInvariantsAcrossComposeFeatures(t *testing.T) {
	cases := map[string]string{
		"env file becomes a ConfigMap": `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
    env_file:
      - Prod.Env
`,
		"secret becomes a volume": `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
    secrets:
      - My_Token
secrets:
  My_Token:
    file: ./token.txt
`,
		// kompose folds the underscore in the case above on both sides of the
		// reference, so it agrees with itself. A leading separator does not
		// survive an object name, which is where the two sides can part ways.
		"secret name starts with a separator": `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
    secrets:
      - _token
secrets:
  _token:
    file: ./token.txt
`,
		"config becomes a ConfigMap": `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
    configs:
      - source: App_Conf
        target: /etc/app.conf
configs:
  App_Conf:
    file: ./app.conf
`,
		"bind mount is copied into a ConfigMap": `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
    volumes:
      - ./conf:/etc/conf
`,
		"daemonset by label": `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
  agent:
    image: alpine:3.20
    command: ["sleep", "3600"]
    ports:
      - "9100:9100"
    labels:
      kompose.controller.type: daemonset
`,
		"mixed protocols on one service": `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
      - "5353:5353/udp"
      - "9090:9090"
`,
		"dotted names with a dotted volume": `services:
  web.ui:
    image: nginx:1.27
    ports:
      - "8080:80"
    volumes:
      - Cache.Data:/var/cache
  api_v2:
    image: nginx:1.27
    ports:
      - "9000:9000"
    depends_on:
      - web.ui
volumes:
  Cache.Data:
`,
		"exposed statefulset with a volume": `services:
  web:
    image: nginx:1.27
    ports:
      - "8080:80"
  store_db:
    image: postgres:16
    ports:
      - "5432:5432"
    volumes:
      - PGData:/var/lib/postgresql/data
    labels:
      kompose.controller.type: statefulset
      kompose.service.expose: "db.example.com"
volumes:
  PGData:
`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			composePath := filepath.Join(root, "compose.yaml")
			if err := os.WriteFile(composePath, []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
			// The files the cases above reference: kompose reads a config, secret
			// or env_file at conversion time and copies it into the chart.
			for _, file := range []string{"Prod.Env", "token.txt", "app.conf"} {
				if err := os.WriteFile(filepath.Join(root, file), []byte("KEY=value\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.MkdirAll(filepath.Join(root, "conf"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "conf", "app.ini"), []byte("x=1\n"), 0o600); err != nil {
				t.Fatal(err)
			}

			chartDir := filepath.Join(root, "sweepapp")
			if _, err := FromCompose(Options{
				ComposeFiles: []string{composePath},
				OutputDir:    chartDir,
				Name:         "sweepapp",
			}); err != nil {
				t.Fatalf("FromCompose() error: %v", err)
			}

			assertChartInvariants(t, chartDir)
			if err := oac.Lint(chartDir, oac.WithAutoOwnerScenarios()); err != nil {
				t.Fatalf("chart must pass lint: %v", err)
			}
		})
	}
}
