//go:build !linux && !darwin && !windows

package trash

import "errors"

func move(string) error { return errors.New("no trash on this system") }
