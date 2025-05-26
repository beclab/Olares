/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package utils

import (
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/shurcooL/httpfs/vfsutil"
)

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

func CreateDir(path string) error {
	if IsExist(path) == false {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func IsSymLink(path string) (bool, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.Mode()&os.ModeSymlink != 0, nil
}

func FileMD5(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", err
	}

	m := md5.New()
	if _, err := io.Copy(m, file); err != nil {
		return "", err
	}

	fileMd5 := fmt.Sprintf("%x", m.Sum(nil))
	return fileMd5, nil
}

func DeleteFile(path string) error {
	return os.Remove(path)
}

func LocalMd5Sum(src string) string {
	md5Str, err := FileMD5(src)
	if err != nil {
		logger.Fatalf("get file md5 failed %v", err)
		return ""
	}
	return md5Str
}

func Mkdir(dirName string) error {
	return os.MkdirAll(dirName, os.ModePerm)
}

func CopyEmbed(assets http.FileSystem, embeddedDir, dst string) error {
	return vfsutil.WalkFiles(assets, embeddedDir, func(path string, info os.FileInfo, rs io.ReadSeeker, err error) error {
		if err != nil {
			return err
		}

		if len(path)+1 < len(embeddedDir) {
			return errors.New("embedded dir is empty")
		}
		targetPath := filepath.Join(dst, path[len(embeddedDir)-1:])

		if info.IsDir() {
			return Mkdir(targetPath)
		}

		data, err := vfsutil.ReadFile(assets, path)
		if err != nil {
			return err
		}

		return ioutil.WriteFile(targetPath, data, os.ModePerm)
	})
}
