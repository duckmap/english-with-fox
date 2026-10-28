//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

const storageKey = "english-with-fox.profile.v1"

func loadSave() (b []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("localStorage: %v", r)
		}
	}()
	v := js.Global().Get("localStorage").Call("getItem", storageKey)
	if v.IsNull() || v.IsUndefined() {
		return nil, nil
	}
	return []byte(v.String()), nil
}
func writeSave(b []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("localStorage: %v", r)
		}
	}()
	js.Global().Get("localStorage").Call("setItem", storageKey, string(b))
	return nil
}

func reducedMotionRequested() (value bool) {
	defer func() { _ = recover() }()
	return js.Global().Call("matchMedia", "(prefers-reduced-motion: reduce)").Get("matches").Bool()
}
