package theme

import (
	"testing"

	"github.com/xmhrnx/giode"
)

func TestLoadIconsRegistersBareNames(t *testing.T) {
	if err := LoadIcons(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"repo", "pull", "branch", "scissors"} {
		if _, ok := giode.LookupIcon(name); !ok {
			t.Errorf("glyph %q not registered", name)
		}
	}
}
