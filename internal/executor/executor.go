// Package executor defines the interface for code execution engines
// that can run code in isolated environments with dependency management.
package executor

import "context"

type Executor interface {
	Execute(ctx context.Context, code string, dependencies []string, envVars map[string]string) (string, error)
}
