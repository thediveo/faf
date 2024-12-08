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

//go:build linux

/*

go test -bench=ParseUint -run=^$ -cpu=1,4 -benchmem -benchtime=10s

goos: linux
goarch: amd64
pkg: github.com/thediveo/faf
cpu: AMD Ryzen 9 7950X 16-Core Processor
BenchmarkStrconvParseUint       452616138               26.17 ns/op            0 B/op          0 allocs/op
BenchmarkStrconvParseUint-4     457458100               26.12 ns/op            0 B/op          0 allocs/op
BenchmarkParseUint              880313850               13.80 ns/op            0 B/op          0 allocs/op
BenchmarkParseUint-4            862120438               13.46 ns/op            0 B/op          0 allocs/op

*/

package faf_test

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/thediveo/faf"
)

var err error
var ok bool

func BenchmarkStrconvParseUint(b *testing.B) {
	b.StopTimer()
	bs := []byte(fmt.Sprintf("%v", rand.Uint64N(math.MaxUint64/2)+math.MaxUint64/2))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		_, err = strconv.ParseUint(string(bs), 10, 64)
	}
}

func BenchmarkParseUint(b *testing.B) {
	b.StopTimer()
	bs := []byte(fmt.Sprintf("%v", rand.Uint64N(math.MaxUint64/2)+math.MaxUint64/2))
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		_, ok = faf.ParseUint(bs)
	}
}
