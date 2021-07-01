package helpers

import (
	"regexp"
	"strings"
)

var defaultDomain string
var regex = regexp.MustCompile(`^(\S+)\s+shared\s*$`)

func DefaultSharedDomain() string {
	if defaultDomain == "" {
		output, _ := CF("domains")
		for _, line := range strings.Split(output, "\n") {
			matches := regex.FindStringSubmatch(line)
			if len(matches) == 2 {
				defaultDomain = matches[1]
			}
		}
	}
	return defaultDomain
}
