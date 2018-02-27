package simplelog

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func init() {
	Now = func() time.Time {
		return time.Unix(0, 0).UTC()
	}
}

func TestNewRecord(t *testing.T) {

	var buf bytes.Buffer
	l := NewLog(&buf)
	r := NewRecord(l)

	eq := reflect.DeepEqual(r.Fields, Fields{})
	if eq {
		fmt.Println("Fields are equal")
	} else {
		fmt.Println("Fields are not equal")
	}
}

func TestWithFields(t *testing.T) {

	var buf bytes.Buffer
	l := NewLog(&buf)
	r := l.WithFields(Fields{"user": "JohnDeer", "id": "123"})

	eq := reflect.DeepEqual(r.Fields, Fields{"user": "JohnDeer", "id": "123"})
	if eq {
		fmt.Println("Fields are equal")
	} else {
		fmt.Println("Fields are not equal")
	}
}

func TestWithField(t *testing.T) {

	var buf bytes.Buffer
	l := NewLog(&buf)
	r := l.WithField("user", "JohnDeer")

	eq := reflect.DeepEqual(r.Fields, Fields{"user": "JohnDeer"})
	if eq {
		fmt.Println("Fields are equal")
	} else {
		fmt.Println("Fields are not equal")
	}
}

func TestPrepare(t *testing.T) {

	var buf bytes.Buffer
	test_str := "test_string"

	l := NewLog(&buf)
	r := NewRecord(l)
	r = r.prepare(InfoLevel, test_str)

	if r.Message != test_str {
		t.Errorf("Expected %s, but got %s.", r.Message, test_str)
	}
}
