package smoke_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSmokeTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SmokeTests Suite")
}
