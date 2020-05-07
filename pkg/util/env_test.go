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
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const envName = "TEST_FALLBACKTOENV"

var _ = Describe("FallbackToEnv", func() {
	It("should fallback on empty string and passthrough value otherwise", func() {
		envValue := "ENV"
		Expect(os.Setenv(envName, envValue)).Should(Succeed())
		Expect(FallbackToEnv("", envName)).Should(Equal(envValue))
		expected := "passthrough"
		Expect(FallbackToEnv(expected, envName)).Should(Equal(expected))
	})
})
