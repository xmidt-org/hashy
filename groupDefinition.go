// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashy

import (
	"errors"
	"strings"
)

var (
	ErrInvalidGroupDefinition = errors.New("invalid group definition")
)

// GroupDefinition holds the discovered information about a group. Typically, group definitions
// will come from TXT records.
type GroupDefinition struct {
	Name     string
	Services []string
}

// ParseGroupDefinition parses the textual representation of a group's discovered information.
// The text must be a valid US-ASCII string with 2 or more fields delimited by whitespace.
// The first field is the group's name, and all subsequent fields are the services (SRV records)
// that hold the members of the group.
func ParseGroupDefinition(txt string) (gdef GroupDefinition, err error) {
	if fields := strings.Fields(txt); len(fields) > 1 {
		gdef.Name = fields[0]
		gdef.Services = fields[1:]
	} else {
		err = ErrInvalidGroupDefinition
	}

	return
}
