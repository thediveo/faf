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
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("parsing uint64s", func() {

	Context("decimal", func() {

		It("returns a correct value", func() {
			Expect(Ok(ParseUint([]byte("42")))).To(Equal(uint64(42)))
			Expect(Ok(ParseUint([]byte(strconv.FormatUint(cutoffDecimalUint64-1, 10))))).
				To(Equal(uint64(cutoffDecimalUint64) - 1))
		})

		It("rejects invalid numbers", func() {
			v, ok := ParseUint([]byte(fmt.Sprintf("%d0", uint64(math.MaxUint64))))
			Expect(ok).NotTo(BeTrue())
			Expect(v).To(BeZero())
		})

		It("rejects trailing junk", func() {
			v, ok := ParseUint([]byte("42DO'H!"))
			Expect(ok).NotTo(BeTrue())
			Expect(v).To(BeZero())
		})

		It("rejects non-number wisdom", func() {
			v, ok := ParseUint([]byte("DO'H!"))
			Expect(ok).NotTo(BeTrue())
			Expect(v).To(BeZero())
		})

	})

	Context("hexadecimal", func() {

		It("returns a correct value", func() {
			Expect(Ok(ParseHexUint([]byte("42")))).To(Equal(uint64(0x42)))
		})

		It("rejects invalid numbers", func() {
			v, ok := ParseHexUint([]byte("1ffffffffffffffff"))
			Expect(ok).NotTo(BeTrue())
			Expect(v).To(BeZero())
		})

		It("rejects trailing junk", func() {
			v, ok := ParseHexUint([]byte("42DO'H!"))
			Expect(ok).NotTo(BeTrue())
			Expect(v).To(BeZero())
		})

		It("rejects non-number wisdom", func() {
			v, ok := ParseHexUint([]byte("GOSH!"))
			Expect(ok).NotTo(BeTrue())
			Expect(v).To(BeZero())
		})

	})

})
