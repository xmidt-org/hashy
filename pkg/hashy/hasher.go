package hashy

import (
	"github.com/billhathaway/consistentHash"
	"go.uber.org/multierr"
)

type Hasher interface {
	Get([]byte) (string, error)
}

func NewHasher(vnodes int, values []string) (Hasher, error) {
	ch := consistentHash.New()
	if err := ch.SetVnodeCount(vnodes); err != nil {
		return nil, err
	}

	for _, v := range values {
		ch.Add(v)
	}

	return ch, nil
}

func NewDatacenterHashers(cfg Config) (m map[Datacenter]Hasher, err error) {
	m = make(map[Datacenter]Hasher, len(cfg.Datacenters))
	for dc, values := range cfg.Datacenters {
		h, hErr := NewHasher(cfg.Vnodes, values)
		err = multierr.Append(err, hErr)
		if hErr == nil {
			m[dc] = h
		}
	}

	return
}
