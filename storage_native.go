//go:build !js

package main

import (
	"os"
	"path/filepath"
)

func savePath() (string, error) {
	p, e := os.UserConfigDir()
	if e != nil {
		return "", e
	}
	p = filepath.Join(p, "english-with-fox")
	if e = os.MkdirAll(p, 0700); e != nil {
		return "", e
	}
	return filepath.Join(p, "profile.json"), nil
}
func loadSave() ([]byte, error) {
	p, e := savePath()
	if e != nil {
		return nil, e
	}
	b, e := os.ReadFile(p)
	if os.IsNotExist(e) {
		return nil, nil
	}
	return b, e
}
func writeSave(b []byte) error {
	p, e := savePath()
	if e != nil {
		return e
	}
	tmp := p + ".tmp"
	if e = os.WriteFile(tmp, b, 0600); e != nil {
		return e
	}
	return os.Rename(tmp, p)
}

func reducedMotionRequested() bool { return false }
