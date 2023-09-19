package hashy

import (
	"strings"

	"github.com/ugorji/go/codec"
)

var msgpackHandle = codec.MsgpackHandle{}

type Datacenter string

type DeviceName string

func (dn DeviceName) GetHashBytes() []byte {
	i := strings.IndexRune(string(dn), ':')
	if i >= 0 {
		return []byte(dn[i+1:])
	}

	return []byte(dn)
}

type DeviceNames []DeviceName

type DeviceHashes map[DeviceName]map[Datacenter][]string

func (dh *DeviceHashes) Add(name DeviceName, dc Datacenter, value string) {
	if *dh == nil {
		*dh = DeviceHashes{
			name: map[Datacenter][]string{
				dc: []string{value},
			},
		}

		return
	}

	if datacenters, exists := (*dh)[name]; exists {
		datacenters[dc] = append(datacenters[dc], value)
	} else {
		(*dh)[name] = map[Datacenter][]string{
			dc: []string{value},
		}
	}
}

func UnmarshalBytes[T any](b []byte) (t T, err error) {
	decoder := codec.NewDecoderBytes(b, &msgpackHandle)
	err = decoder.Decode(&t)
	return
}

func MarshalBytes[T any](t T) (b []byte, err error) {
	encoder := codec.NewEncoderBytes(&b, &msgpackHandle)
	err = encoder.Encode(t)
	return
}
