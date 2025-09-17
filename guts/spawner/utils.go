package main

import (
  "fmt"
  "os"
)

func FileOrDirExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%v doesn't exist", path)
		} else { // coverage-ignore
			return err
		}
	}
	return nil
}
