package gigachat

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func New() fx.Option {
	return fx.Module("gigaChatClient",
		fx.Provide(
			NewGigaChatClient,
		),
		//fx.Invoke(
		//	func(lc fx.Lifecycle, s *Client) {
		//		lc.Append(fx.Hook{
		//			OnStart: s.OnStart,
		//		})
		//	},
		//),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("gigaChatClient")
		}),
	)
}
