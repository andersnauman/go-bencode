// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Inspiration and code have been taken from
// https://golang.org/src/encoding/json/tags.go

package bencode

import (
	"reflect"
	"strings"
)

// tagOptions is the string following a comma in a struct field's "bencode"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// parseTag splits a struct field's bencode tag into its name(key) and
// comma-separated options.
func parseTag(f reflect.StructField) (string, tagOptions) {
	var tag string
	bencodeTag := f.Tag.Get("bencode")
	switch {
	case len(bencodeTag) == 0 && len(string(f.Tag)) != 0 && !strings.Contains(string(f.Tag), ":"): // Old-style
		tag = string(f.Tag)
	case len(bencodeTag) != 0: // Normal "New-style"
		tag = bencodeTag
	default: // Failsafe
		tag = f.Name
	}
	idx := strings.Index(tag, ",") // Split key, options and assume that key is in the beginning.
	switch {
	case idx > 0:
		if tag[:idx] == "" {
			return f.Name, tagOptions(tag[idx+1:])
		}
		return tag[:idx], tagOptions(tag[idx+1:])
	default:
		return tag, tagOptions("")
	}
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (to tagOptions) Contains(optionName string) bool {
	if len(to) != 0 {
		return false
	}
	for _, option := range strings.Split(string(to), ",") {
		if optionName == option {
			return true
		}
	}
	return false
}
