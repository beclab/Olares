package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"k8s.io/utils/strings/slices"

	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"

	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

var (
	containerdConfigDir  = "/etc/containerd"
	containerdConfigFile = filepath.Join(containerdConfigDir, "config.toml")

	instance *MirrorWatcher
	once     sync.Once
	initErr  error
)

func getMirrorWatcher() (*MirrorWatcher, error) {
	once.Do(func() {
		instance, initErr = newMirrorWatcher()
		if initErr != nil {
			klog.Errorf("Failed to initialize containerd mirror watcher: %v", initErr)
		}
	})
	return instance, initErr
}

type MirrorWatcher struct {
	watcher *fsnotify.Watcher
	mu      sync.RWMutex
	mirrors []string
}

// newMirrorWatcher creates a new config watcher
func newMirrorWatcher() (*MirrorWatcher, error) {
	if _, err := os.Stat(containerdConfigDir); err != nil {
		klog.Warningf("Containerd config dir %s does not exist, file watching disabled", containerdConfigDir)
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create config dir watcher: %w", err)
	}

	w := &MirrorWatcher{
		watcher: watcher,
	}
	w.loadConfig()

	err = watcher.Add(containerdConfigDir)
	if err != nil {
		watcher.Close()
		klog.Errorf("Failed to watch containerd config dir: %v", err)
		return nil, err
	}

	go w.watch()

	return w, nil
}

func (w *MirrorWatcher) getMirrors() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.watcher != nil {
		return w.mirrors
	}
	return utils.GetMirrorsEndpoint()
}

// GetMirrors returns the current mirror endpoints, initializing the watcher if needed
// this is the main exported function that other services should use
func GetMirrors() []string {
	watcher, err := getMirrorWatcher()
	if err != nil {
		// fallback to static config if watcher initialization fails
		klog.Warningf("Failed to get config watcher instance, using static config: %v", err)
		return utils.GetMirrorsEndpoint()
	}
	return watcher.getMirrors()
}

// loadConfig loads the containerd config and extracts mirror endpoints
func (w *MirrorWatcher) loadConfig() {
	endpoints := utils.GetMirrorsEndpoint()

	w.mu.Lock()
	oldEndpoints := w.mirrors
	w.mirrors = endpoints
	w.mu.Unlock()

	if !slices.Equal(oldEndpoints, endpoints) {
		klog.Infof("Containerd mirror endpoints changed from %v to %v", oldEndpoints, endpoints)
	}
}

func (w *MirrorWatcher) watch() {
	if w.watcher == nil {
		return
	}

	defer w.watcher.Close()
	defer func() { w.watcher = nil }()

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				klog.Warningf("Containerd config dir watcher closed, exit")
				return
			}

			if event.Name != containerdConfigFile {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				klog.Infof("Containerd config file changed: %s", event.Name)
				w.loadConfig()
			} else if event.Op&fsnotify.Remove != 0 {
				klog.Warningf("Containerd config file removed: %s", event.Name)
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				klog.Warningf("Containerd config dir watcher error, exit")
				return
			}
			klog.Errorf("Error watching containerd config dir: %v", err)
		}
	}
}
