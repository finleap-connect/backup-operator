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

package mongodb

import (
	"github.com/finleap-connect/backup-operator/pkg/testutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MongoDBSource", func() {
	It("should dump to file", func() {
		src, err := NewMongoDBSource(srcURI, "", "backup.tgz")
		Expect(err).ToNot(HaveOccurred())
		Expect(src).ToNot(BeNil())
		dst, err := NewMongoDBDestination(dstURI)
		Expect(err).ToNot(HaveOccurred())
		Expect(dst).ToNot(BeNil())
		_, err = src.Stream(dst)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.FindTestData(dstURI)
		Expect(err).ToNot(HaveOccurred())
	})
})
