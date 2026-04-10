package processor

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func New() fx.Option {
	return fx.Module("bot",
		fx.Provide(
			NewProcessorBot,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, s *ProcessorBot) {
				lc.Append(fx.Hook{
					OnStart: s.OnStart,
					OnStop:  s.OnStop,
				})
			},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("bot")
		}),
	)
}
