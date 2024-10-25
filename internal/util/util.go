package util

import (
	"os"
)

func PathExists(loc string) (bool, error) {
	if _, err := os.Stat(loc); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}
