package bencode

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type unmarshalStreamTest struct {
	in  []string
	out interface{}
	s   interface{}
	err error
}

var streamTest = []unmarshalStreamTest{
	{in: []string{"li42e3:abce", ""}, out: B{X: 42, S: "abc"}, s: new(B)},
	{in: []string{"li42e3:abce", "li43ee"}, out: B{X: 42, S: "abc"}, s: new(B)},
	{in: []string{"l3:abc", "i43ee"}, out: B{X: 43, S: "abc"}, s: new(B)},
}

func TestDecoder(t *testing.T) {
	for _, sv := range streamTest {
		r := strings.NewReader(sv.in[0])
		dec := NewDecoder(r)
		if err := dec.Decode(sv.s); err != nil {
			t.Error(err.Error())
		}
		dec.r = strings.NewReader(sv.in[1])
		if err := dec.Decode(sv.s); err != nil {
			t.Error(err.Error())
		} else {
			if !(reflect.DeepEqual(reflect.ValueOf(sv.s).Elem().Interface(), sv.out)) {
				t.Error(errors.New("Doesnt match up"))
				fmt.Println(reflect.ValueOf(sv.s).Elem().Interface())
				fmt.Println(sv.out)
			}
		}
	}
}
