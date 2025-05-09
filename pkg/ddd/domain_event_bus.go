package ddd

import (
	"context"

	"github.com/mehdihadeli/go-mediatr"
	"go.uber.org/fx"
)

type DomainEventBus struct {
}

func NewDomainEventBus() *DomainEventBus {
	return &DomainEventBus{}
}

func (d *DomainEventBus) Publish(ctx context.Context, agg AggregateRoot) error {

	events := agg.GetDomainEvents()
	agg.ClearDomainEvents()

	for _, evt := range events {
		if err := mediatr.Publish(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}

func DomainEventBusModule(eventHandlerRegistrations ...fx.Option) fx.Option {
	return fx.Options(
		fx.Provide(NewDomainEventBus),
		fx.Options(eventHandlerRegistrations...),
	)
}

func RegisterDomainEventHandlers[T DomainEvent](ctors ...func() mediatr.NotificationHandler[T]) fx.Option {
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
