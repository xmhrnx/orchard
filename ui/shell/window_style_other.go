//go:build !darwin

package shell

import "github.com/xmhrnx/giode"

// windowControls is the Windows look, also used on Linux: minimize,
// maximize and close on the right.
const windowControls = giode.TitleBarWindows
