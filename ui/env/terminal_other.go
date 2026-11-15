//go:build !windows && !darwin && !linux

package env

import "errors"

func openTerminal(string) error { return errors.New("opening a terminal is not supported here") }
