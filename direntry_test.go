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
	"bytes"
	"encoding/binary"
	"unsafe"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sys/unix"
)

func makeRawDirEntry64(ino uint64, name string, typ DirEntryType) RawDirEntry64 {
	GinkgoHelper()
	rde := &unix.Dirent{
		Ino:    ino,
		Type:   uint8(typ),
		Reclen: uint16(int(unsafe.Offsetof(unix.Dirent{}.Name)) + len(name) + 1),
	}
	for idx, c := range name {
		rde.Name[idx] = int8(uint8(c))
	}
	buff := bytes.Buffer{}
	Expect(binary.Write(&buff, binary.NativeEndian, rde))
	return RawDirEntry64(buff.Bytes()[:int(unsafe.Offsetof(unix.Dirent{}.Name))+len(name)+1])
}

var _ = Describe("directory entries", func() {

	Context("reading uints in native endianess from buffers", func() {

		It("rejects out of bounds", func() {
			v, ok := readUint(nil, 0, 8)
			Expect(ok).To(BeFalse())
			Expect(v).To(BeZero())

			v, ok = readUint([]byte{1, 2, 3, 4}, 1, 4)
			Expect(ok).To(BeFalse())
			Expect(v).To(BeZero())
		})

		It("rejects unsupported uint sizes", func() {
			v, ok := readUint([]byte{1, 2, 3, 4}, 0, 3)
			Expect(ok).To(BeFalse())
			Expect(v).To(BeZero())
		})

		It("returns correct values", func() {
			b := binary.NativeEndian.AppendUint32(
				binary.NativeEndian.AppendUint32(make([]byte, 4), 42),
				123)
			v, ok := readUint(b, 4, 4)
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(uint64(42)))

			b = binary.NativeEndian.AppendUint64(
				binary.NativeEndian.AppendUint64(make([]byte, 8), 42),
				123)
			v, ok = readUint(b, 8, 8)
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(uint64(42)))

			b = binary.NativeEndian.AppendUint16(
				binary.NativeEndian.AppendUint16(make([]byte, 2), 42),
				123)
			v, ok = readUint(b, 2, 2)
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(uint64(42)))
		})

	})

	DescribeTable("DirEntryType",
		func(t DirEntryType, expected string) {
			Expect(t.String()).To(ContainSubstring(expected))
		},
		Entry(nil, DirEntryFIFO, "pipe"),
		Entry(nil, DirEntryChar, "char device"),
		Entry(nil, DirEntryBlock, "block device"),
		Entry(nil, DirEntryDir, "dir"),
		Entry(nil, DirEntryRegular, "reg"),
		Entry(nil, DirEntrySymlink, "link"),
		Entry(nil, DirEntrySocket, "socket"),
		Entry(nil, DirEntryType(42), "DirEntryType(42)"),
	)

	Context("cooked directory entries", func() {

		It("returns a textual description", func() {
			Expect(DirEntry{
				Ino:  666,
				Name: []byte("foobar"),
				Type: DirEntryFIFO,
			}.String()).To(Equal("DirEntry ino: 666, name: \"foobar\", type: FIFO/pipe"))
		})

		DescribeTable("checking type",
			func(t DirEntryType, f func(_ DirEntry) bool) {
				Expect(f(DirEntry{Type: t})).To(BeTrue())
				Expect(f(DirEntry{})).NotTo(BeTrue())
			},
			Entry(nil, DirEntryDir, DirEntry.IsDir),
			Entry(nil, DirEntryRegular, DirEntry.IsRegular),
			Entry(nil, DirEntrySymlink, DirEntry.IsSymlink),
			Entry(nil, DirEntryFIFO, DirEntry.IsPipe),
			Entry(nil, DirEntrySocket, DirEntry.IsSocket),
			Entry(nil, DirEntryChar, DirEntry.IsCharDev),
			Entry(nil, DirEntryBlock, DirEntry.IsBlockDev),
		)

	})

	Context("raw directory entries", func() {

		When("presented with invalid directory entries", func() {

			It("rejects getting the name", func() {
				rde := makeRawDirEntry64(666, "foobarz", DirEntryDir)
				name, ok := rde[:int(unsafe.Offsetof(unix.Dirent{}.Type))].Name()
				Expect(ok).To(BeFalse())
				Expect(name).To(BeEmpty())

				// entry is truncated, missing the length
				name, ok = rde[:int(unsafe.Offsetof(unix.Dirent{}.Reclen))].Name()
				Expect(ok).To(BeFalse())
				Expect(name).To(BeEmpty())

				// entry is truncated, lacking the name
				name, ok = rde[:int(unsafe.Offsetof(unix.Dirent{}.Name))-1].Name()
				Expect(ok).To(BeFalse())
				Expect(name).To(BeEmpty())

				name, ok = rde[:int(unsafe.Offsetof(unix.Dirent{}.Name))].Name()
				Expect(ok).To(BeFalse())
				Expect(name).To(BeEmpty())
			})

			It("rejects getting the type", func() {
				rde := makeRawDirEntry64(666, "foobarz", DirEntryDir)
				t, ok := rde[:int(unsafe.Offsetof(unix.Dirent{}.Type))].Type()
				Expect(ok).To(BeFalse())
				Expect(t).To(BeZero())
			})

		})

		It("returns correct properties", func() {
			rde := makeRawDirEntry64(666, "foobarz", DirEntrySocket)
			Expect(Ok(rde.Ino())).To(Equal(uint64(666)))
			Expect(Ok(rde.Type())).To(Equal(DirEntrySocket))
			Expect(Ok(rde.Name())).To(Equal([]byte("foobarz")))

			rde[len(rde)-1] = 'x'
			Expect(Ok(rde.Name())).To(Equal([]byte("foobarzx")))

			rde = makeRawDirEntry64(666, "foobar\000z", DirEntrySocket)
			Expect(Ok(rde.Name())).To(Equal([]byte("foobar")))
		})

	})

})
