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

import (
	"fmt"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("byteline", func() {

	When("checking for EOL", func() {

		It("returns EOL for empty line", func() {
			bstr := NewBytestring([]byte{})
			Expect(bstr.EOL()).To(BeTrue())
		})

		It("returns EOL for empty line", func() {
			bstr := NewBytestring([]byte("foo"))
			Expect(bstr.EOL()).To(BeFalse())
			bstr.pos += 3
			Expect(bstr.EOL()).To(BeTrue())
		})

	})

	When("skipping space", func() {

		It("reports EOL", func() {
			bstr := NewBytestring([]byte("   "))
			Expect(bstr.SkipSpace()).To(BeTrue())
		})

		It("advances past spaces", func() {
			bstr := NewBytestring([]byte("   foo"))
			Expect(bstr.SkipSpace()).To(BeFalse())
			Expect(bstr.pos).To(Equal(3))
		})

	})

	When("skipping text", func() {

		It("skips only expected text", func() {
			bstr := NewBytestring([]byte("foobar"))
			Expect(bstr.SkipText("foo")).To(BeTrue())
			Expect(bstr.pos).To(Equal(3))
		})

		It("doesn't skip unexpected things", func() {
			bstr := NewBytestring([]byte("bar"))
			Expect(bstr.SkipText("baz")).To(BeFalse())
			Expect(bstr.pos).To(Equal(0))

			Expect(bstr.SkipText("barz")).To(BeFalse())
			Expect(bstr.pos).To(Equal(0))

			Expect(bstr.SkipText("bar")).To(BeTrue())
			Expect(bstr.pos).To(Equal(3))
		})

	})

	When("getting the next byte", func() {

		It("returns !ok at eol", func() {
			bstr := NewBytestring([]byte(""))
			ch, ok := bstr.Next()
			Expect(ok).To(BeFalse())
			Expect(ch).To(BeZero())
		})

		It("returns the next byte", func() {
			bstr := NewBytestring([]byte("AB"))
			Expect(Ok(bstr.Next())).To(Equal(byte('A')))
			Expect(Ok(bstr.Next())).To(Equal(byte('B')))
			Expect(bstr.EOL()).To(BeTrue())
		})

	})

	When("parsing numbers", func() {

		It("requires at least one digit", func() {
			bstr := &Bytestring{b: []byte("")}
			_, ok := bstr.Uint64()
			Expect(ok).To(BeFalse())
			Expect(bstr.pos).To(Equal(0))

			bstr = &Bytestring{b: []byte("foo")}
			_, ok = bstr.Uint64()
			Expect(ok).To(BeFalse())
			Expect(bstr.pos).To(Equal(0))

			bstr = &Bytestring{b: []byte("!!!")}
			_, ok = bstr.Uint64()
			Expect(ok).To(BeFalse())
			Expect(bstr.pos).To(Equal(0))
		})

		It("returns a correct number", func() {
			bstr := NewBytestring([]byte("4"))
			Expect(Ok(bstr.Uint64())).To(Equal(uint64(4)))
			Expect(bstr.pos).To(Equal(1))

			bstr = NewBytestring([]byte("7foo"))
			Expect(Ok(bstr.Uint64())).To(Equal(uint64(7)))
			Expect(bstr.pos).To(Equal(1))

			bstr = NewBytestring([]byte("1234567890123"))
			Expect(Ok(bstr.Uint64())).To(Equal(uint64(1234567890123)))
			Expect(bstr.pos).To(Equal(13))
		})

		It("rejects numbers outside the uint64 range", func() {
			bstr := NewBytestring([]byte(fmt.Sprintf("%d0", uint64(math.MaxUint64))))
			v, ok := bstr.Uint64()
			Expect(ok).To(BeFalse())
			Expect(v).To(BeZero())
		})
	})

	When("counting fields", func() {

		It("returns nothing from nothing", func() {
			bstr := NewBytestring([]byte(""))
			Expect(bstr.NumFields()).To(BeZero())

			bstr = NewBytestring([]byte(" "))
			Expect(bstr.NumFields()).To(BeZero())
		})

		It("counts correctly", func() {
			bstr := NewBytestring([]byte(" F  BAR BAZ"))
			Expect(bstr.NumFields()).To(Equal(3))
			bstr = NewBytestring([]byte(" F  BAR BAZ RATZ "))
			Expect(bstr.NumFields()).To(Equal(4))
		})

	})

})
