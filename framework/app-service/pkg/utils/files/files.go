package files

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

// Untar extracts the contents of a tar archive from the source file  to the destination directory.
func Untar(src, dst string) error {
	fr, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fr.Close()

	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case hdr == nil:
			continue
		}

		dstPath := filepath.Join(dst, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if !IsExist(dstPath) && IsDir(dstPath) {
				if err := CreateDir(dstPath); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			if dir := filepath.Dir(dstPath); !IsExist(dir) {
				if err := CreateDir(dir); err != nil {
					return err
				}
			}

			file, err := os.OpenFile(dstPath, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}

			if _, err = io.Copy(file, tr); err != nil {
				return err
			}

			fmt.Println(dstPath)
			file.Close()
		}
	}
}

// IsExist checks if the specified path exists.
// It returns true if the path exists, false otherwise.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

// CreateDir creates a new directory at the specified path.
func CreateDir(path string) error {
	if !IsExist(path) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// IsDir checks if the specified path is a directory.
// It returns true if the path is a directory, false otherwise.
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// CountDirFiles counts the number of files in the specified directory.
// It returns the total count of files found in the directory.
func CountDirFiles(dirName string) int {
	if !IsDir(dirName) {
		return 0
	}
	var count int
	err := filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		count++
		return nil
	})
	if err != nil {
		klog.Errorf("Failed to count dir files err=%v", err)
		return 0
	}
	return count
}

// FileMD5 calculates the MD5 hash of the file at the specified path.
func FileMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	m := md5.New()
	if _, err := io.Copy(m, file); err != nil {
		return "", err
	}

	fileMd5 := fmt.Sprintf("%x", m.Sum(nil))
	return fileMd5, nil
}

// MkFileFullPathDir is used to file create the full path.
// eg. there is a file "./aa/bb/xxx.txt", and dir ./aa/bb is not exist, and will create the full path dir.
func MkFileFullPathDir(fileName string) error {
	localDir := filepath.Dir(fileName)
	err := Mkdir(localDir)
	if err != nil {
		return fmt.Errorf("create local dir %s failed: %v", localDir, err)
	}
	return nil
}

// Mkdir creates a directory named path, along with any necessary parents, and returns nil,
// or else returns an error.
func Mkdir(dirName string) error {
	return os.MkdirAll(dirName, os.ModePerm)
}

// RemoveAll removes path and any children it contains.
func RemoveAll(path string) error {
	return os.RemoveAll(path)
}
