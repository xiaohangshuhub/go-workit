package ddd

import (
	"github.com/mehdihadeli/go-mediatr"
	"go.uber.org/fx"
)

func RegisterDomainEventHandlers[T IDomainEvent](ctors ...func() mediatr.NotificationHandler[T]) fx.Option {
	var opts []fx.Option
	for _, ctor := range ctors {
		opts = append(opts,
			fx.Provide(ctor),
			fx.Invoke(func(h mediatr.NotificationHandler[T]) error {
				return mediatr.RegisterNotificationHandler(h)
			}),
		)
	}
	return fx.Options(opts...)
}
