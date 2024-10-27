package hashycfg

import "go.uber.org/fx"

// Provide establishes the configuration, possibly unmarshaled externally,
// for a hashy process.
func Provide() fx.Option {
	return fx.Provide(
		NewViper,
		Unmarshal,
	)
}
