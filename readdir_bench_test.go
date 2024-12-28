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
	"flag"
	"os"
	"strconv"
	"testing"

	"github.com/thediveo/faf"
	"golang.org/x/sys/unix"
)

var testdataDirEntriesNum uint // number of fake process directory entries to create for benchmarking

func init() {
	flag.UintVar(&testdataDirEntriesNum, "dir-entries", 1024,
		"number of directory entries to use in ReadDir-related benchmarks")
}

func BenchmarkReadDir(b *testing.B) {
	testdatadir := b.TempDir()
	b.Logf("using transient testdata directory %s", testdatadir)
	for num := range testdataDirEntriesNum {
		if err := os.Mkdir(testdatadir+"/"+strconv.FormatUint(uint64(num+1), 10), 0755); err != nil {
			b.Fatalf("cannot create pseudo procfs process directory, reason: %s",
				err.Error())
		}
	}

	f := func(fn func(b *testing.B, tmpdir string)) func(*testing.B) {
		return func(b *testing.B) {
			fn(b, testdatadir)
		}
	}
	b.Run("os.ReadDir", f(bmOsReadDir))
	b.Run("os.File.ReadDir", f(bmFileReadDir))
	b.Run("os.NewFile", f(bmNewFile))
	b.Run("faf.ReadDir", f(bmReadDir))
}

var (
	direntries []os.DirEntry
)

func bmOsReadDir(b *testing.B, testdatadir string) {
	for n := 0; n < b.N; n++ {
		var err error
		direntries, err = os.ReadDir(testdatadir)
		if err != nil {
			b.Fatalf("cannot read directory, reason: %s", err)
		}
	}
}

func bmFileReadDir(b *testing.B, testdatadir string) {
	for n := 0; n < b.N; n++ {
		dir, err := os.Open(testdatadir)
		if err != nil {
			b.Fatalf("cannot open directory, reason: %s", err)
		}
		direntries, err = dir.ReadDir(-1)
		dir.Close()
		if err != nil {
			b.Fatalf("cannot read directory, reason: %s", err)
		}
	}
}

func bmNewFile(b *testing.B, testdatadir string) {
	for n := 0; n < b.N; n++ {
		fd, err := unix.Open(testdatadir, unix.O_RDONLY, 0)
		if err != nil {
			b.Fatalf("cannot open directory, reason: %s", err)
		}
		dir := os.NewFile(uintptr(fd), testdatadir)
		direntries, err = dir.ReadDir(-1)
		dir.Close()
		if err != nil {
			b.Fatalf("cannot read directory, reason: %s", err)
		}
	}
}

func bmReadDir(b *testing.B, testdatadir string) {
	for n := 0; n < b.N; n++ {
		for range faf.ReadDir(testdatadir) {
		}
	}
}
