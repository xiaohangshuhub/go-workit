package workit

import "context"

type Host interface {
	Run(ctx ...context.Context) error
}
