package os

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// LogCollectOptions holds options for collecting logs
type LogCollectOptions struct {
	// Time duration to collect logs for (empty means all available logs)
	Since string
	// Maximum number of lines to collect per log source
	MaxLines int
	// Output directory for collected logs
	OutputDir string
	// Components to collect logs from (empty means all)
	Components []string
	// Whether to ignore errors from kubectl commands
	IgnoreKubeErrors bool
	// Skip retrieving logs from kube-apiserver
	SkipKubeAPISserver bool
	// PodNamespaces restricts /var/log/pods collection to the given
	// namespaces. Nil/empty means collect all namespaces (backward
	// compatible). Ignored when SkipPodLogs is set.
	PodNamespaces []string
	// Section switches. All default to false to preserve the original
	// full-collection behavior when no flag is given.
	SkipSystemd     bool
	SkipDmesg       bool
	SkipNetwork     bool
	SkipClusterInfo bool
	SkipPodLogs     bool
	// DmesgPrevBoots is how many boots before the current one to also
	// collect kernel logs for. Part of the dmesg section, so SkipDmesg
	// turns it off too.
	DmesgPrevBoots int
}

var servicesToCollectLogs = []string{"k3s", "containerd", "olaresd", "kubelet", "juicefs", "redis", "minio", "etcd", "NetworkManager"}

// journalctlBin is the journalctl executable journalctlToTar runs. It is a
// variable so tests can substitute a stub instead of requiring a host with a
// populated journal.
var journalctlBin = "journalctl"

// setSkipIfK8sNotReachable checks if the Kubernetes API server port is reachable
// and automatically sets skip-kube-apiserver to true if not reachable
func setSkipIfK8sNotReachable(options *LogCollectOptions) {
	// if the env is not set explicitly by user
	// fallback to k3s config path as it's a non-standard path
	if os.Getenv(clientcmd.RecommendedConfigPathEnvVar) == "" {
		os.Setenv(clientcmd.RecommendedConfigPathEnvVar, "/etc/rancher/k3s/k3s.yaml")
	}
	config, err := ctrl.GetConfig()
	if err != nil {
		fmt.Printf("Warning: failed to get kubeconfig: %v\n", err)
		fmt.Println("Automatically setting skip-kube-apiserver option")
		options.SkipKubeAPISserver = true
		return
	}
	url, _, err := rest.DefaultServerUrlFor(config)
	if err != nil {
		fmt.Printf("Warning: failed to parse server url in kubeconfig: %v\n", err)
		fmt.Println("Automatically setting skip-kube-apiserver option")
		options.SkipKubeAPISserver = true
		return
	}
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	conn, err := net.DialTimeout("tcp", url.Host, timeout)
	if err != nil {
		fmt.Printf("Warning: Kubernetes API server at %s is not reachable: %v\n", config.Host, err)
		fmt.Println("Automatically setting skip-kube-apiserver option")
		options.SkipKubeAPISserver = true
		return
	}
	conn.Close()
}

func collectLogs(options *LogCollectOptions) error {
	if os.Getuid() != 0 {
		return fmt.Errorf("os: please run as root")
	}

	if !options.SkipClusterInfo && !options.SkipKubeAPISserver {
		setSkipIfK8sNotReachable(options)
	}

	if err := os.MkdirAll(options.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	archiveName := filepath.Join(options.OutputDir, fmt.Sprintf("olares-logs-%s.tar.gz", timestamp))

	archive, err := os.Create(archiveName)
	if err != nil {
		return fmt.Errorf("failed to create archive: %v", err)
	}
	defer archive.Close()

	gw := gzip.NewWriter(archive)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	if !options.SkipSystemd {
		if err := collectSystemdLogs(tw, options); err != nil {
			return fmt.Errorf("failed to collect systemd logs: %v", err)
		}
	}

	if !options.SkipDmesg {
		fmt.Println("collecting dmesg logs ...")
		if err := collectDmesgLogs(tw, options); err != nil {
			return fmt.Errorf("failed to collect dmesg logs: %v", err)
		}
		if options.DmesgPrevBoots > 0 {
			fmt.Println("collecting kernel logs of previous boots ...")
			if err := collectPrevBootKernelLogs(tw, options); err != nil {
				return fmt.Errorf("failed to collect kernel logs of previous boots: %v", err)
			}
		}
	}

	fmt.Println("collecting logs from kubernetes cluster...")
	if err := collectKubernetesLogs(tw, options); err != nil {
		return fmt.Errorf("failed to collect kubernetes logs: %v", err)
	}

	// olares-cli's own logs are host-level Olares operational logs; tie
	// them to cluster-info so a pods-only scoped collection excludes them.
	if !options.SkipClusterInfo {
		fmt.Println("collecting olares-cli logs...")
		if err := collectOlaresCLILogs(tw, options); err != nil {
			return fmt.Errorf("failed to collect OlaresCLI logs: %v", err)
		}
	}

	if !options.SkipNetwork {
		fmt.Println("collecting network configs...")
		if err := collectNetworkConfigs(tw, options); err != nil {
			return fmt.Errorf("failed to collect network configs: %v", err)
		}
	}

	fmt.Printf("logs have been collected and archived in: %s\n", archiveName)
	return nil
}

func collectOlaresCLILogs(tw *tar.Writer, options *LogCollectOptions) error {
	basedir, err := getBaseDir()
	if err != nil {
		return err
	}
	cliLogDir := filepath.Join(basedir, "logs")
	if _, err := os.Stat(cliLogDir); err != nil {
		fmt.Printf("warning: directory %s does not exist, skipping collecting olares-cli logs\n", cliLogDir)
		return nil
	}
	err = filepath.Walk(cliLogDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", path, err)
		}
		defer srcFile.Close()

		relPath, err := filepath.Rel(cliLogDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		header := &tar.Header{
			Name:    filepath.Join("olares-cli", relPath),
			Mode:    0644,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write header for %s: %v", path, err)
		}

		// stream file contents to tar
		if _, err := io.CopyN(tw, srcFile, header.Size); err != nil {
			return fmt.Errorf("failed to write data for %s: %v", path, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to collect olares-cli logs from %s: %v", cliLogDir, err)
	}
	return nil
}

func collectSystemdLogs(tw *tar.Writer, options *LogCollectOptions) error {
	// Create temp directory for log files
	tempDir, err := os.MkdirTemp("", "olares-logs-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	var services []string
	if len(options.Components) > 0 {
		services = options.Components
	} else {
		services = servicesToCollectLogs
	}
	for _, service := range services {
		if !checkServiceExists(service) {
			if len(options.Components) > 0 {
				fmt.Printf("warning: required service %s not found\n", service)
			}
			continue
		}

		fmt.Printf("collecting logs for service: %s\n", service)

		// create temp file for this service's logs
		tempFile := filepath.Join(tempDir, fmt.Sprintf("%s.log", service))
		logFile, err := os.Create(tempFile)
		if err != nil {
			return fmt.Errorf("failed to create temp file for %s: %v", service, err)
		}

		args := []string{"-u", service}
		if options.Since != "" {
			if !strings.HasPrefix(options.Since, "-") {
				options.Since = "-" + options.Since
			}
			args = append(args, "--since", options.Since)
		}
		if options.MaxLines > 0 {
			args = append(args, "-n", fmt.Sprintf("%d", options.MaxLines))
		}

		if options.Since != "" && options.MaxLines > 0 {
			// this is a journalctl bug
			// where -S and -n combined results in the latest logs truncated
			// rather than the old logs
			// a -r corrects the truncate behavior
			args = append(args, "-r")
		}

		// execute journalctl and write directly to temp file
		// don't just use the command output because that's too memory-consuming
		// the same logic goes to the os.Open and io.Copy rather than os.ReadFile
		cmd := exec.Command("journalctl", args...)
		cmd.Stdout = logFile
		if err := cmd.Run(); err != nil {
			logFile.Close()
			return fmt.Errorf("failed to collect logs for %s: %v", service, err)
		}
		logFile.Close()

		// get file info for the tar header
		fi, err := os.Stat(tempFile)
		if err != nil {
			return fmt.Errorf("failed to stat temp file for %s: %v", service, err)
		}

		logFile, err = os.Open(tempFile)
		if err != nil {
			return fmt.Errorf("failed to open temp file for %s: %v", service, err)
		}
		defer logFile.Close()

		header := &tar.Header{
			Name:    fmt.Sprintf("%s.log", service),
			Mode:    0644,
			Size:    fi.Size(),
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write header for %s: %v", service, err)
		}

		if _, err := io.CopyN(tw, logFile, header.Size); err != nil {
			return fmt.Errorf("failed to write logs for %s: %v", service, err)
		}
	}
	return nil
}

func collectDmesgLogs(tw *tar.Writer, options *LogCollectOptions) error {
	cmd := exec.Command("dmesg", "-T")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	header := &tar.Header{
		Name:    "dmesg.log",
		Mode:    0644,
		Size:    int64(len(output)),
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write dmesg header: %v", err)
	}
	if _, err := tw.Write(output); err != nil {
		return fmt.Errorf("failed to write dmesg data: %v", err)
	}
	return nil
}

// collectPrevBootKernelLogs adds the kernel log of earlier boots, which
// `dmesg` cannot reach because it only reads the running kernel's ring
// buffer. When a machine froze or was hard-reset, its OOM kills and hardware
// errors are recorded only there, and rebooting to run this command is
// exactly what destroys the evidence. The per-service collection above does
// span boots, but journalctl -u never includes kernel messages, so without
// this the incident's kernel side is simply unavailable.
//
// Every failure here is reported as a warning rather than returned: a host
// with volatile journaling (no /var/log/journal) or with fewer boots on
// record than requested is a normal configuration, not a collection failure.
func collectPrevBootKernelLogs(tw *tar.Writer, options *LogCollectOptions) error {
	if _, err := util.GetCommand(journalctlBin); err != nil {
		fmt.Println("warning: journalctl not found, skipping kernel logs of previous boots")
		return nil
	}

	tempDir, err := os.MkdirTemp("", "olares-kernel-logs-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// The boot list resolves the -N offsets used below to boot IDs and time
	// ranges, which is what lets a reader tell which file covers the
	// incident being investigated.
	if err := journalctlToTar(tw, tempDir, "boots.txt", "--list-boots"); err != nil {
		fmt.Printf("warning: failed to collect the boot list: %v\n", err)
	}

	for i := 1; i <= options.DmesgPrevBoots; i++ {
		args := []string{"-k", "-b", fmt.Sprintf("-%d", i)}
		if options.MaxLines > 0 {
			args = append(args, "-n", fmt.Sprintf("%d", options.MaxLines))
		}
		if err := journalctlToTar(tw, tempDir, fmt.Sprintf("dmesg-prev-%d.log", i), args...); err != nil {
			// Boots are requested newest first, so once one is out of
			// reach every older one is too.
			fmt.Printf("warning: kernel log of boot -%d is unavailable, not looking further back: %v\n", i, err)
			return nil
		}
		fmt.Printf("collected kernel log of boot -%d\n", i)
	}
	return nil
}

// journalctlToTar runs one journalctl invocation and stores its output in the
// archive under name. The output is staged in a temp file rather than held in
// memory because a single boot's kernel log can be large.
func journalctlToTar(tw *tar.Writer, tempDir, name string, args ...string) error {
	tempFile := filepath.Join(tempDir, name)
	out, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}

	cmd := exec.Command(journalctlBin, append([]string{"--no-pager"}, args...)...)
	cmd.Stdout = out
	// journalctl explains an unreachable boot on stderr (e.g. "Data from the
	// specified boot is not available"); that message is the whole point of
	// the warning the caller prints.
	var stderr strings.Builder
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	out.Close()
	if runErr != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("%v: %s", runErr, msg)
		}
		return runErr
	}

	fi, err := os.Stat(tempFile)
	if err != nil {
		return fmt.Errorf("failed to stat temp file: %v", err)
	}
	// Some journalctl versions report an unreachable boot on stderr while
	// still exiting 0, leaving an empty file. Treat that as unavailable so
	// the archive does not carry a misleading empty log.
	if fi.Size() == 0 {
		return fmt.Errorf("journalctl returned no output")
	}

	in, err := os.Open(tempFile)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %v", err)
	}
	defer in.Close()

	header := &tar.Header{
		Name:    name,
		Mode:    0644,
		Size:    fi.Size(),
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write header for %s: %v", name, err)
	}
	if _, err := io.CopyN(tw, in, header.Size); err != nil {
		return fmt.Errorf("failed to write data for %s: %v", name, err)
	}
	return nil
}

func collectPodLogs(tw *tar.Writer, options *LogCollectOptions) error {
	if options.SkipPodLogs {
		return nil
	}

	podsLogDir := "/var/log/pods"
	if _, err := os.Stat(podsLogDir); err != nil {
		fmt.Printf("warning: directory %s does not exist, skipping collecting pod logs\n", podsLogDir)
		return nil
	}

	var nsFilter map[string]struct{}
	if len(options.PodNamespaces) > 0 {
		nsFilter = make(map[string]struct{}, len(options.PodNamespaces))
		for _, ns := range options.PodNamespaces {
			nsFilter[ns] = struct{}{}
		}
	}

	err := filepath.Walk(podsLogDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// /var/log/pods/<namespace>_<pod>_<uid>: prune whole pod dirs
			// whose namespace is not requested. namespace/pod names never
			// contain '_', so the prefix is unambiguous.
			if nsFilter != nil && filepath.Dir(path) == podsLogDir {
				ns := strings.SplitN(info.Name(), "_", 2)[0]
				if _, ok := nsFilter[ns]; !ok {
					return filepath.SkipDir
				}
			}
			return nil
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", path, err)
		}
		defer srcFile.Close()

		relPath, err := filepath.Rel(podsLogDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		header := &tar.Header{
			Name:    filepath.Join("pods", relPath),
			Mode:    0644,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write header for %s: %v", path, err)
		}

		// stream file contents to tar
		if _, err := io.CopyN(tw, srcFile, header.Size); err != nil {
			return fmt.Errorf("failed to write data for %s: %v", path, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to collect pod logs from /var/log/pods: %v", err)
	}
	return nil
}

func collectKubernetesLogs(tw *tar.Writer, options *LogCollectOptions) error {
	if err := collectPodLogs(tw, options); err != nil {
		return err
	}

	if options.SkipClusterInfo {
		return nil
	}

	if options.SkipKubeAPISserver {
		return nil
	}

	if _, err := util.GetCommand("kubectl"); err != nil {
		fmt.Printf("warning: kubectl not found, skipping collecting cluster info from kube-apiserver\n")
		return nil
	}

	var cmd *exec.Cmd
	var output []byte
	var err error

	cmd = exec.Command("kubectl", "get", "pods", "--all-namespaces", "-o", "wide")
	output, err = tryKubectlCommand(cmd, "get pods", options)
	if err != nil && !options.IgnoreKubeErrors {
		return err
	}
	if err == nil {
		header := &tar.Header{
			Name:    "pods-list.txt",
			Mode:    0644,
			Size:    int64(len(output)),
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write pods list header: %v", err)
		}
		if _, err := tw.Write(output); err != nil {
			return fmt.Errorf("failed to write pods list data: %v", err)
		}
	}

	resourceTypes := []string{"node", "pod", "statefulset", "deployment", "replicaset", "service", "configmap"}

	for _, res := range resourceTypes {
		cmd = exec.Command("kubectl", "describe", res, "--all-namespaces")
		output, err = tryKubectlCommand(cmd, fmt.Sprintf("describe %s", res), options)
		if err != nil && !options.IgnoreKubeErrors {
			return err
		}
		if err == nil {
			header := &tar.Header{
				Name:    fmt.Sprintf("%s-describe.txt", res),
				Mode:    0644,
				Size:    int64(len(output)),
				ModTime: time.Now(),
			}
			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("failed to write %s description header: %v", res, err)
			}
			if _, err := tw.Write(output); err != nil {
				return fmt.Errorf("failed to write %s description data: %v", res, err)
			}
		}
	}

	fmt.Println("collecting envoy config...")
	if err := collectEnvoyConfig(tw); err != nil && !options.IgnoreKubeErrors {
		return fmt.Errorf("failed to collect envoy config: %v", err)
	}

	return nil
}

func collectNetworkConfigs(tw *tar.Writer, options *LogCollectOptions) error {
	if _, err := util.GetCommand("ip"); err == nil {
		cmd := exec.Command("ip", "address")
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		header := &tar.Header{
			Name:    "ip-address.txt",
			Mode:    0644,
			Size:    int64(len(output)),
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write ip address header: %v", err)
		}
		if _, err := tw.Write(output); err != nil {
			return fmt.Errorf("failed to write ip address data: %v", err)
		}

		cmd = exec.Command("ip", "route")
		output, err = cmd.Output()
		if err != nil {
			return err
		}
		header = &tar.Header{
			Name:    "ip-route.txt",
			Mode:    0644,
			Size:    int64(len(output)),
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write ip route header: %v", err)
		}
		if _, err := tw.Write(output); err != nil {
			return fmt.Errorf("failed to write ip route data: %v", err)
		}
	}

	if _, err := util.GetCommand("iptables-save"); err == nil {
		cmd := exec.Command("iptables-save")
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		header := &tar.Header{
			Name:    "iptables.txt",
			Mode:    0644,
			Size:    int64(len(output)),
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write iptables header: %v", err)
		}
		if _, err := tw.Write(output); err != nil {
			return fmt.Errorf("failed to write iptables data: %v", err)
		}
	}

	if _, err := util.GetCommand("nft"); err == nil {
		cmd := exec.Command("nft", "list", "ruleset")
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		header := &tar.Header{
			Name:    "nftables.txt",
			Mode:    0644,
			Size:    int64(len(output)),
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write nftables header: %v", err)
		}
		if _, err := tw.Write(output); err != nil {
			return fmt.Errorf("failed to write nftables data: %v", err)
		}
	}

	return nil
}
func collectEnvoyConfig(tw *tar.Writer) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		fmt.Printf("  skipping envoy config: failed to get kubeconfig: %v\n", err)
		return nil
	}
	scheme := kruntime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add apps/v1 scheme: %v", err)
	}
	c, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var deploy appsv1.Deployment
	key := ctrlclient.ObjectKey{Namespace: "os-network", Name: "l4-bfl-proxy"}
	if err := c.Get(ctx, key, &deploy); err != nil {
		fmt.Printf("  skipping envoy config: l4-bfl-proxy deployment not found: %v\n", err)
		return nil
	}
	if deploy.Status.AvailableReplicas == 0 {
		fmt.Println("  skipping envoy config: l4-bfl-proxy deployment is not ready (no available replicas)")
		return nil
	}

	url := "http://127.0.0.1:19000/config_dump"
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to request envoy config: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("envoy config API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("kubeerror: failed to read envoy config body: %v", err)
	}

	header := &tar.Header{
		Name:    "envoy-config.json",
		Mode:    0644,
		Size:    int64(len(body)),
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write envoy-config.json header: %v", err)
	}
	if _, err := tw.Write(body); err != nil {
		return fmt.Errorf("failed to write envoy-config.json data: %v", err)
	}

	return nil
}

func getBaseDir() (string, error) {
	basedir := viper.GetString(common.FlagBaseDir)
	if basedir != "" {
		return basedir, nil
	}
	homeDir, err := util.Home()
	if err != nil {
		return "", fmt.Errorf("failed to get home dir: %v", err)
	}
	return filepath.Join(homeDir, ".olares"), nil
}

func tryKubectlCommand(cmd *exec.Cmd, description string, options *LogCollectOptions) ([]byte, error) {
	output, err := cmd.Output()
	if err != nil {
		if options.IgnoreKubeErrors {
			fmt.Printf("warning: failed to %s: %v\n", description, err)
			return nil, err
		}
		return nil, fmt.Errorf("failed to %s: %v", description, err)
	}
	return output, nil
}

// checkService verifies if a systemd service exists
func checkServiceExists(service string) bool {
	if !strings.HasSuffix(service, ".service") {
		service += ".service"
	}
	cmd := exec.Command("systemctl", "list-unit-files", "--no-legend", service)
	return cmd.Run() == nil
}

func NewCmdLogs() *cobra.Command {
	options := &LogCollectOptions{
		Since:            "7d",
		MaxLines:         20000,
		OutputDir:        "./olares-logs",
		IgnoreKubeErrors: false,
		DmesgPrevBoots:   1,
	}

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Collect logs from all Olares system components",
		Long: `Collect logs from various Olares system components, that may or may not be installed on this machine, including:
- K3s/Kubelet logs
- Containerd logs
- JuiceFS logs
- Redis logs
- MinIO logs
- etcd logs
- Olaresd logs
- olares-cli logs
- kernel logs of the current boot, and of previous boots when the journal is persistent
- network configurations
- Kubernetes pod info and logs
- Kubernetes node info`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := collectLogs(options); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	cmd.Flags().StringVar(&options.Since, "since", options.Since, "Only return logs newer than a relative duration like 5s, 2m, or 3h, to limit the log file")
	cmd.Flags().IntVar(&options.MaxLines, "max-lines", options.MaxLines, "Maximum number of lines to collect per log source, to limit the log file size")
	cmd.Flags().StringVar(&options.OutputDir, "output-dir", options.OutputDir, "Directory to store collected logs, will be created if not existing")
	cmd.Flags().StringSliceVar(&options.Components, "components", nil, "Specific components (systemd service) to collect logs from (comma-separated). If empty, collects from all Olares-related components that can be found")
	cmd.Flags().BoolVar(&options.IgnoreKubeErrors, "ignore-kube-errors", options.IgnoreKubeErrors, "Continue collecting logs even if kubectl commands fail")
	cmd.Flags().BoolVar(&options.SkipKubeAPISserver, "skip-kube-apiserver", options.SkipKubeAPISserver, "Skip retrieving logs from kube-apiserver, it's automatically set if apiserver is not reachable. To tolerate other cases, set the ignore-kube-errors")
	cmd.Flags().StringSliceVar(&options.PodNamespaces, "pod-namespaces", nil, "Restrict /var/log/pods collection to these namespaces (comma-separated). If empty, collects pod logs from all namespaces")
	cmd.Flags().BoolVar(&options.SkipSystemd, "skip-systemd", options.SkipSystemd, "Skip collecting systemd service logs")
	cmd.Flags().BoolVar(&options.SkipDmesg, "skip-dmesg", options.SkipDmesg, "Skip collecting dmesg (kernel) logs")
	cmd.Flags().IntVar(&options.DmesgPrevBoots, "dmesg-prev-boots", options.DmesgPrevBoots, "How many boots before the current one to also collect kernel logs for, as dmesg-prev-N.log (0 disables). Needed to investigate a freeze or hard reset, since dmesg only covers the running kernel; requires persistent journaling")
	cmd.Flags().BoolVar(&options.SkipNetwork, "skip-network", options.SkipNetwork, "Skip collecting network configs (ip/iptables/nft)")
	cmd.Flags().BoolVar(&options.SkipClusterInfo, "skip-cluster-info", options.SkipClusterInfo, "Skip collecting cluster info (kubectl describe/pods-list, envoy config) and olares-cli logs")
	cmd.Flags().BoolVar(&options.SkipPodLogs, "skip-pod-logs", options.SkipPodLogs, "Skip collecting pod logs from /var/log/pods")

	cmd.AddCommand(newCmdLogsUpload())

	return cmd
}
