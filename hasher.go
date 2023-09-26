package hashy

import (
	"errors"

	"github.com/billhathaway/consistentHash"
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

type Hashers struct {
	groups  []string
	hashers []Hasher
}

func NewHashers(cfg Config) (*Hashers, error) {
	if len(cfg.Groups) == 0 {
		return nil, errors.New("No groups defined")
	}

	h := &Hashers{
		groups:  make([]string, 0, len(cfg.Groups)),
		hashers: make([]Hasher, 0, len(cfg.Groups)),
	}

	for group, values := range cfg.Groups {
		hasher, err := NewHasher(cfg.Vnodes, values)
		if err != nil {
			return nil, err
		}

		h.groups = append(h.groups, group)
		h.hashers = append(h.hashers, hasher)
	}

	return h, nil
}

func (h Hashers) Len() int {
	return len(h.groups)
}

func (h Hashers) HashName(name []byte) (groups, values []string, err error) {
	groups = h.groups
	values = make([]string, len(groups))

	for i := 0; err == nil && i < len(groups); i++ {
		values[i], err = h.hashers[i].Get(name)
	}

	return
}
