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
	"golang.org/x/sys/unix"
)

// ReadFile reads the contents of the named file into the supplied buffer,
// growing the buffer as necessary, returning the contents and true. Only the
// capacity of the passed buffer matters, the buffer's current length gets
// ignored. If the file cannot be read for whatever reason, ReadFile returns
// false. Please note that if the read problem appears in midstream, the buffer
// might have been already reallocated and thus the most recent buffer is
// returned even in case of error.
//
// The buffer underlying the returned contents can be same buffer if the
// capacity was sufficient, otherwise it will be a new backing buffer.
//
// As with [os.ReadFile], reaching the end of the file is not considered to be
// an error but normal operation, and thus not reported.
func ReadFile(name string, buffer []byte) ([]byte, bool) {
	fd, err := unix.Open(name, unix.O_RDONLY, 0)
	if err != nil {
		return buffer, false
	}
	defer unix.Close(fd)

	// If no backing buffer or a buffer with too small capacity was supplied,
	// set up a new initial buffer.
	size := 512
	buffer = buffer[:0]
	if size > cap(buffer) {
		buffer = make([]byte, 0, size)
	}

	// Read in file content chunk by chunk, growing the backing buffer's
	// capacity as needed.
	for {
		n, err := unix.Read(fd, buffer[len(buffer):cap(buffer)])
		buffer = buffer[:len(buffer)+n]
		if err != nil {
			return buffer, false
		}
		if n == 0 {
			return buffer, true
		}
		if len(buffer) == cap(buffer) {
			d := append(buffer[:cap(buffer)], 0)
			buffer = d[:len(buffer)]
		}
	}
}
