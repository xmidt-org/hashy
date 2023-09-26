package hashy

type Config struct {
	Address               string
	Network               string
	MaxPacketSize         int
	MaxConcurrentRequests int

	Groups map[string][]string
	Vnodes int
}
