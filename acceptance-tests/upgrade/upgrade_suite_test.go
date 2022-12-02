package upgrade_test

import (
	"flag"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
)

var (
	developmentBuildDir   string
	releasedBuildDir      string
	extraEnvOptions       = make([]brokers.Option, 0)
	releasedBrokerpakV140 = false
)

func init() {
	flag.StringVar(&releasedBuildDir, "releasedBuildDir", "../../../aws-released", "location of released version of built broker and brokerpak")
	flag.StringVar(&developmentBuildDir, "developmentBuildDir", "../../", "location of development version of built broker and brokerpak")
}

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)

	releasedBrokerpakV140 = detectBrokerpakV140()
	if releasedBrokerpakV140 {
		extraEnvOptions = []brokers.Option{brokers.WithLegacyMysqlEnv()}
	}

	RunSpecs(t, "Upgrade Suite")
}

func detectBrokerpakV140() bool {
	dir, err := os.Open(releasedBuildDir)
	if err != nil {
		fmt.Printf("Cannot open released build directory: %#v\n", err)
	}
	files, err := dir.Readdir(0)
	if err != nil {
		fmt.Printf("Cannot list files in released build directory: %#v\n", err)
	}
	for _, f := range files {
		fmt.Printf("%s", f.Name())
		if f.Name() == "aws-services-1.4.0.brokerpak" {
			return true
		}
	}
	return false
}
