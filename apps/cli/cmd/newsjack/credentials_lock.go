package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var credentialsLockTimeout = 45 * time.Second

func credentialsLockPath() string {
	return credentialsPath() + ".lock"
}

func lockCredentialsFile() (func(), error) {
	path := credentialsLockPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		file.Close()
		return nil, err
	}
	if err := lockFileWithTimeout(file, credentialsLockTimeout); err != nil {
		file.Close()
		return nil, fmt.Errorf("could not acquire credentials refresh lock %s: %w", path, err)
	}
	return func() {
		_ = unlockFile(file)
		_ = file.Close()
	}, nil
}
