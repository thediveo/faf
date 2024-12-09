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
	"os"
	"unsafe"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("ReadFile", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			// The code under test is synchronous with respect to file handling,
			// so we expect everthing cleaned up correctly already at the very
			// end of each test.
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("reports when given an unreadable file", func() {
		buff, ok := ReadFile("./_testdata/non-existing", nil)
		Expect(ok).To(BeFalse())
		Expect(buff).To(BeNil())
	})

	It("reads a file correctly, growing a buffer as needed", func() {
		osrContents := Successful(os.ReadFile("LICENSE"))
		contents := Ok(ReadFile("LICENSE", nil))
		Expect(contents).To(Equal(osrContents))
	})

	It("reads a file correctly into a large enough buffer", func() {
		osrContents := Successful(os.ReadFile("LICENSE"))
		buff := make([]byte, len(osrContents)+1) // avoid superfluous re-alloc for exact size
		contents := Ok(ReadFile("LICENSE", buff))
		Expect((*uint8)(unsafe.Pointer(&contents[0]))).To(BeIdenticalTo((*uint8)(unsafe.Pointer(&buff[0]))))
		Expect(contents).To(Equal(osrContents))
	})

})
