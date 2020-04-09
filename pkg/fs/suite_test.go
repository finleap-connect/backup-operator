package fs

import (
	"github.com/onsi/ginkgo/reporters"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSecrets(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("../../reports/fs-junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "FS", []Reporter{junitReporter})
}
