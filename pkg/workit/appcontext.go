package workit

import "context"

type AppContext struct {
	context.Context
}

func newAppContext(ctx context.Context) *AppContext {
	return &AppContext{
		Context: ctx,
	}
}
