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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kubism/backup-operator/pkg/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FileSource", func() {
	It("should read file to buffer", func() {
		data := []byte("temporarycontent")
		dir, err := ioutil.TempDir("", "fsrc")
		Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(dir)
		fp := filepath.Join(dir, "tmpfile")
		err = ioutil.WriteFile(fp, data, 0644)
		Expect(err).ToNot(HaveOccurred())
		src, err := NewFileSource(fp)
		Expect(err).ToNot(HaveOccurred())
		Expect(src).ToNot(BeNil())
		dst, _ := util.NewBufferDestination()
		err = src.Stream(dst)
		Expect(err).ToNot(HaveOccurred())
		Expect(dst.Data["tmpfile"]).Should(Equal(data))
	})
	It("should read file and write to file", func() {
		data := []byte("temporarycontent")
		dir, err := ioutil.TempDir("", "fsrcdst")
		Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(dir)
		sfp := filepath.Join(dir, "srcfile")
		err = ioutil.WriteFile(sfp, data, 0644)
		Expect(err).ToNot(HaveOccurred())
		src, err := NewFileSource(sfp)
		Expect(err).ToNot(HaveOccurred())
		Expect(src).ToNot(BeNil())
		dfp := filepath.Join(dir, "backup", "srcfile")
		err = os.Mkdir(filepath.Dir(dfp), 0777)
		Expect(err).ToNot(HaveOccurred())
		dst, err := NewDirDestination(filepath.Dir(dfp))
		Expect(err).ToNot(HaveOccurred())
		Expect(dst).ToNot(BeNil())
		err = src.Stream(dst)
		Expect(err).ToNot(HaveOccurred())
		Expect(dfp).Should(BeAnExistingFile())
		res, err := ioutil.ReadFile(dfp)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).Should(Equal(data))
	})
})
