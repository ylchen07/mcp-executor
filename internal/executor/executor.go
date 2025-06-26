package executor

import "context"

type PythonExecutor interface {
	Execute(ctx context.Context, code string, modules []string) (string, error)
}