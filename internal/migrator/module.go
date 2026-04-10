package migrator

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func New() fx.Option {
	return fx.Module("migrator",
		fx.Provide(
			NewMigrator,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, s *Migrator) {
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
