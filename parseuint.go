// Copyright 2024 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package faf

// ParseUint parses the given byte slice with a decimal number, returning its
// uint64 value and ok, or a zero value and false in case of error. It is an
// error for the given decimal number overflows the uint64 range or if there
// bytes for characters other than "0" to "9" are encounterd.
//
// Compared with [strconv.ParseUint], this version is around 50% faster, albeit
// that is moaning at high level. Both versions do not allocate on the heap in
// case of non-errors, and as our ParseUint does returns a bool instead of an
// error message, it is without heap allocation also in case of error.
//
// Please note that Go's [strconv.ParseUint] has been optimized so that the
// compiler sees that while the input string escapes to the error value,
// ParseUint copies the input string to the error message so that in turn the
// compiler can now use optimized []byte to string conversions.
func ParseUint(b []byte) (uint64, bool) {
	buff := newBytestring(b) // go-es without heap alloc/escape.
	val, ok := buff.Uint64()
	if !ok {
		return 0, ok
	}
	if !buff.EOL() {
		return 0, false
	}
	return val, ok
}
