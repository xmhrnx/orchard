package project

// GoRun runs the module's main package.
func GoRun() []Command { return []Command{{Args: []string{"go", "run", "."}}} }

// GoBuild compiles every package.
func GoBuild() []Command { return []Command{{Args: []string{"go", "build", "./..."}}} }
