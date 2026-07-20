package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// In here it will resolve the path args

func Resolve(pathArg string) (absPath string, err error) {

	absPath, err = filepath.Abs(pathArg)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return "", fmt.Errorf("-path %q is not a directory", absPath)
	}
	return absPath, nil
}

//func FindDockerfile(dir string) (io.Reader, error) {
//
//}
