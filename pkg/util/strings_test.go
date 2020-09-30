/*
Copyright 2020 Backup Operator Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContainsString", func() {
	It("finds expected string", func() {
		expected := "bar"
		list := []string{"foo", expected}
		Expect(ContainsString(list, expected)).Should(Equal(true))
	})
	It("does not find unexpected string", func() {
		list := []string{"foo", "bar"}
		Expect(ContainsString(list, "baz")).Should(Equal(false))
	})
})

var _ = Describe("RemoveString", func() {
	It("removes element if found", func() {
		element := "bar"
		expected := []string{"foo"}
		input := append(expected, element)
		Expect(RemoveString(input, element)).Should(Equal(expected))
	})
	It("does not remove anything without match", func() {
		element := "bar"
		expected := []string{"foo", "baz"}
		Expect(RemoveString(expected, element)).Should(Equal(expected))
	})
})

var _ = Describe("NilIfEmpty", func() {
	It("returns nil if empty", func() {
		input := ""
		var expected *string
		Expect(NilIfEmpty(input)).Should(Equal(expected))
	})
	It("returns *string if not empty", func() {
		input := "foo"
		Expect(NilIfEmpty(input)).Should(Equal(&input))
	})
})
