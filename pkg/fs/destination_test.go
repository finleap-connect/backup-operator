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
