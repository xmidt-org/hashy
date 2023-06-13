package main

import (
	"fmt"
	"os"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		provideConfig(),
		provideLogger(),
	)

	app.Run()
	if err := app.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
