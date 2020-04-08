package mongodb

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MongoDBSource", func() {
	It("should dump to buffer", func() {
		src, err := NewMongoDBSource("", "")
		Expect(err).ToNot(HaveOccurred())
		Expect(src).ToNot(BeNil())
	})
})
