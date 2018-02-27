package json

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/Diminho/MK_practice/simplelog"
)

type Handler struct {
	sync.Mutex
	*json.Encoder
}

func New(w io.Writer) *Handler {
	return &Handler{Encoder: json.NewEncoder(w), Mutex: sync.Mutex{}}
}

func (h *Handler) HandleLog(r *simplelog.Record) error {
	h.Lock()
	defer h.Unlock()
	return h.Encoder.Encode(r)
}
