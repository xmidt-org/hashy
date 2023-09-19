package hashy

type Handler interface {
	ServeHash(ResponseWriter, *Request)
}

type HandlerFunc func(ResponseWriter, *Request)

func (hf HandlerFunc) ServeHash(rw ResponseWriter, r *Request) { hf(rw, r) }

type DefaultHandler struct {
	Hashers DatacenterHashers
}

func (dh *DefaultHandler) ServeHash(rw ResponseWriter, r *Request) {
	for _, name := range r.DeviceNames {
		for dc, hasher := range dh.Hashers {
			// TODO: error handling
			value, _ := hasher.Get(name.GetHashBytes())
			rw.Add(name, dc, value)
		}
	}
}
