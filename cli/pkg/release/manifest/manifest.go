package manifest

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"github.com/Masterminds/semver/v3"
	dockerref "github.com/containerd/containerd/reference/docker"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sort"
	"strings"
)

type OlaresManifest struct {
	APIVersion string         `yaml:"apiVersion"`
	Target     string         `yaml:"target"`
	Output     OutputManifest `yaml:"output"`
}

type OutputManifest struct {
	Binaries   []BinaryOutput    `yaml:"binaries"`
	Containers []ContainerOutput `yaml:"containers"`
}

type BinaryOutput struct {
	ID    string `yaml:"id"`
	Name  string `yaml:"name"`
	AMD64 string `yaml:"amd64"`
	ARM64 string `yaml:"arm64"`
}

type ContainerOutput struct {
	Name string `yaml:"name"`
}

type Manager struct {
	olaresRepoRoot      string
	distPath            string
	cdnURL              string
	ignoreMissingImages bool
	extractedImages     []string
	extractedComponents []BinaryOutput
}

func NewManager(olaresRepoRoot, distPath, cdnURL string, ignoreMissingImages bool) *Manager {
	return &Manager{
		olaresRepoRoot:      olaresRepoRoot,
		distPath:            distPath,
		cdnURL:              cdnURL,
		ignoreMissingImages: ignoreMissingImages,
	}
}

func (m *Manager) Generate() error {
	if err := m.scan(); err != nil {
		return fmt.Errorf("failed to scan Olares repository for images and components: %v", err)
	}

	manifestPath := filepath.Join(m.distPath, "installation.manifest")
	f, err := os.OpenFile(manifestPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := m.writeComponents(f); err != nil {
		return err
	}

	return m.writeImages(f)
}

func (m *Manager) downloadChecksum(name string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s.checksum.txt", m.cdnURL, name))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// as for now
	// the response status code of fetching a missing checksum
	// is 403 rather than 404
	// update: it seems that sometimes 404 is also returned
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
			return "", nil
		}
		return "", fmt.Errorf("failed to download checksum, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read http body for checksum: %v", err)
	}

	return strings.Fields(string(body))[0], nil
}

func (m *Manager) writeComponents(out io.Writer) error {
	for _, component := range m.extractedComponents {
		var fileName, path string
		fields := strings.Split(component.Name, ",")
		fileName = strings.TrimSpace(fields[0])
		if len(fields) > 1 {
			path = strings.TrimSpace(fields[1])
		} else {
			path = "pkg/components"
		}
		md5Name := fmt.Sprintf("%x", md5.Sum([]byte(fileName)))

		urlAMD64 := md5Name
		urlARM64 := "arm64/" + md5Name

		fmt.Printf("downloading md5 checksum for dependency %s, object: %s\n", fileName, md5Name)

		checksumAMD64, err := m.downloadChecksum(urlAMD64)
		if err != nil {
			return err
		}

		checksumARM64, err := m.downloadChecksum(urlARM64)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(out, "%s,%s,%s,%s,%s,%s,%s,%s\n",
			fileName, path, "", urlAMD64, checksumAMD64, urlARM64, checksumARM64, component.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) writeImages(out io.Writer) error {
	for _, image := range m.extractedImages {
		// Generate MD5 hash of the image name
		md5Name := fmt.Sprintf("%x", md5.Sum([]byte(image)))

		// Define URLs for both architectures
		urlAMD64 := md5Name + ".tar.gz"
		urlARM64 := "arm64/" + md5Name + ".tar.gz"

		fmt.Printf("downloading checksum for image %s, object: %s\n", image, md5Name)

		checksumAMD64, err := m.downloadChecksum(md5Name)
		if err != nil {
			return fmt.Errorf("failed to download AMD64 checksum for %s: %v", image, err)
		}
		if checksumAMD64 == "" {
			if m.ignoreMissingImages {
				fmt.Printf("skipping image %s due to missing checksum\n", image)
				continue
			}
			return fmt.Errorf("got empty checksum for image %s", image)
		}

		checksumARM64, err := m.downloadChecksum("arm64/" + md5Name)
		if err != nil {
			return fmt.Errorf("failed to download ARM64 checksum for %s: %v", image, err)
		}

		_, err = fmt.Fprintf(out, "%s.tar.gz,%s,%s,%s,%s,%s,%s,%s\n",
			md5Name, "images", "images.mf", urlAMD64, checksumAMD64, urlARM64, checksumARM64, image)
		if err != nil {
			return fmt.Errorf("failed to write to manifest file: %v", err)
		}
	}

	return nil
}

func (m *Manager) scan() error {
	var images []string
	uniqueComponents := make(map[string]BinaryOutput)

	err := filepath.Walk(m.olaresRepoRoot, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		// shortcut to Olares manifest file
		if strings.EqualFold(info.Name(), "olares.yaml") {
			content, err := os.ReadFile(filePath)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			olaresManifest := &OlaresManifest{}
			err = yaml.Unmarshal(content, olaresManifest)
			if err != nil {
				return fmt.Errorf("failed to unmarshal olares manifest %s: %v", filePath, err)
			}

			for _, c := range olaresManifest.Output.Containers {
				images = append(images, c.Name)
			}

			for _, b := range olaresManifest.Output.Binaries {
				uniqueComponents[b.ID] = b
			}

			return nil
		}

		// extract image from ordinary kubernetes yaml files
		if !info.IsDir() && (strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml")) {
			targetFile, err := os.Open(filePath)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			scanner := bufio.NewScanner(targetFile)
			for scanner.Scan() {
				line := scanner.Text()
				if !strings.HasPrefix(strings.TrimSpace(line), "image:") {
					continue
				}
				image := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "image:"))
				image = strings.Trim(image, "'")
				image = strings.Trim(image, "\"")
				images = append(images, image)
			}
		}

		return nil
	})

	uniqueImages := make(map[string]struct{})

	for _, image := range images {
		image = strings.TrimSpace(image)
		if image == "" {
			continue
		}

		image, err = m.patchImage(image)
		if err != nil {
			return fmt.Errorf("failed to patch image %s: %v", image, err)
		}

		if _, err := dockerref.ParseDockerRef(image); err != nil {
			continue
		}

		uniqueImages[image] = struct{}{}
	}

	var sortedImages []string
	for image := range uniqueImages {
		sortedImages = append(sortedImages, image)
	}
	sort.Strings(sortedImages)
	m.extractedImages = sortedImages

	for _, component := range uniqueComponents {
		component, err = m.patchComponent(component)
		if err != nil {
			return err
		}
		m.extractedComponents = append(m.extractedComponents, component)
	}

	return nil
}
func (m *Manager) getLatestDailyBuildTag() (string, error) {
	cmd := exec.Command("git", "tag", "-l")
	cmd.Dir = m.olaresRepoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get git tags: %v", err)
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 0 || (len(tags) == 1 && tags[0] == "") {
		return "", fmt.Errorf("no git tags found")
	}

	var dailyTags []string
	dailyBuildRegex := regexp.MustCompile(`^\d+\.\d+\.\d-\d{8}$`)

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if dailyBuildRegex.MatchString(tag) {
			dailyTags = append(dailyTags, tag)
		}
	}

	if len(dailyTags) == 0 {
		return "", fmt.Errorf("no daily build tags found")
	}

	sort.Slice(dailyTags, func(i, j int) bool {
		iv, err := semver.NewVersion(dailyTags[i])
		if err != nil {
			return true
		}
		jv, err := semver.NewVersion(dailyTags[j])
		if err != nil {
			return false
		}
		return iv.LessThan(jv)
	})
	return dailyTags[len(dailyTags)-1], nil
}

// Helper function to patch extracted image name
// before validating it
// for now just backup-server is patched
func (m *Manager) patchImage(image string) (string, error) {
	backupServerImageVersionTpl := "{{ $backupVersion }}"
	if !strings.Contains(image, backupServerImageVersionTpl) {
		return image, nil
	}
	backupConfigPath := filepath.Join(m.olaresRepoRoot, "framework/backup-server/.olares/config/cluster/deploy/backup_server.yaml")
	content, err := os.ReadFile(backupConfigPath)
	if err != nil {
		return "", err
	}

	// Extract backup version using regex
	re := regexp.MustCompile(`{{ \$backupVersion := "(.*)" }}`)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) != 2 {
		return "", fmt.Errorf("backup version not found in config")
	}
	backupVersion := matches[1]

	// Replace version
	fmt.Printf("patching backup server version to %s\n", backupVersion)
	image = strings.ReplaceAll(image, backupServerImageVersionTpl, backupVersion)
	return image, nil
}

func (m *Manager) patchComponent(component BinaryOutput) (BinaryOutput, error) {
	if component.ID != "olaresd" {
		return component, nil
	}

	latestDailyBuildTag, err := m.getLatestDailyBuildTag()
	if err != nil {
		return BinaryOutput{}, fmt.Errorf("failed to get latest daily build tag (required to replace olaresd version): %v", err)
	}

	fmt.Printf("patching olaresd version to %s\n", latestDailyBuildTag)

	component.Name = strings.ReplaceAll(component.Name, "#__VERSION__", latestDailyBuildTag)
	component.AMD64 = strings.ReplaceAll(component.AMD64, "#__VERSION__", latestDailyBuildTag)
	component.ARM64 = strings.ReplaceAll(component.ARM64, "#__VERSION__", latestDailyBuildTag)
	return component, nil

}
