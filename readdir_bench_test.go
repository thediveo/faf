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

taskset -c 24-31 go test -bench=ReadDir -run=^$ -cpu=1,2 -benchmem -benchtime=10s -dir-entries=1024

goos: linux
goarch: amd64
pkg: github.com/thediveo/faf
cpu: AMD Ryzen 9 7950X 16-Core Processor
BenchmarkReadDir/os.ReadDir                45370            265333 ns/op          128647 B/op       2055 allocs/op
BenchmarkReadDir/os.ReadDir-2              52126            230698 ns/op          128720 B/op       2055 allocs/op
BenchmarkReadDir/os.File.ReadDir           73788            163475 ns/op          128646 B/op       2055 allocs/op
BenchmarkReadDir/os.File.ReadDir-2         76035            158164 ns/op          128657 B/op       2055 allocs/op
BenchmarkReadDir/os.NewFile                74116            162175 ns/op          128646 B/op       2055 allocs/op
BenchmarkReadDir/os.NewFile-2              76341            156573 ns/op          128658 B/op       2055 allocs/op
BenchmarkReadDir/faf.ReadDir              106318            112801 ns/op              48 B/op          1 allocs/op
BenchmarkReadDir/faf.ReadDir-2            106552            111842 ns/op              48 B/op          1 allocs/op

taskset -c 24-31 go test -bench=ReadDir -run=^$ -cpu=1,2 -benchmem -benchtime=10s -dir-entries=16

goos: linux
goarch: amd64
pkg: github.com/thediveo/faf
cpu: AMD Ryzen 9 7950X 16-Core Processor
BenchmarkReadDir/os.ReadDir              1999430              5991 ns/op            1718 B/op         32 allocs/op
BenchmarkReadDir/os.ReadDir-2            2119070              5650 ns/op            1719 B/op         32 allocs/op
BenchmarkReadDir/os.File.ReadDir         1908222              6264 ns/op            1718 B/op         32 allocs/op
BenchmarkReadDir/os.File.ReadDir-2       1973565              6106 ns/op            1718 B/op         32 allocs/op
BenchmarkReadDir/os.NewFile              2180108              5519 ns/op            1718 B/op         32 allocs/op
BenchmarkReadDir/os.NewFile-2            2218638              5338 ns/op            1718 B/op         32 allocs/op
BenchmarkReadDir/faf.ReadDir             2854731              4207 ns/op              48 B/op          1 allocs/op
BenchmarkReadDir/faf.ReadDir-2           2889433              4180 ns/op              48 B/op          1 allocs/op

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
