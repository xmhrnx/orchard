package screens

import "fmt"

// count writes a number with its noun in the right form: "1 release",
// "3 releases".
func count(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", one)
	}
	return fmt.Sprintf("%d %s", n, many)
}
