package db

import (
	"regexp"
	"sync"
)

const (
	spaceSlugRegexStr         = `^[a-z0-9][a-z0-9\-]{2,30}[a-z0-9]$`
	spaceSlugReservedRegexStr = `^(?:sys\-.*|system\-?.*|(?:felk.*|.*\-felk(?:\-.*)*))$`
)

var (
	rxSpaceSlug         = lazyRegex(spaceSlugRegexStr)
	rxSpaceSlugReserved = lazyRegex(spaceSlugReservedRegexStr)
)

func lazyRegex(str string) func() *regexp.Regexp {
	var regex *regexp.Regexp
	var once sync.Once
	return func() *regexp.Regexp {
		once.Do(func() {
			regex = regexp.MustCompile(str)
		})
		return regex
	}
}
