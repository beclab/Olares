package file

import (
	"fmt"
	"os"
)

func Exists(name string) bool {
	_, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func WriteFile(filePath string, data string, force bool) error {
	if !force && data == "" {
		return fmt.Errorf("write file error, empty data")
	}
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(data)
	return err
}
