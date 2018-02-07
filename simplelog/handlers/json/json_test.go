package json_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/Diminho/MK_practice/simplelog"
	logjson "github.com/Diminho/MK_practice/simplelog/handlers/json"
)

func Test(t *testing.T) {
	var buf bytes.Buffer

	logger := simplelog.NewLog(&buf)
	logger.SetHandler(logjson.New(logger))

	simplelog.Now = func() time.Time {
		return time.Unix(0, 0).UTC()
	}

	logger.WithFields(simplelog.Fields{"user": "JohnDoe", "id": "12345"}).Info("INFO")

	expected := `{"level":"info","timestamp":"1970-01-01T00:00:00Z","message":"INFO","fields":{"user":"JohnDoe", "id":"12345"}}
`

	if buf.String() != expected {
		t.Errorf("Expected %s, but got %s.", buf.String(), expected)
	}
}
