package json

import (
	"encoding/json"

	"github.com/Diminho/MK_practice/simplelog"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	*json.Encoder
	*simplelog.Log
}

// TODO (BH): Implement concurrent logger
func New(l *simplelog.Log) *Handler {
	logrus.Fatal()
	return &Handler{Encoder: json.NewEncoder(l.Out), Log: l}
}

func (h *Handler) HandleLog(r *simplelog.Record) error {
	h.Log.Mu.Lock()
	defer h.Log.Mu.Unlock()
	return h.Encoder.Encode(r)
}
