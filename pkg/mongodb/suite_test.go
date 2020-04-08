package mongodb

import (
	"github.com/onsi/ginkgo/reporters"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSecrets(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("../../reports/util-junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "MongoDB", []Reporter{junitReporter})
}
