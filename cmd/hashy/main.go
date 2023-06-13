package main

import (
	"fmt"
	"os"
	"time"

	"github.com/miekg/dns"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.StartTimeout(5*time.Second),
		provideConfig(),
		provideLogger(),
		provideHandler(),
		provideServer(),
		fx.Invoke(
			func(*dns.Server) {},
		),
	)

	app.Run()
	if err := app.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
