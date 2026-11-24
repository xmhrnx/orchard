package project

// installHints tell how to get a missing tool.
var installHints = map[string]string{
	"giode": "giode isn't installed: go install github.com/xmhrnx/giode-cli/cmd/giode@latest (or see https://github.com/xmhrnx/giode-cli)",
	"go":    "Go isn't installed: get it from https://go.dev/dl",
	"npm":   "npm isn't installed: it comes with Node.js from https://nodejs.org",
	"pnpm":  "pnpm isn't installed: npm install -g pnpm, or see https://pnpm.io/installation",
	"yarn":  "Yarn isn't installed: npm install -g yarn, or enable it with corepack enable",
	"bun":   "Bun isn't installed: see https://bun.sh",
}

// InstallHint says how to install a tool a task needs, or "".
func InstallHint(tool string) string { return installHints[tool] }
