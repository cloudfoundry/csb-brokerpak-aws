// Package brokerpaks manages brokerpak metadata
package brokerpaks

import (
	"fmt"
	"os"
)

var _releasedBrokerpakV140 *bool

func DetectBrokerpakV140(brokerpakDir string) bool {
	if _releasedBrokerpakV140 == nil {
		fileEntries, err := os.ReadDir(brokerpakDir)
		if err != nil {
			fmt.Printf("Cannot open released build directory: %#v\n", err)
		}

		for _, f := range fileEntries {
			if f.Name() == "aws-services-1.4.0.brokerpak" {
				_releasedBrokerpakV140 = boolPtr(true)
				break
			}
		}
		_releasedBrokerpakV140 = boolPtr(false)
	}
	return *_releasedBrokerpakV140
}

func boolPtr(value bool) *bool {
	return &value
}
