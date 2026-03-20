// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashycfg

import (
	"errors"

	"github.com/spf13/viper"
)

// NewViper sets up the viper environment for the hashy process.
func NewViper() (v *viper.Viper, err error) {
	v = viper.New()
	v.SetConfigName("hashy")

	// POSIX-style search paths
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.hashy")
	v.AddConfigPath("/etc/hashy")

	err = v.ReadInConfig()

	var notFoundErr viper.ConfigFileNotFoundError
	if errors.As(err, &notFoundErr) {
		// TODO: ignore for now
		err = nil
	}

	return
}

// Unmarshal extracts a hashy configuration from the viper environment.
// This function also applies certain defaults to the configuration.
func Unmarshal(v *viper.Viper) (cfg Config, err error) {
	err = v.UnmarshalExact(&cfg)
	return
}
