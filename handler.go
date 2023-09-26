package hashy

type Handler interface {
	ServeHash(ResponseWriter, *Request)
}

type HandlerFunc func(ResponseWriter, *Request)

func (hf HandlerFunc) ServeHash(rw ResponseWriter, r *Request) { hf(rw, r) }

type DefaultHandler struct {
	Hashers *Hashers
}

func (dh *DefaultHandler) ServeHash(rw ResponseWriter, r *Request) {
	for {
		name, err := r.Names.Next()
		if err != nil {
			break
		}

		groups, values, err := dh.Hashers.HashName(name)
		if err != nil {
			break
		}

		rw.AddResult(name, groups, values)
	}
}
