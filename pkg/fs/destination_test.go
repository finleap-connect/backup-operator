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

package fs

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FileDestination", func() {
	It("should write buffer to file", func() {
		data := []byte("temporarycontent")
		dir, err := ioutil.TempDir("", "fdst")
		Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(dir)
		fp := filepath.Join(dir, "tmpfile")
		dst, err := NewFileDestination(fp)
		Expect(err).ToNot(HaveOccurred())
		Expect(dst).ToNot(BeNil())
		buf := bytes.NewBuffer(data)
		Expect(buf).ToNot(BeNil())
		err = dst.Store(buf)
		Expect(err).ToNot(HaveOccurred())
		Expect(fp).Should(BeAnExistingFile())
		res, err := ioutil.ReadFile(fp)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).Should(Equal(data))
	})
})
