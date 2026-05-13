package utils

import (
	"context"
	"encoding/json"
	"os"
	"sync"
)

type MountRecord struct {
	Type       DeviceType `json:"type"`
	MountPoint string     `json:"mountPoint"`

	// SMB fields
	SmbPath  string `json:"smbPath,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`

	// NFS fields
	Server    string `json:"server,omitempty"`
	NfsPath   string `json:"nfsPath,omitempty"`
	MountName string `json:"mountName,omitempty"`
}

var mountStoreMu sync.Mutex

func LoadMountRecords(filePath string) ([]MountRecord, error) {
	mountStoreMu.Lock()
	defer mountStoreMu.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var records []MountRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func saveMountRecords(filePath string, records []MountRecord) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0600)
}

func AddMountRecord(filePath string, record MountRecord) error {
	mountStoreMu.Lock()
	defer mountStoreMu.Unlock()

	data, err := os.ReadFile(filePath)
	var records []MountRecord
	if err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &records); err != nil {
			records = nil
		}
	}

	for i, r := range records {
		if r.MountPoint == record.MountPoint {
			records[i] = record
			return saveMountRecords(filePath, records)
		}
	}

	records = append(records, record)
	return saveMountRecords(filePath, records)
}

// GetMountedPoints returns a list of currently mounted path strings.
func GetMountedPoints(ctx context.Context) ([]string, error) {
	paths, err := MountedPath(ctx)
	if err != nil {
		return nil, err
	}

	var points []string
	for _, p := range paths {
		points = append(points, p.Path)
	}
	return points, nil
}

func RemoveMountRecord(filePath string, mountPoint string) error {
	mountStoreMu.Lock()
	defer mountStoreMu.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var records []MountRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	filtered := records[:0]
	for _, r := range records {
		if r.MountPoint != mountPoint {
			filtered = append(filtered, r)
		}
	}

	return saveMountRecords(filePath, filtered)
}
