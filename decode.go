// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Represents JSON data structure using native Go types: booleans, floats,
// strings, arrays, and maps.

// Inspiration and code have been taken from
// https://golang.org/src/encoding/json/decode.go
// https://github.com/jackpal/bencode-go

package bencode

import (
	"errors"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

// TODO:
// Error handling.
// Double check if interface{} target i correct handled.

type decodeState struct {
	data             []byte
	off              int
	lastReadValueOff int // Used in stream
	err              error
}

// Unmarshal parses the Bencode-encoded data and stores the result
// in the value pointed to by v.
func Unmarshal(data []byte, v interface{}) (err error) {
	var d decodeState
	d.init(data)
	return d.unmarshal(v)
}

// unmarshal starts the recover-failsafe, then moves on with parsing
// the Bencode-encoded data.
func (d *decodeState) unmarshal(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("unmarshal - Bad kind")
	}
	d.value(rv)
	return //d.err
}

func (d *decodeState) init(data []byte) {
	d.data = data
	d.off = 0
}

// Loop through the Bencode-encoded data.
func (d *decodeState) value(v reflect.Value) {
	for d.off < len(d.data) && d.data[d.off] != 'e' {
		pv, err := d.next()
		if err != nil {
			d.err = err
			continue // must be a better way?
		}
		//fmt.Println(pv)
		d.lastReadValueOff = d.off
		d.populateData(pv, v)
	}
}

// Find and return the next object. Will go deeper if necessary.
func (d *decodeState) next() (interface{}, error) {
	switch {
	case d.data[d.off] == 'i':
		return d.parseInt()
	case d.data[d.off] >= '0' && d.data[d.off] <= '9':
		return d.parseString()
	case d.data[d.off] == 'd':
		d.off++
		return d.parseDictionary(1)
	case d.data[d.off] == 'l':
		d.off++
		return d.parseList()
	default:
		d.off++
		return nil, errors.New("Cannot reconize next object")
	}
}

// Dereference pointers until nil or non-pointer is found.
func (d *decodeState) indirect(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		if v.Kind() != reflect.Ptr {
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

// Called upon each parsed value and based on the type it will populate target interface (as reflect.Value)
// If parsed value is a map/slice, it will loop through every value and if necessary recrusivly go deeper.
func (d *decodeState) populateData(pv interface{}, v reflect.Value) {
	//fmt.Println(pv)
	ve := v.Elem()
	switch pvt := pv.(type) {
	case int:
		//fmt.Println(pvt)
	//case float64:
	//	fmt.Println(pvt)
	case string:
		//fmt.Println(pvt)
	case map[string]interface{}: // map == struct with matching tag
		for key, value := range pvt { // loop through parsed values
			var t reflect.Value
		findMatchingTag:
			for _, inputTag := range []string{key, strings.ToLower(key)} {
				for i := 0; i < ve.NumField(); i++ {
					targetTag, _ := parseTag(ve.Type().Field(i))
					if inputTag != targetTag && inputTag != strings.ToLower(targetTag) {
						continue
					}
					t = d.indirect(ve.Field(i))
					break findMatchingTag
				}
			}
			if !t.IsValid() {
				continue
			}
			if t.Kind() != reflect.ValueOf(value).Kind() {
				if reflect.ValueOf(value).Kind() == reflect.Map {
					if t.CanAddr() != true {
						continue
					}
					d.populateData(value, t.Addr())
					continue
				}
				continue
			}
			if t.CanSet() != true {
				continue
			}
			switch t.Interface().(type) {
			case []interface{}, int64, string, float64:
				t.Set(reflect.ValueOf(value))
			default: // Cases where type is a slice but not of type interface
				for _, x := range value.([]interface{}) {
					if reflect.TypeOf(x) == reflect.TypeOf(t.Interface()).Elem() {
						t.Set(reflect.Append(t, reflect.ValueOf(x)))
					}
				}
			}
		}
	case []interface{}: // array == struct with no matching tag
		for _, value := range pvt {
		veLoopList:
			for i := 0; i < ve.NumField(); i++ {
				rv := d.indirect(ve.Field(i))
				if rv.Kind() != reflect.TypeOf(value).Kind() {
					continue
				}
				if rv.CanSet() != true && rv.IsValid() == true {
					continue
				}
				if !(reflect.DeepEqual(rv.Interface(), reflect.Zero(reflect.TypeOf(rv.Interface())).Interface())) {
					continue
				}
				rv.Set(reflect.ValueOf(value))
				break veLoopList
			}
		}
	}
}

func (d *decodeState) parseInt() (interface{}, error) {
	var b []byte
	if d.data[d.off] == 'i' {
		d.off++
	}
	for d.data[d.off] != 'e' {
		b = append(b, d.data[d.off])
		d.off++
	}
	d.off++
	if i, err := strconv.ParseInt(string(b), 10, 64); err == nil {
		return i, err
	}
	if ui, err := strconv.ParseUint(string(b), 10, 64); err == nil {
		return ui, err
	}
	if f, err := strconv.ParseFloat(string(b), 64); err == nil {
		return f, err
	}
	return 0, errors.New("Integer could not be parsed")
}

func (d *decodeState) parseString() (s string, err error) {
	var size []byte
	for d.data[d.off] != ':' {
		if d.data[d.off] < '0' || d.data[d.off] > '9' {
			d.off++
			return "", errors.New("Bad int before :")
		}
		size = append(size, d.data[d.off])
		d.off++
	}
	d.off++
	n, err := strconv.Atoi(string(size))
	if err != nil {
		return
	}
	s = string(d.data[d.off:(d.off + n)])
	d.off += n
	return
}

func (d *decodeState) parseDictionary(level int) (i interface{}, err error) {
	m := make(map[string]interface{})
	var value interface{}
	for d.off < len(d.data) && d.data[d.off] != 'e' {
		i, err = d.next()
		if err != nil {
			return
		}
		if level == 2 {
			return
		}
		value, err = d.parseDictionary(2)
		if err != nil {
			return
		}
		switch i.(type) {
		case string:
			m[i.(string)] = value
		case int64:
			m[string(strconv.FormatInt(i.(int64), 10))] = value
		default:
			return nil, errors.New("parseDictionary - Bad key type")
		}
	}
	d.off++
	return m, err
}

func (d *decodeState) parseList() (ia []interface{}, err error) {
	var i interface{}
	for d.off < len(d.data) && d.data[d.off] != 'e' {
		switch {
		case d.data[d.off] == 'i':
			i, err = d.parseInt()
			if err != nil {
				return
			}
		case d.data[d.off] >= '0' && d.data[d.off] <= '9':
			i, err = d.parseString()
			if err != nil {
				return
			}
		default: // return error or not?
			return ia, errors.New("parseList - Found a bad list-type")
		}
		ia = append(ia, i)
	}
	// clean-up?
	if d.off >= len(d.data) || d.data[d.off] != 'e' {
		return nil, errors.New("parseList - Bad ending on list, more data coming?")
	}
	if d.data[d.off] == 'e' {
		d.off++
	}
	return
}
