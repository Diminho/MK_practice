package json

import (
	"encoding/json"

	"github.com/Diminho/MK_practice/simplelog"
)

func New(l *simplelog.Log) simplelog.HandlerFunc {
	encoder := json.NewEncoder(l.Out)

	return func(r *simplelog.Record) error {
		l.Mu.Lock()
		defer l.Mu.Unlock()
		return encoder.Encode(r)
	}
}
