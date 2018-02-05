package simplelog_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/Diminho/MK_practice/simplelog"
)

func TestNewRecord(t *testing.T) {

	var buf bytes.Buffer
	l := simplelog.NewLog(&buf)
	r := simplelog.NewRecord(l)

	eq := reflect.DeepEqual(r.Fields, simplelog.Fields{})
	if eq {
		fmt.Println("Fields are equal")
	} else {
		fmt.Println("Fields are not equal")
	}
}

func TestWithFields(t *testing.T) {

	var buf bytes.Buffer
	l := simplelog.NewLog(&buf)
	r := l.WithFields(simplelog.Fields{"user": "JohnDeer", "id": "123"})

	eq := reflect.DeepEqual(r.Fields, simplelog.Fields{"user": "JohnDeer", "id": "123"})
	if eq {
		fmt.Println("Fields are equal")
	} else {
		fmt.Println("Fields are not equal")
	}
}

func TestWithField(t *testing.T) {

	var buf bytes.Buffer
	l := simplelog.NewLog(&buf)
	r := l.WithField("user", "JohnDeer")

	eq := reflect.DeepEqual(r.Fields, simplelog.Fields{"user": "JohnDeer"})
	if eq {
		fmt.Println("Fields are equal")
	} else {
		fmt.Println("Fields are not equal")
	}
}
