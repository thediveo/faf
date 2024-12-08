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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("64bit dirents", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			// The code under test is synchronous with respect to file handling,
			// so we expect everthing cleaned up correctly already at the very
			// end of each test.
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("reads the test directory correctly", func() {
		dentries := []DirEntry{}
		for dentry := range ReadDir("./_testdata/foo") {
			dentries = append(dentries, DirEntry{
				Ino:  dentry.Ino,
				Name: dentry.Name[:],
				Type: dentry.Type,
			})
		}
		Expect(dentries).To(HaveLen(2))
		Expect(dentries).To(ConsistOf(
			And(HaveField("Name", []byte("bar")),
				HaveField("IsRegular()", BeTrue()),
				HaveField("Ino", Not(BeZero()))),
			And(HaveField("Name", []byte("baz")),
				HaveField("IsDir()", BeTrue()),
				HaveField("Ino", Not(BeZero())))))
	})

	It("returns nothing when directory does not exist", func() {
		count := 0
		for range ReadDir("./_testdata/non-existing") {
			count++
		}
		Expect(count).To(BeZero())
	})

})
