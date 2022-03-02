package upgrade_test

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	developmentBuildDir string
	releasedBuildDir    string
)

func init() {
	flag.StringVar(&releasedBuildDir, "releasedBuildDir", "../../../aws-released", "location of released version of built broker and brokerpak")
	flag.StringVar(&developmentBuildDir, "developmentBuildDir", "../../", "location of development version of built broker and brokerpak")
}

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Upgrade Suite")
}
