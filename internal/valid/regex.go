package valid

import (
	"regexp"
	"sync"
)

const (
	domainNameRegexStr     = `^(([a-z0-9]{1,63}|[a-z0-9][a-z0-9-]{0,61}[a-z0-9])\.)+[a-z]{1,63}$`
	emailLocalPartRegexStr = `^(?:\w+|[a-z0-9]+-[a-z0-9]+)+(?:\.(\w+|[a-z0-9]+-[a-z0-9]+)+)*(\+[\w-]+)?$`
)

var (
	rxDomainName     = lazyRegex(domainNameRegexStr)
	rxEmailLocalPart = lazyRegex(emailLocalPartRegexStr)
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
