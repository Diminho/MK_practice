package simplelog

import (
	"encoding/json"
	"io"
	"sync"
)

type JSONHandler struct {
	sync.Mutex
	*json.Encoder
}

func NewJSONHandler(w io.Writer) *JSONHandler {
	return &JSONHandler{Encoder: json.NewEncoder(w), Mutex: sync.Mutex{}}
}

func (h *JSONHandler) HandleLog(r *Record) error {
	h.Lock()
	defer h.Unlock()
	return h.Encoder.Encode(r)
}
