package request

type Header struct {
	Authorization string
}

func (h *Header) GetToken() string {
	return h.Authorization[len("Bearer "):]
}
