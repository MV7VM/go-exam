package repository

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func New() fx.Option {
	return fx.Module("migrator",
		fx.Provide(
			NewRepo,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, s *Repository) {
				lc.Append(fx.Hook{
					OnStart: s.OnStart,
				})
			},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("migrator")
		}),
	)
}
