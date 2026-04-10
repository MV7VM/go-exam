package app

import (
	"assistant/internal/config"
	"assistant/internal/domain/assistant/gigachat"
	"assistant/internal/domain/assistant/processor"
	"assistant/internal/domain/assistant/repository"
	"assistant/internal/domain/assistant/salutspeech"
	"assistant/internal/migrator"
	"context"
	"os/signal"
	"syscall"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// New constructs the Fx application wiring together config, logging,
// repositories, use-cases, HTTP server and metrics pipeline.
func New() *fx.App {
	return fx.New(
		fx.Options(
			NewContext(),
			migrator.New(),
			repository.New(),

			gigachat.New(),
			processor.New(),
			salutspeech.New(),
		),
		fx.Provide(
			config.NewConfig,
			//context.Background,
			zap.NewDevelopment,
		),
		fx.WithLogger(
			func(log *zap.Logger) fxevent.Logger {
				return &fxevent.ZapLogger{Logger: log}
			},
		),
	)
}

type ctx struct {
	ctx  context.Context
	stop context.CancelFunc
}

func NewContext() fx.Option {
	return fx.Module("ctx",
		fx.Provide(
			func() (*ctx, context.Context) {
				c, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
				return &ctx{ctx: c, stop: stop}, c
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, s *ctx) {
				lc.Append(fx.Hook{
					//OnStart: s.OnStart,
					OnStop: func(ctx context.Context) error {
						s.stop()
						return nil
					},
				})
			},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("ctx")
		}),
	)
}
