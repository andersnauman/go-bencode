package bencode

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type unmarshalTest struct {
	in  string
	out interface{}
	s   interface{}
	err error
}

type A struct {
	BB B `bencode:"bb"`
	S  string
	I  []interface{}
}
type B struct {
	X int64 `bencode:"x"`
	F float64
	S string
}
type C struct {
	AA  A `bencode:"aa"`
	AAA *A
}

// http://www.bittorrent.org/beps/bep_0005.html
// DHT Queries - ping example
type PQ struct {
	A PQID   `bencode:"a"`
	T string `bencode:"t"`
	Y string `bencode:"y"`
	Q string `bencode:"q"`
}
type PQID struct {
	ID string `bencode:"id"`
}

type RWP struct {
	R RWPR   `bencode:"r"`
	T string `bencode:"t"`
	Y string `bencode:"y"`
}
type RWPR struct {
	ID     string   `bencode:"id"`
	Token  string   `bencode:"token"`
	Values []string `bencode:"values"`
}

var unmarshalTests = []unmarshalTest{
	{in: "li42e3:abce", out: B{X: 42, S: "abc"}, s: new(B)},
	{in: "li4.2e3:abce", out: B{F: 4.2, S: "abc"}, s: new(B)},
	{in: "l3:abci42ee", out: B{X: 42, S: "abc"}, s: new(B)},
	{in: "l3:abci4.2ee", out: B{F: 4.2, S: "abc"}, s: new(B)},
	//{in: "i7.5e", out: B{F: 7.5}, s: new(B)},
	{in: "d1:xi120ee", out: B{X: 120}, s: new(B)},
	{in: "d1:xi120ee.....2s:.....li4.2ee", out: B{X: 120, F: 4.2}, s: new(B)},
	{in: "d1:xi120e1:S3:abce", out: B{X: 120, S: "abc"}, s: new(B)},
	{in: "d1:ad1:xi120e1:S3:abcee", out: A{}, s: new(A)},
	{in: "d2:bbd1:xi120e1:S3:abcee", out: A{BB: B{X: 120, S: "abc"}}, s: new(A)},
	{in: "d2:bbd1:xi120e1:S3:abce1:s4:edfge", out: A{BB: B{X: 120, S: "abc"}, S: "edfg"}, s: new(A)},
	{in: "d2:bbd1:xi120e1:S3:abce1:S4:edfge", out: A{BB: B{X: 120, S: "abc"}, S: "edfg"}, s: new(A)},
	{in: "d2:bbd1:xi120e1:S3:abce1:S4:edfg1:Ili42e3:abcee", out: A{BB: B{X: 120, S: "abc"}, S: "edfg", I: []interface{}{int64(42), "abc"}}, s: new(A)},
	{in: "d2:aad2:bbd1:xi120e1:S3:abceee", out: C{AA: A{BB: B{X: 120, S: "abc"}}}, s: new(C)},
	{in: "d3:AAAd3:BBBd1:xi120e1:S3:abce1:S3:defee", out: C{AAA: &A{S: "def"}}, s: new(C)},
	{in: "d1:ad2:id20:abcdefghij0123456789e1:q4:ping1:t2:aa1:y1:qe", out: PQ{A: PQID{"abcdefghij0123456789"}, T: "aa", Y: "q", Q: "ping"}, s: new(PQ)},
	{in: "d1:rd2:id20:abcdefghij01234567895:token8:aoeusnth6:valuesl6:axje.u6:idhtnmee1:t2:aa1:y1:re", out: RWP{R: RWPR{ID: "abcdefghij0123456789", Token: "aoeusnth", Values: []string{"axje.u", "idhtnm"}}, T: "aa", Y: "r"}, s: new(RWP)},
}

func TestUnmarshal(t *testing.T) {
	for _, sv := range unmarshalTests {
		if err := Unmarshal([]byte(sv.in), sv.s); err != nil {
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
func TestTags(t *testing.T) {
	a := A{}
	tags := []string{"bb", "S", "I"}
	ve := reflect.ValueOf(a)
	for i := 0; i < ve.NumField(); i++ {
		if tag, _ := parseTag(ve.Type().Field(i)); tag != tags[i] {
			t.Error("Doesnt match up")
		}
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	for i := 0; i <= b.N; i++ {
		Unmarshal([]byte("d1:ad2:id20:abcdefghij0123456789e1:q4:ping1:t2:aa1:y1:qe"), new(PQ))
	}
}
