package hashy

import (
	"maps"
	"slices"
	"sort"
)

// serviceCollector collects service -> server name mappings.
type serviceCollector map[string]map[string]struct{}

func (sc serviceCollector) clear() {
	clear(sc)
}

func (sc *serviceCollector) add(service, target string) {
	if sc == nil {
		*sc = map[string]map[string]struct{}{
			service: map[string]struct{}{
				target: {},
			},
		}

		return
	}

	targets := (*sc)[service]
	if targets == nil {
		(*sc)[service] = map[string]struct{}{
			target: {},
		}

		return
	}

	targets[target] = struct{}{}
}

// targets returns a slices of all the targets associated with any of the services.
// The targets will be sorted and deduped.
func (sc serviceCollector) targets(services ...string) []string {
	dedupeTargets := make(map[string]struct{})

	for _, service := range services {
		for target := range sc[service] {
			dedupeTargets[target] = struct{}{}
		}
	}

	targets := slices.AppendSeq(
		make([]string, 0, len(dedupeTargets)),
		maps.Keys(dedupeTargets),
	)

	sort.Strings(targets)
	return targets
}
