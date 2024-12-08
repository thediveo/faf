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

package faf

import (
	"iter"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

// readDirBufferSize is the size of the buffer to read directory entries into.
// As on Linux the maximum path size defaults to 4096 bytes, we need a buffer
// larger than 4k.
const readDirBufferSize = 8192

// keep the linters happy and avoid pointer to slice in the wake of sync.Pool.
type readBuffer struct{ buff []byte }

// let's do it as the stdlib does: keep a buffer pool for directory entries read
// buffers. This gives us another speed kick, as we avoid repeated allocations of
// directory entry buffers including zeroing them out.
var readDirBuffer = &sync.Pool{
	New: func() any { return &readBuffer{make([]byte, readDirBufferSize)} },
}

// ReadDir returns an iterator over entries of the specified directory. In case
// the specified directory does not exist, the iterator does not produce any
// entries.
//
// Please note that ReadDir produces [DirEntry] directory entries where their
// Name fields are stored as []byte, referencing the underlying raw directory
// entry. The lifetime of the Name field is thus limited to the body of the
// iteration loop. This design avoids heap allocations and puts the need, if
// any, into the hands of the loop body.
//
// Due to its Go iterator design, ReadDir needs only a single allocation per
// full iteration, as opposed to O(n) for [os.File.ReadDir]. A second allocation
// only happens for the first run or if there are concurrent directory reads
// going on and a new directory entries read buffer needs to be allocated. This
// directory entries buffer allocation behavior mimics the one exhibited by
// os.File.ReadDir.
//
// ReadDir is roughly 25% faster compared to [os.File.ReadDir] and only needs a
// constant(!) 24 B/op heap allocation, as opposed to O(n) heap allocations for
// the stdlib's ReadDir.
func ReadDir(name string) iter.Seq[DirEntry] {
	return func(yield func(DirEntry) bool) {
		fd, err := unix.Open(name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY, 0)
		if err != nil {
			return
		}
		defer unix.Close(fd)

		rb := readDirBuffer.Get().(*readBuffer)
		defer readDirBuffer.Put(rb)
		buff := rb.buff
		pos := 0
		avail := 0
		var dirEntry DirEntry
		for {
			// (re)fill the buffer when necessary...
			if pos >= avail {
				pos = 0
				var err error
				avail, err = syscall.Getdents(fd, buff)
				if err != nil || avail <= 0 {
					return
				}
			}
			// now drain the buffer, pushing directory entries to the iterator
			// consumer as we progress entry by entry.
			dentry := RawDirEntry64(buff[pos:avail])
			dentryLen, ok := dentry.Len()
			if !ok || dentryLen > uint64(len(dentry)) {
				return // we've fallen off the edge of the disc, erm, directory
			}
			pos += int(dentryLen)
			ino, ok := dentry.Ino()
			if !ok || ino == 0 {
				continue
			}
			// Please note that we here work with the name as a byte slice
			// instead of a string; the rationale being deferring any conversion
			// up into the user code consuming it as this allows the compiler to
			// see how the underlying byte slice is being used and applying
			// optimized conversions. If we would handle name as a string here
			// we would end up with the string escaping to the heap, so needing
			// a heap allocation with a copy.
			//
			// The command "go build -gcflags '-m -l'" then confirms
			// "string(name) does not escape".
			name, ok := dentry.Name()
			if !ok {
				return
			}
			if string(name) == "." || string(name) == ".." {
				continue
			}
			typ, ok := dentry.Type()
			if !ok {
				return
			}
			dirEntry.Ino = ino
			dirEntry.Name = name
			dirEntry.Type = DirEntryType(typ)
			if !yield(dirEntry) {
				return
			}
		}
	}
}
