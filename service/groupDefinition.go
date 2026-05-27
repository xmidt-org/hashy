// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errInvalidGroupDefinition = errors.New("group definitions must have at least 2 fields")

	errBlankGroup           = errors.New("group names cannot be blank")
	errInvalidGroupLeading  = errors.New("group names may only start with a lowercase character")
	errInvalidGroupTrailing = errors.New("group names may only end with a lowercase or digit character")
	errInvalidGroup         = errors.New("group names may only consist of lowercase, hyphen, or digit characters")
)

// validCharacters is a lookup table for valid characters within group definitions.
var validCharacters = [256]struct {
	groupLeading  bool
	group         bool
	groupTrailing bool
}{}

func init() {
	validCharacters['-'].group = true

	for i := byte('0'); i < '9'; i++ {
		validCharacters[i].group = true
		validCharacters[i].groupTrailing = true
	}

	for i := byte('a'); i < 'z'; i++ {
		validCharacters[i].groupLeading = true
		validCharacters[i].group = true
		validCharacters[i].groupTrailing = true
	}
}

func validateGroupName(name string) error {
	if len(name) == 0 {
		return errBlankGroup
	}

	if !validCharacters[name[0]].groupLeading {
		return errInvalidGroupLeading
	}

	if !validCharacters[name[len(name)-1]].groupTrailing {
		return errInvalidGroupTrailing
	}

	for i := 1; i < len(name)-1; i++ {
		if !validCharacters[name[i]].group {
			return errInvalidGroup
		}
	}

	return nil
}

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
		err = validateGroupName(gdef.Name)
	} else {
		err = errInvalidGroupDefinition
	}

	if err != nil {
		err = fmt.Errorf("invalid group definition [%s]: %s", txt, err)
	}

	return
}
