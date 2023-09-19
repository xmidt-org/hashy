package main

import (
	"fmt"
	"os"

	"github.com/xmidt-org/hashy/pkg/hashy"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		fx.Supply(
			hashy.Config{},
		),
		fx.WithLogger(func(l *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: l}
		}),
		fx.Provide(
			zap.NewDevelopment,
			hashy.NewDatacenterHashers,
			func(dh hashy.DatacenterHashers) hashy.Handler {
				return &hashy.DefaultHandler{
					Hashers: dh,
				}
			},
			func(logger *zap.Logger, lc fx.Lifecycle, sh fx.Shutdowner, h hashy.Handler, cfg hashy.Config) (s *hashy.Server, err error) {
				s, err = hashy.NewServer(cfg)
				if err == nil {
					s.Handler = h
					lc.Append(
						fx.StartStopHook(
							func() {
								go func() {
									logger.Info("server exited", zap.Error(s.ListenAndServe()))
								}()
							},
							s.Shutdown,
						),
					)
				}

				return
			},
		),
		fx.Invoke(
			func(*hashy.Server) {},
		),
	)

	app.Run()
	if err := app.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
