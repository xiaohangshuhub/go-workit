package host

import "context"

type Application interface {
	Run(ctx context.Context) error
}
