package hashy

type Config struct {
	Address               string
	Network               string
	MaxPacketSize         int
	MaxConcurrentRequests int

	Datacenters map[Datacenter][]string
	Vnodes      int
}
