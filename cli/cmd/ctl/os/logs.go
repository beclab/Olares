package os

import (
	"archive/tar"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
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
}

var servicesToCollectLogs = []string{"k3s", "containerd", "olaresd", "kubelet", "juicefs", "redis", "minio", "etcd", "NetworkManager"}

func collectLogs(options *LogCollectOptions) error {
	if os.Getuid() != 0 {
		return fmt.Errorf("os: please run as root")
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

	// collect systemd service logs
	if err := collectSystemdLogs(tw, options); err != nil {
		return fmt.Errorf("failed to collect systemd logs: %v", err)
	}

	fmt.Println("collecting dmesg logs ...")
	if err := collectDmesgLogs(tw, options); err != nil {
		return fmt.Errorf("failed to collect dmesg logs: %v", err)
	}

	fmt.Println("collecting logs from kubernetes cluster...")
	if err := collectKubernetesLogs(tw, options); err != nil {
		return fmt.Errorf("failed to collect kubernetes logs: %v", err)
	}

	fmt.Println("collecting olares-cli logs...")
	if err := collectOlaresCLILogs(tw, options); err != nil {
		return fmt.Errorf("failed to collect OlaresCLI logs: %v", err)
	}

	fmt.Println("collecting network configs...")
	if err := collectNetworkConfigs(tw, options); err != nil {
		return fmt.Errorf("failed to collect network configs: %v", err)
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
			continue
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
	cmd := exec.Command("dmesg")
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

func collectKubernetesLogs(tw *tar.Writer, options *LogCollectOptions) error {
	podsLogDir := "/var/log/pods"
	if _, err := os.Stat(podsLogDir); err != nil {
		fmt.Printf("warning: directory %s does not exist, skipping collecting pod logs\n", podsLogDir)
	} else {
		err := filepath.Walk(podsLogDir, func(path string, info os.FileInfo, err error) error {
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
	}

	if _, err := util.GetCommand("kubectl"); err != nil {
		fmt.Printf("warning: kubectl not found, skipping collecting cluster info from kube-apiserver\n")
		return nil
	}

	cmd := exec.Command("kubectl", "get", "pods", "--all-namespaces", "-o", "wide")
	output, err := tryKubectlCommand(cmd, "get pods", options)
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

	cmd = exec.Command("kubectl", "describe", "pods", "--all-namespaces")
	output, err = tryKubectlCommand(cmd, "describe pods", options)
	if err != nil && !options.IgnoreKubeErrors {
		return err
	}
	if err == nil {
		header := &tar.Header{
			Name:    "pods-describe.txt",
			Mode:    0644,
			Size:    int64(len(output)),
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write pods description header: %v", err)
		}
		if _, err := tw.Write(output); err != nil {
			return fmt.Errorf("failed to write pods description data: %v", err)
		}
	}

	cmd = exec.Command("kubectl", "describe", "node")
	output, err = tryKubectlCommand(cmd, "describe node", options)
	if err != nil && !options.IgnoreKubeErrors {
		return err
	}
	if err == nil {
		header := &tar.Header{
			Name:    "node-describe.txt",
			Mode:    0644,
			Size:    int64(len(output)),
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write node description header: %v", err)
		}
		if _, err := tw.Write(output); err != nil {
			return fmt.Errorf("failed to write node description data: %v", err)
		}
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

func getBaseDir() (string, error) {
	// quick path to get basedir from argument instance
	arg := &common.Argument{}
	if err := arg.LoadReleaseInfo(); err != nil {
		return "", fmt.Errorf("failed to load olares release info: %v", err)
	}
	if arg.BaseDir != "" {
		return arg.BaseDir, nil
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
		MaxLines:         3000,
		OutputDir:        "./olares-logs",
		IgnoreKubeErrors: false,
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
	cmd.Flags().BoolVar(&options.IgnoreKubeErrors, "ignore-kube-errors", options.IgnoreKubeErrors, "Continue collecting logs even if kubectl commands fail, e.g., when kube-apiserver is not reachable")

	return cmd
}
