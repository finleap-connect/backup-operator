package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kubism-io/backup-operator/pkg/util"

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
		err = src.Backup(dst)
		Expect(err).ToNot(HaveOccurred())
		Expect(dst.Data).Should(Equal(data))
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
		dfp := filepath.Join(dir, "dstfile")
		dst, err := NewFileDestination(dfp)
		Expect(err).ToNot(HaveOccurred())
		Expect(dst).ToNot(BeNil())
		err = src.Backup(dst)
		Expect(err).ToNot(HaveOccurred())
		Expect(dfp).Should(BeAnExistingFile())
		res, err := ioutil.ReadFile(dfp)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).Should(Equal(data))
	})
})
