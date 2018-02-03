package json

import (
	"encoding/json"

	"github.com/Diminho/MK_practice/simplelog"
)

type JSONHandler struct {
	*json.Encoder
}

// func New(w io.Writer) simplelog.HandlerFunc {
// 	jsonHandler := JSONHandler{
// 		Encoder: json.NewEncoder(w),
// 	}
// 	return func(*simplelog.Record) error {
// 		logger.mu.Lock()
// 		defer logger.mu.Unlock()
// 		return h.Encoder.Encode(e)
// 	}
// }

func New(logger *simplelog.Logger) simplelog.HandlerFunc {
	jsonHandler := JSONHandler{
		Encoder: json.NewEncoder(logger.Out),
	}

	return func(record *simplelog.Record) error {
		logger.Mu.Lock()
		defer logger.Mu.Unlock()
		return jsonHandler.Encoder.Encode(record)
	}
}
