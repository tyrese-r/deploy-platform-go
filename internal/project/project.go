package project

import (
	"fmt"
	"log"
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

// FindDockerfile returns absolute path to Dockerfile
func FindDockerfile(dir string) (string, error) {
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	dfInfo, err := os.Stat(dockerfilePath)

	if err != nil {
		log.Fatalf("no Dockerfile file found at %q: %v", dockerfilePath, err)
	}

	if dfInfo.IsDir() {
		log.Fatalf("%q is a directory")
	}

	return dockerfilePath, nil
}
