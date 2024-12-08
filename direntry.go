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
	"encoding/binary"
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

type DirEntryType uint8

const (
	DirEntryFIFO    DirEntryType = unix.DT_FIFO
	DirEntryChar    DirEntryType = unix.DT_CHR
	DirEntryBlock   DirEntryType = unix.DT_BLK
	DirEntryDir     DirEntryType = unix.DT_DIR
	DirEntryRegular DirEntryType = unix.DT_REG
	DirEntrySymlink DirEntryType = unix.DT_LNK
	DirEntrySocket  DirEntryType = unix.DT_SOCK
)

func (t DirEntryType) String() string {
	if desc, ok := dirEntryTypeDescription[t]; ok {
		return desc
	}
	return fmt.Sprintf("DirEntryType(%d)", t)
}

var dirEntryTypeDescription = map[DirEntryType]string{
	DirEntryFIFO:    "FIFO/pipe",
	DirEntryChar:    "char device",
	DirEntryBlock:   "block device",
	DirEntryDir:     "directory",
	DirEntryRegular: "regular file",
	DirEntrySymlink: "symbolic link",
	DirEntrySocket:  "socket",
}

// DirEntry represents a single directory entry with only name, type and inode
// number.
//
// Please note that in order to avoid heap escaping allocations, DirEntry
// represents the name in the directory entry as a slice of bytes with its
// backing comes from the underlying raw getdents64 syscall. This allows us to
// defer converting the name to a string to user code, where there is a much
// better chance of [compiler optimizations for conversions between strings and
// byte slices].
//
// See also [getdents(2)] for background details.
//
// [getdents(2)]: https://man7.org/linux/man-pages/man2/getdents.2.html
// [compiler optimizations for conversions between strings and byte slices]: https://go101.org/article/string.html
type DirEntry struct {
	Ino  uint64
	Name []byte
	Type DirEntryType
}

// String returns a compact textual description of this DirEntry.
func (d DirEntry) String() string {
	return fmt.Sprintf("DirEntry ino: %d, name: %q, type: %s", d.Ino, string(d.Name), d.Type)
}

// IsDir returns true if this directory entry is about a directory.
func (d DirEntry) IsDir() bool { return d.Type == unix.DT_DIR }

// IsRegular returns true if this directory entry is about a regular file.
func (d DirEntry) IsRegular() bool { return d.Type == unix.DT_REG }

// IsSymlink returns true if this directory entry is about a symbolic link.
func (d DirEntry) IsSymlink() bool { return d.Type == unix.DT_LNK }

// IsPipe returns true if this directory entry is about a pipe, also known as “FIFO”.
func (d DirEntry) IsPipe() bool { return d.Type == unix.DT_FIFO }

// IsSocket returns true if this directory entry is about a socket.
func (d DirEntry) IsSocket() bool { return d.Type == unix.DT_SOCK }

// IsCharDev returns true if this directory entry is about a character device.
func (d DirEntry) IsCharDev() bool { return d.Type == unix.DT_CHR }

// IsBlockDev returns true if this directory entry is about a block device.
func (d DirEntry) IsBlockDev() bool { return d.Type == unix.DT_BLK }

// RawDirEntry64 provides convenient access to a directory entry within a byte
// slice, as returned by the getdents64() syscall.
//
// See also [getdents(2)] for background details.
//
// [getdents(2)]: https://man7.org/linux/man-pages/man2/getdents.2.html
type RawDirEntry64 []byte

// Ino returns the inode number of this directory entry and ok; otherwise, it
// returns false.
func (d RawDirEntry64) Ino() (uint64, bool) {
	return readUint(d, unsafe.Offsetof(unix.Dirent{}.Ino), unsafe.Sizeof(unix.Dirent{}.Ino))
}

// Len returns the length of this directory entry, or false.
func (d RawDirEntry64) Len() (uint64, bool) {
	return readUint(d, unsafe.Offsetof(unix.Dirent{}.Reclen), unsafe.Sizeof(unix.Dirent{}.Reclen))
}

// Type returns the type of this directory entry, or false.
func (d RawDirEntry64) Type() (DirEntryType, bool) {
	offset := uint64(unsafe.Offsetof(unix.Dirent{}.Type))
	if offset >= uint64(len(d)) {
		return 0, false
	}
	return DirEntryType(d[offset]), true
}

// Name returns the name of the directory entry, or false. The name returned
// references the underlying directory entry and thus becomes invalid as soon as
// the underlying directory entry gets overwritten.
func (d RawDirEntry64) Name() ([]byte, bool) {
	dentryLen, ok := d.Len()
	if !ok {
		return nil, false
	}
	nameoffset := uint64(unsafe.Offsetof(unix.Dirent{}.Name))
	namelen := dentryLen - nameoffset
	if nameoffset+namelen > uint64(len(d)) {
		return nil, false
	}
	for pos := nameoffset; pos < nameoffset+namelen; pos++ {
		if d[pos] == 0 {
			return d[nameoffset:pos], true
		}
	}
	return d[nameoffset : nameoffset+namelen], true
}

// readUint returns the unsigned integer of specified size (4 or 8) at the given
// offset of b, together with true. If the buffer b is to small to hold the
// requested unsigned integer, then false is returned instead.
func readUint(b []byte, offset, size uintptr) (val uint64, ok bool) {
	if int(offset+size) > len(b) {
		return 0, false
	}
	switch size {
	case 8:
		return binary.NativeEndian.Uint64(b[offset : offset+size]), true
	case 4:
		return uint64(binary.NativeEndian.Uint32(b[offset : offset+size])), true
	case 2:
		return uint64(binary.NativeEndian.Uint16(b[offset : offset+size])), true
	default:
		return 0, false
	}
}
