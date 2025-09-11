package app

import "context"

type AppContext struct {
	context.Context
}

func NewAppContext(ctx context.Context) *AppContext {
	return &AppContext{
		Context: ctx,
	}
}
