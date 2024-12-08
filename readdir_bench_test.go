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

go test -bench=ReadDir -run=^$ -cpu=1,4 -benchmem -benchtime=10s

goos: linux
goarch: amd64
pkg: github.com/thediveo/faf
cpu: AMD Ryzen 9 7950X 16-Core Processor
BenchmarkFileReadDir             1419416              8493 ns/op            1736 B/op         35 allocs/op
BenchmarkFileReadDir-4           1394614              8615 ns/op            1738 B/op         35 allocs/op
BenchmarkReadDir                 1918305              6128 ns/op              24 B/op          1 allocs/op
BenchmarkReadDir-4               1928253              6154 ns/op              24 B/op          1 allocs/op

*/

package faf_test

import (
	"os"
	"testing"

	"github.com/thediveo/faf"
)

func BenchmarkFileReadDir(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dir, err := os.Open("./_testdata/bench")
		if err != nil {
			b.Fatalf("cannot open directory, reason: %s", err)
		}
		_, err = dir.ReadDir(-1)
		dir.Close()
		if err != nil {
			b.Fatalf("cannot read directory, reason: %s", err)
		}
	}
}

func BenchmarkReadDir(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for range faf.ReadDir("./_testdata/bench") {
		}
	}
}
