// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashy

import (
	"errors"
	"fmt"
	"maps"
	"slices"
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

// GroupDefinitionCollector allows a series of GroupDefinitions to be collected
// and deduped. The zero value of this type is ready to use.
type GroupDefinitionCollector map[string]GroupDefinition

// add adds one or more prevalidated definitions.
func (gdc GroupDefinitionCollector) add(more []GroupDefinition) {
	for _, gdef := range more {
		existing := gdc[gdef.Name]
		existing.Name = gdef.Name
		existing.Services = append(existing.Services, gdef.Services...)
		gdc[gdef.Name] = existing
	}
}

// Add adds more definitions. Each definition is validated. If any errors
// occur, this method returns that error and no definitions will have
// been added.
//
// Definitions with the same names are merged into a single definition.
//
// This method will allocate this collector if it has not yet been allocated.
func (gdc *GroupDefinitionCollector) Add(more ...GroupDefinition) (err error) {
	for _, gdef := range more {
		if err = validateGroupName(gdef.Name); err != nil {
			return
		}
	}

	if gdc == nil {
		*gdc = make(GroupDefinitionCollector, len(more))
	}

	gdc.add(more)
	return
}

// AddText parses and validates each text. Only if all values are parsed and
// validated does this method then add the group definitions to this collector.
func (gdc *GroupDefinitionCollector) AddText(more ...string) (err error) {
	gdefs := make([]GroupDefinition, len(more))
	for i, text := range more {
		if gdefs[i], err = ParseGroupDefinition(text); err != nil {
			return
		}
	}

	if gdc == nil {
		*gdc = make(GroupDefinitionCollector, len(gdefs))
	}

	gdc.add(gdefs)
	return
}

// Clear wipes out this collector's definitions.
func (gdc GroupDefinitionCollector) Clear() {
	clear(gdc)
}

// Collect returns a slice of this collector's definitions. The returned slice will
// be sorted by group Name. Each definition's services will be deduped, but not sorted.
func (gdc GroupDefinitionCollector) Collect() (gdefs []GroupDefinition) {
	if len(gdc) == 0 {
		return
	}

	gdefs = slices.AppendSeq(
		make([]GroupDefinition, 0, len(gdc)),
		maps.Values(gdc),
	)

	dedupedServices := make(map[string]struct{}, len(gdefs)*2) // preallocation guestimate
	for i := range len(gdefs) {
		gdef := &gdefs[i]
		if len(gdef.Services) < 2 {
			continue
		}

		clear(dedupedServices)
		for _, service := range gdef.Services {
			dedupedServices[service] = struct{}{}
		}

		if len(gdef.Services) != len(dedupedServices) {
			gdef.Services = slices.AppendSeq(
				gdef.Services[:],
				maps.Keys(dedupedServices),
			)
		}
	}

	slices.SortFunc(
		gdefs,
		func(left, right GroupDefinition) int {
			return strings.Compare(left.Name, right.Name)
		},
	)

	return
}
