package tasks

import "regexp"

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]|\x1b\][^\x07]*\x07`)

// stripANSI removes terminal colour and cursor codes from a line.
func stripANSI(s string) string { return ansiEscape.ReplaceAllString(s, "") }
