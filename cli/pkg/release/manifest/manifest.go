package manifest

import (
	"bufio"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"crypto/md5"
	"fmt"
	dockerref "github.com/containerd/containerd/reference/docker"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Manager struct {
	olaresRepoRoot      string
	cdnURL              string
	ignoreMissingImages bool
}

func NewManager(olaresRepoRoot, cdnURL string, ignoreMissingImages bool) *Manager {
	return &Manager{
		olaresRepoRoot:      olaresRepoRoot,
		cdnURL:              cdnURL,
		ignoreMissingImages: ignoreMissingImages,
	}
}

func (m *Manager) generateImageManifest() error {
	manifestDir := filepath.Join(m.olaresRepoRoot, ".manifest")
	if err := os.RemoveAll(manifestDir); err != nil {
		return err
	}
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return err
	}

	imageManifest := filepath.Join(manifestDir, "images.mf")

	// Copy default base images
	if err := util.CopyFile(
		filepath.Join(m.olaresRepoRoot, "build/manifest/images"),
		imageManifest,
	); err != nil {
		return err
	}

	if err := util.CopyFile(
		filepath.Join(m.olaresRepoRoot, "build/manifest/images.node.mf"),
		filepath.Join(manifestDir, "images.node.mf"),
	); err != nil {
		return err
	}

	// Find images in app modules
	modules := []string{"frameworks", "libs", "apps", "third-party"}
	tmpManifest := filepath.Join(manifestDir, "tmp.image.manifest")

	for _, mod := range modules {
		entries, err := os.ReadDir(filepath.Join(m.olaresRepoRoot, mod))
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			chartPath := filepath.Join(m.olaresRepoRoot, mod, entry.Name(), "config")
			if err := m.findImagesInPath(chartPath, tmpManifest); err != nil {
				return err
			}
		}
	}

	// Process temporary manifest
	return m.processTemporaryManifest(tmpManifest, imageManifest)
}

func (m *Manager) generateDepsManifest() error {
	depsDir := filepath.Join(m.olaresRepoRoot, ".dependencies")
	if err := os.RemoveAll(depsDir); err != nil {
		return err
	}
	if err := os.MkdirAll(depsDir, 0755); err != nil {
		return err
	}

	// Copy components and pkgs
	files := []string{"components", "pkgs"}
	for _, file := range files {
		if err := util.CopyFile(
			filepath.Join(m.olaresRepoRoot, "build/manifest", file),
			filepath.Join(depsDir, file),
		); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) Build() error {
	if err := m.generateImageManifest(); err != nil {
		return fmt.Errorf("image manifest generation failed: %v", err)
	}

	if err := m.generateDepsManifest(); err != nil {
		return fmt.Errorf("deps manifest generation failed: %v", err)
	}

	// Move dependencies
	if err := util.MoveDirectory(
		filepath.Join(m.olaresRepoRoot, ".dependencies"),
		filepath.Join(m.olaresRepoRoot, ".manifest"),
	); err != nil {
		return fmt.Errorf("failed to move dependencies: %v", err)
	}

	manifestFile := filepath.Join(m.olaresRepoRoot, ".manifest/installation.manifest")

	// Process components and pkgs
	deps := []string{"components", "pkgs"}
	for _, dep := range deps {
		if err := m.processDependencies(dep, manifestFile); err != nil {
			return err
		}
	}

	return m.processImages(manifestFile)
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

func (m *Manager) processDependencies(depType, manifestFile string) error {
	file, err := os.Open(filepath.Join(m.olaresRepoRoot, ".manifest", depType))
	if err != nil {
		return err
	}
	defer file.Close()

	manifestOut, err := os.OpenFile(manifestFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer manifestOut.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			return fmt.Errorf("invalid format: %s", line)
		}

		filename := fields[0]
		path := fields[1]
		name := fmt.Sprintf("%x", md5.Sum([]byte(filename)))

		urlAMD64 := name
		urlARM64 := "arm64/" + name

		fmt.Printf("downloading md5 checksum for dependency %s, object: %s\n", filename, name)

		checksumAMD64, err := m.downloadChecksum(urlAMD64)
		if err != nil {
			return err
		}

		checksumARM64, err := m.downloadChecksum(urlARM64)
		if err != nil {
			return err
		}

		fmt.Fprintf(manifestOut, "%s,%s,%s,%s,%s,%s,%s,%s\n",
			filename, path, depType, urlAMD64, checksumAMD64, urlARM64, checksumARM64, fields[4])
	}

	return scanner.Err()
}

func (m *Manager) processImages(manifestFile string) error {
	path := "images"
	manifestOut, err := os.OpenFile(manifestFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer manifestOut.Close()

	imagesList, err := os.Open(filepath.Join(m.olaresRepoRoot, ".manifest/images.mf"))
	if err != nil {
		return err
	}
	defer imagesList.Close()

	scanner := bufio.NewScanner(imagesList)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Generate MD5 hash of the image name
		name := fmt.Sprintf("%x", md5.Sum([]byte(line)))

		// Define URLs for both architectures
		urlAMD64 := name + ".tar.gz"
		urlARM64 := "arm64/" + name + ".tar.gz"

		fmt.Printf("downloading checksum for image %s, object: %s\n", line, urlAMD64)

		checksumAMD64, err := m.downloadChecksum(name)
		if err != nil {
			return fmt.Errorf("failed to download AMD64 checksum for %s: %v", line, err)
		}
		if checksumAMD64 == "" {
			if m.ignoreMissingImages {
				fmt.Printf("skipping image %s due to missing checksum\n", line)
				continue
			}
			return fmt.Errorf("got empty checksum for image %s", line)
		}

		checksumARM64, err := m.downloadChecksum("arm64/" + name)
		if err != nil {
			return fmt.Errorf("failed to download ARM64 checksum for %s: %v", line, err)
		}

		// Write to manifest file
		_, err = fmt.Fprintf(manifestOut, "%s.tar.gz,%s,%s,%s,%s,%s,%s,%s\n",
			name, path, "images.mf", urlAMD64, checksumAMD64, urlARM64, checksumARM64, line)
		if err != nil {
			return fmt.Errorf("failed to write to manifest file: %v", err)
		}
	}

	return scanner.Err()
}

func (m *Manager) findImagesInPath(path string, tmpManifest string) error {
	// Open temporary manifest file for appending
	tmpFile, err := os.OpenFile(tmpManifest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	// Walk through all yaml files in the path
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				relPath, _ := filepath.Rel(m.olaresRepoRoot, filePath)
				fmt.Printf("skipping non existing path: %s\n", relPath)
				return nil
			}
			return err
		}

		// Only process yaml files
		if !info.IsDir() && (strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml")) {
			targetFile, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer targetFile.Close()

			relPath, _ := filepath.Rel(m.olaresRepoRoot, filePath)
			scanner := bufio.NewScanner(targetFile)
			for scanner.Scan() {
				line := scanner.Text()
				if !strings.HasPrefix(strings.TrimSpace(line), "image:") {
					continue
				}
				image := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "image:"))
				image = strings.Trim(image, "'")
				image = strings.Trim(image, "\"")

				// probably some other config yaml
				// instead of a container spec
				if image == "" {
					fmt.Printf("skipping empty image key in file: %s, line: %s\n", relPath, line)
					continue
				}

				image, err = m.patchImage(image)
				if err != nil {
					return fmt.Errorf("failed to patch image %s: %v", image, err)
				}

				if _, err := dockerref.ParseDockerRef(image); err != nil {
					fmt.Printf("skipping invalid image key: %s in file: %s, line: %s\n", image, relPath, line)
					continue
				}

				// Skip specific images
				if strings.Contains(image, "nitro") || strings.Contains(image, "orion") {
					fmt.Printf("skipping image %s in file: %s, line: %s\n", image, relPath, line)
					continue
				}

				fmt.Printf("found image %s in file %s\n", image, relPath)
				// Write image to temporary manifest
				if _, err := fmt.Fprintln(tmpFile, image); err != nil {
					return err
				}
			}
		}
		return nil
	})

	return err
}

func (m *Manager) processTemporaryManifest(tmpManifest, imageManifest string) error {
	// Read temporary manifest
	tmpContent, err := os.ReadFile(tmpManifest)
	if err != nil {
		return err
	}

	// Create a map to store unique images
	uniqueImages := make(map[string]struct{})

	// remove this logic for now
	// to maintain a same order as the one
	// built from shell script
	// Add existing images from image manifest
	//existingContent, err := os.ReadFile(imageManifest)
	//if err != nil {
	//	return err
	//}
	//for _, line := range strings.Split(string(existingContent), "\n") {
	//	if line = strings.TrimSpace(line); line != "" {
	//		uniqueImages[line] = struct{}{}
	//	}
	//}

	// Add new images from temporary manifest
	for _, line := range strings.Split(string(tmpContent), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			uniqueImages[line] = struct{}{}
		}
	}

	// Write unique images back to image manifest
	manifestFile, err := os.OpenFile(imageManifest, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer manifestFile.Close()

	// Convert map to sorted slice for consistent output
	var images []string
	for image := range uniqueImages {
		images = append(images, image)
	}
	sort.Strings(images)

	// Write sorted images to manifest
	for _, image := range images {
		if _, err := fmt.Fprintln(manifestFile, image); err != nil {
			return err
		}
	}

	// Clean up temporary manifest
	return os.Remove(tmpManifest)
}

// Helper function to patch extracted image name
// before validating it
// for now just backup-server is patched
func (m *Manager) patchImage(image string) (string, error) {
	backupServerImageVersionTpl := "{{ $backupVersion }}"
	if !strings.Contains(image, backupServerImageVersionTpl) {
		return image, nil
	}
	backupConfigPath := filepath.Join(m.olaresRepoRoot, "frameworks/backup-server/config/cluster/deploy/backup_server.yaml")
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
