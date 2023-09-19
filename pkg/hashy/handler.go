package hashy

type Handler interface {
	ServeHash(ResponseWriter, *Request)
}

type HandlerFunc func(ResponseWriter, *Request)

func (hf HandlerFunc) ServeHash(rw ResponseWriter, r *Request) { hf(rw, r) }

type DefaultHandler struct {
	Datacenters map[Datacenter]Hasher
}

func (dh *DefaultHandler) ServeHash(rw ResponseWriter, r *Request) {
	for _, name := range r.DeviceNames {
		for dc, hasher := range dh.Datacenters {
			// TODO: error handling
			value, _ := hasher.Get(name.GetHashBytes())
			rw.Add(name, dc, value)
		}
	}
}
