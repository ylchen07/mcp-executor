package executor

import "context"

type Executor interface {
	Execute(ctx context.Context, code string, dependencies []string) (string, error)
}