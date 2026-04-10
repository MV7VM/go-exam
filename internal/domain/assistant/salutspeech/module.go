package salutspeech

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func New() fx.Option {
	return fx.Module("salutSpeechClient",
		fx.Provide(
			NewSalutSpeechClient,
		),
		//fx.Invoke(
		//	func(lc fx.Lifecycle, s *Client) {
		//		lc.Append(fx.Hook{
		//			OnStart: s.OnStart,
		//		})
		//	},
		//),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("salutSpeechClient")
		}),
	)
}
